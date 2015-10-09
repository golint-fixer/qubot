package qubot

import (
	"fmt"
	"logger"
	"reflect"
	"sync"

	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

// Qubot at your service!
type Qubot struct {
	config   *Config
	handlers []Handler
	m        Messenger
	client   *slack.Client
	rtm      *slack.RTM

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	ready  chan struct{}
	done   chan struct{}
}

// Init creates the object and returns a pointer to it.
func Init(config *Config) *Qubot {
	q := Qubot{
		config: config,
		ready:  make(chan struct{}),
		done:   make(chan struct{}),
	}
	root := context.Background()
	q.ctx, q.cancel = context.WithCancel(root)
	return &q
}

// Handle registers a new Handler with Qubot.
func (q *Qubot) Handle(handlers ...Handler) {
	for _, h := range handlers {
		q.handlers = append(q.handlers, h)
		logger.Info("qubot", fmt.Sprintf("Registering handler %s", reflect.TypeOf(h)))
	}
}

// Start the service without blocking.
func (q *Qubot) Start() error {
	// This is primarily for testing purposes, so we can verify from outside
	// that this function does not block.
	defer func() {
		close(q.ready)
	}()

	// Initialize all the listeners that has been registered.
	for _, h := range q.handlers {
		q.wg.Add(1)
		go func(h Handler) {
			defer q.wg.Done()
			err := h.Start(q.ctx)
			if err != nil {
				logger.Warn("qubot", fmt.Sprintf("Handler %s terminated", reflect.TypeOf(h)))
			}
		}(h)
	}

	// Connect to Slack.
	err := q.connect()
	if err != nil {
		return err
	}

	// Start messenger.
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.m = InitMessenger(q.ctx, q.rtm)
		<-q.ctx.Done()
		q.m.Close()
	}()

	// Start event listener.
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		q.listenEvents()
	}()

	return nil
}

func (q *Qubot) connect() error {
	q.client = slack.New(q.config.Slack.Key)
	q.rtm = q.client.NewRTM()

	// Don't get too far until we validate our credentials.
	resp, err := q.client.AuthTest()
	if err != nil {
		logger.Error("qubot", "Authentication request test failed")
		return err
	}
	logger.Info("qubot", "Authentication request test succeeded", "url", resp.URL)

	go q.rtm.ManageConnection()

	return err
}

// listenEvents starts a new goroutine for each event received.
func (q *Qubot) listenEvents() {
	for {
		select {
		case event := <-q.rtm.IncomingEvents:
			q.wg.Add(1)
			go func() {
				defer q.wg.Done()
				q.handleEvent(&event)
			}()
		case <-q.ctx.Done():
			logger.Info("qubot", "Disconnecting from Slack RTM")
			q.rtm.Disconnect()
			return
		}
	}
}

// handleEvents processes incoming events from Slack and route them where it's
// needed, e.g. broadcast messages to the handlers or deal with errors.
// A full list of events can be found in the source code: https://goo.gl/ESCO4K.
func (q *Qubot) handleEvent(event *slack.RTMEvent) {
	switch e := event.Data.(type) {
	case *slack.ConnectingEvent:
		logger.Debug("qubot", "Connection attempt", "count", e.Attempt)
	case *slack.ConnectedEvent:
		logger.Info("qubot", "Connected to Slack!")
	case *slack.HelloEvent:
		logger.Info("qubot", "Slack sent greetings!")
	case *slack.LatencyReport:
		logger.Debug("qubot", "Latency report", "duration", e.Value)
	case *slack.MessageEvent:
		logger.Debug("qubot", "Message received")
		// Broadcast messages to handlers
		for _, h := range q.handlers {
			r := NewResponse(&e.Msg)
			m, ok := h.(HandlerMatcher)
			if ok && !m.Match(r) {
				continue
			}
			h.Handle(r)
		}
	case *slack.RTMError:
	case *slack.InvalidAuthEvent:
	case *slack.AckErrorEvent:
	case *slack.ConnectionErrorEvent:
	case *slack.DisconnectedEvent:
	case *slack.MessageTooLongEvent:
	case *slack.OutgoingErrorEvent:
	default:
		logger.Debug("qubot", "Unknown event received", "type", event.Type)
	}
}

// Report makes Qubot log some vitals about the service. Nothing serious here
// yet.
func (q *Qubot) Report() {
	info := q.rtm.GetInfo()
	logger.Info("qubot", "Status report", "team", fmt.Sprintf("[%s] %s (%s)", info.Team.ID, info.Team.Name, info.Team.Domain))
}

// Done returns a channel that will be closed when the service is totally done.
// It's convenient if you want to wait until the service shuts down.
func (q *Qubot) Done() chan struct{} {
	return q.done
}

// Close shuts down the service cleanly.
func (q *Qubot) Close() {
	q.cancel()    // Emit cancellation signal.
	q.wg.Wait()   // Wait until all the goroutines are done.
	close(q.done) // Signal external receivers.
}
