package qubot

import (
	"fmt"
	"logger"
	"sync"

	"config"

	"github.com/nlopes/slack"
)

// Qubot represents the bot service.
type Qubot struct {
	config   *config.Config
	db       *DB
	shutdown chan struct{}
	errors   chan error
	Quit     chan struct{}
	wg       sync.WaitGroup
	client   *slack.Client
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

	q.start()

	return &q, nil
}

func (q *Qubot) start() {
	q.wg.Add(3)

	// Listening on the errors channel.
	go func() {
		select {
		case <-q.shutdown:
			q.wg.Done()
			return
		case <-q.errors:
			q.wg.Done()
			q.Close()
			return
		}
	}()

	// Chatroom message processing goroutine.
	go func() {
		for {
			select {
			case <-q.shutdown:
				// Chatroom shutting down.
				q.wg.Done()
				return
			}
		}
	}()

	// Chatroom connection goroutine.
	go func() {
		client := slack.New(q.config.Slack.Key)
		q.client = client

		// Test authentication
		if resp, err := q.client.AuthTest(); err == nil {
			logger.Info("msg", "Authentication request test succeeded", "url", resp.URL)
		} else {
			logger.Error("msg", "Authentication request test failed")
			q.errors <- err
			q.wg.Done()
			return
		}

		rtm := client.NewRTM()
		go rtm.ManageConnection()

		for {
			select {
			case event := <-rtm.IncomingEvents:
				q.processIncomingEvent(&event)
			case <-q.shutdown:
				logger.Info("msg", "Disconnecting from Slack RTM")
				rtm.Disconnect()
				q.wg.Done()
				return
			}
		}
	}()
}

// processIncomingEvent processes incoming events from the real time API. We are
// not listening to all the types of events, e.g. slack.HelloEvent,
// slack.ConnectedEvent, slack.PresenceChangeEvent, slack.LatencyReport,
// slack.RTMError, slack.InvalidAuthEvent
func (q *Qubot) processIncomingEvent(event *slack.RTMEvent) {
	// Remember that the variable declared in the type switch will have the
	// corresponding type in each clause.
	switch e := event.Data.(type) {
	case *slack.MessageEvent:
		logger.Info("msg", "MessageEvent", "value", fmt.Sprintf("%v", e))
	}
}

// Close shutdowns the bot cleanly by signaling other goroutines to stop.
func (q *Qubot) Close() {
	close(q.shutdown)
	close(q.Quit)
	q.wg.Wait()
}
