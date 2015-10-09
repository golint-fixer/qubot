package qubot

import (
	"logger"
	"sync"

	"config"

	"github.com/nlopes/slack"
)

// Qubot represents the bot service.
type Qubot struct {
	config *config.Config
	db     *DB
	wg     sync.WaitGroup
	client *slack.Client
	rtm    *slack.RTM

	// shutdown is a channel used to coordinate shutting down all the
	// goroutines in this service cleanly.
	shutdown chan struct{}

	// errors is a channel used internally to coordinate management of
	// errors, though this is something I'm not sure yet if I'll be using.
	errors chan error

	// Quit is a channel that we close when Qubot finishes. It is exported
	// so others can tell when we quit.
	Quit chan struct{}

	handlers []Handler
}

// New starts a new instance of Qubot and returns it. It doesn't block but move
// processing to a new goroutine.
func New(config *config.Config) (*Qubot, error) {
	db := DB{}
	if err := db.Open(config.Database.Location, 0640); err != nil {
		return nil, err
	}
	defer db.Close()

	q := Qubot{
		config:   config,
		db:       &db,
		shutdown: make(chan struct{}),
		errors:   make(chan error),
		Quit:     make(chan struct{}),
	}

	q.Handle(
		PingHandler,
		TauntHandler,
		RedmineHandler,
	)

	q.start()

	return &q, nil
}

// Handle registers a new handlers with Qubot
func (q *Qubot) Handle(handlers ...interface{}) {
	for _, h := range handlers {
		nh, err := NewHandler(h)
		if err != nil {
			logger.Crit("msg", "Handle coult not be registered", "error", err)
			panic(err)
		}
		q.handlers = append(q.handlers, nh)
	}
}

// Handlers returns the robot's handlers
func (q *Qubot) Handlers() []Handler {
	return q.handlers
}

func (q *Qubot) start() {
	q.wg.Add(2)

	// Listening on the errors channel.
	go func() {
		for {
			select {
			case err := <-q.errors:
				// TODO - You have an error, do something about it.
				// Error with no recovery could: q.wg.Done(); q.Close()
				logger.Error("msg", "Error handler received an error", "error", err)
			case <-q.shutdown:
				q.wg.Done()
				return
			}
		}
	}()

	// Chatroom connection goroutine.
	go func() {
		client := slack.New(q.config.Slack.Key)
		q.client = client

		// Test authentication and exit if failed.
		if resp, err := q.client.AuthTest(); err == nil {
			logger.Info("msg", "Authentication request test succeeded", "url", resp.URL)
		} else {
			logger.Error("msg", "Authentication request test failed")
			q.errors <- err
			q.wg.Done()
			q.Close()
			return
		}

		defer q.wg.Done()

		q.rtm = client.NewRTM()
		go q.rtm.ManageConnection()

		for {
			select {
			case event := <-q.rtm.IncomingEvents:
				// Process the incoming event in a new goroutine
				// so we can keep listening.
				q.wg.Add(1)
				go q.processIncomingEvent(&event)
			case <-q.shutdown:
				logger.Info("msg", "Disconnecting from Slack RTM")
				q.rtm.Disconnect()
				return
			}
		}
	}()
}

// receive takes the Slack message to all the handlers registered.
func (q *Qubot) receive(msg *slack.Msg) error {
	resp := &Response{
		Qubot: q,
		Msg:   msg,
	}
	for _, h := range q.handlers {
		err := h.Handle(resp)
		if err != nil {
			return err
		}
	}
	return nil
}

// processIncomingEvent processes incoming events from the real time API. We are
// not listening to all the types of events, e.g. slack.HelloEvent,
// slack.ConnectedEvent, slack.PresenceChangeEvent, slack.LatencyReport
func (q *Qubot) processIncomingEvent(event *slack.RTMEvent) {
	defer q.wg.Done()
	// Remember that the variable declared in the type switch will have the
	// corresponding type in each clause.
	switch e := event.Data.(type) {
	case *slack.ConnectedEvent:
		logger.Info("msg", "Connected!")
	case *slack.MessageEvent:
		_ = q.receive(&e.Msg)
	case *slack.RTMError:
	case *slack.InvalidAuthEvent:
	case *slack.DisconnectedEvent:
		// TODO - Define my own error type so I can send to errors ch?
		// See slack.RTMEvent and slack.RTMError
		logger.Info("msg", "Some kind of error event received", "type", event.Type)
	default:
		logger.Debug("msg", "Unknown event received", "type", event.Type)
	}
}

// Close announces other goroutines that we are leaving and waits for them.
func (q *Qubot) Close() {
	logger.Info("msg", "Shutting down Qubot")
	close(q.shutdown)
	q.wg.Wait()
	close(q.Quit)
	logger.Info("msg", "Qubot was shut down successfully. ¡Adiós!")
}
