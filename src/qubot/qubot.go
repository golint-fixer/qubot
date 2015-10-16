package qubot

import (
	"fmt"
	"logger"
	"reflect"
	"sync"
	"time"

	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

var ignoreUserList = []string{"USLACKBOT"}
var eventTimeout = time.Second * 10

// Qubot at your service!
type Qubot struct {
	config   *Config
	handlers []Handler
	m        Messenger
	db       *DB
	client   slackClient
	rtm      slackRTMClient

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	ready  chan struct{}
	done   chan struct{}

	me    *slack.User
	users map[string]*slack.User
}

// Init creates the Qubot object and returns a pointer to it.
func Init(config *Config) *Qubot {
	q := Qubot{
		config: config,
		ready:  make(chan struct{}),
		done:   make(chan struct{}),
		users:  make(map[string]*slack.User),
	}
	q.client = newSlackClient(config.Slack.Key)
	q.rtm = q.client.NewRTM()

	q.db = &DB{}
	err := q.db.Open(q.config.Database.Location, 0600)
	if err != nil {
		panic(err)
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

	// Connect to Slack.
	err := q.connect()
	if err != nil {
		return err
	}

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
	// Don't get too far until we validate our credentials.
	resp, err := q.client.AuthTest()
	if err != nil {
		logger.Error("qubot", "Authentication request test failed")
		return err
	}
	logger.Info("qubot", "Authentication request test succeeded", "url", resp.URL)

	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		// ManageConnection will exit sometime after q.rtm.Disconnect()
		q.rtm.ManageConnection()
	}()

	return err
}

// listenEvents starts a new goroutine for each event received.
func (q *Qubot) listenEvents() {
	var wg sync.WaitGroup
	c := make(chan error, 1)
	for {
		select {
		case event := <-q.rtm.Events():
			wg.Add(2)
			go func() {
				defer wg.Done()
				go func() {
					defer wg.Done()
					defer func() {
						if p := recover(); p != nil {
							logger.Warn("qubot", "listenEvents", "Panic! Regained control.", p)
						}
					}()
					c <- q.handleEvent(&event)
				}()
				// Await until one of the channels send.
				select {
				case <-c:
				case <-time.After(eventTimeout):
					return
				}
			}()
		case err := <-c:
			if err != nil {
				logger.Warn("qubot", "listenEvents", "error", err)
			}
		case <-q.ctx.Done():
			wg.Wait()
			logger.Info("qubot", "Disconnecting from Slack RTM")
			q.rtm.Disconnect()
			return
		}
	}
}

// handleEvents takes each type of event to its corresponding callback.
// A full list of events can be found in the source code: https://goo.gl/ESCO4K.
func (q *Qubot) handleEvent(event *slack.RTMEvent) error {
	switch e := event.Data.(type) {
	case *slack.ConnectingEvent:
		logger.Debug("qubot", "Connection attempt", "count", e.Attempt)
	case *slack.ConnectedEvent:
		logger.Info("qubot", "Connected to Slack!")
		return q.onConnectedEvent(e)
	case *slack.HelloEvent:
		logger.Info("qubot", "Slack sent greetings!")
	case *slack.LatencyReport:
		logger.Debug("qubot", "Latency report", "duration", e.Value)
	case *slack.MessageEvent:
		logger.Debug("qubot", "Message received")
		return q.onMessageEvent(e)
	case *slack.InvalidAuthEvent:
		panic("Unrecoverable error: InvalidAuthEvent")
	case *slack.RTMError:
	case *slack.AckErrorEvent:
	case *slack.ConnectionErrorEvent:
	case *slack.DisconnectedEvent:
	case *slack.MessageTooLongEvent:
	case *slack.OutgoingErrorEvent:
	default:
		logger.Debug("qubot", "Unknown event received", "type", event.Type)
	}
	return nil
}

// onConnectedEvent retrieves information about the team and persist it.
func (q *Qubot) onConnectedEvent(_ *slack.ConnectedEvent) error {
	info := q.rtm.GetInfo()
	for _, user := range info.Users {
		if user.Name == q.config.Slack.Nickname {
			q.me = &user
			continue
		}
		for _, iu := range ignoreUserList {
			if user.Name == iu || user.ID == iu {
				continue
			}
		}
		if user.IsBot { // and user.Deleted?
			continue
		}
		q.users[user.ID] = &user

		// Persist user to the database
		err := q.db.Update(func(tx *Tx) error {
			u, err := tx.User(user.ID)
			if err != nil {
				return err
			}
			if u != nil {
				return nil
			}
			return tx.SaveUser(&User{
				ID:       user.ID,
				Name:     user.Name,
				Email:    user.Profile.Email,
				Creation: time.Now(),
			})
		})
		if err != nil {
			logger.Error("qubot", "...", "error", err)
		}
	}

	logger.Info("qubot", fmt.Sprintf("%d users have been identified (not incluing me or slackbot)", len(q.users)))
	return nil
}

// onMessageEvent broadcasts incoming messages to handlers. Each handler runs
// in a separate goroutine.
func (q *Qubot) onMessageEvent(e *slack.MessageEvent) error {
	for _, h := range q.handlers {
		msg := NewMessage(&e.Msg)
		r := NewResponse(q.m)
		m, ok := h.(HandlerMatcher)
		if ok && !m.Match(r, msg) {
			continue
		}
		h.Handle(r, msg)
	}
	return nil
}

// Report makes Qubot log some vitals about the service.
// Nothing serious here yet.
func (q *Qubot) Report() {
	info := q.rtm.GetInfo()
	logger.Info("qubot", "Status report", "team", fmt.Sprintf("[%s] %s (%s)", info.Team.ID, info.Team.Name, info.Team.Domain))
}

// Done returns a channel that will be closed when the service is totally done.
// It's a convenience for external observers that wants to wait until the
// service finishes.
func (q *Qubot) Done() chan struct{} {
	return q.done
}

// Close shuts down the service cleanly.
func (q *Qubot) Close() {
	q.cancel()    // Emit cancellation signal.
	q.wg.Wait()   // Wait until all the goroutines are done.
	close(q.done) // Signal external receivers.
}
