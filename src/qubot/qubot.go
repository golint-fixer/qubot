package qubot

import (
	"fmt"
	"logger"
	"sync"
)

var ignoreUserList = []string{"USLACKBOT"}

// Qubot at your service!
type Qubot struct {
	config  *Config
	m       Messenger
	db      *DB
	es      *eventService
	adapter adapter

	wg       sync.WaitGroup
	shutdown chan struct{}
	Closed   chan struct{}
}

// Init creates the Qubot object and returns a pointer to it.
func Init(config *Config) *Qubot {
	q := Qubot{
		config:   config,
		shutdown: make(chan struct{}),
		Closed:   make(chan struct{}),
	}

	q.db = &DB{}
	err := q.db.Open(q.config.Database.Location, 0600)
	if err != nil {
		panic(err)
	}

	return &q
}

// Run the service.
func (q *Qubot) Run() error {
	q.adapter = newSlackAdapter(q.config.Slack.Key)
	err := q.adapter.Connect()
	if err != nil {
		return fmt.Errorf("Adapter failed: %s", err)
	}

	q.m = InitMessenger(q.adapter)

	q.es = newEventService(q)

	return nil
}

// Report makes Qubot log some vitals about the service.
// Nothing serious here yet.
func (q *Qubot) Report() {
	logger.Info("qubot", "Status report!")
}

// Close closes the service and blocks until it's completely done.
func (q *Qubot) Close() {
	defer close(q.Closed)
	close(q.shutdown)

	q.m.Close()

	q.es.Close()

	q.adapter.Close()

	q.wg.Wait()
}
