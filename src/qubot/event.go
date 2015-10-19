package qubot

import (
	"logger"
	"sync"
	"time"

	"golang.org/x/net/context"
)

var eventHandlerTimeout = time.Second * 10

// eventService is the service responsible of handling incoming events from the
// adapter. The events and their callbacks are processed in separate goroutines.
type eventService struct {
	shutdown chan struct{}
	wg       sync.WaitGroup
	ch       chan *event
	q        *Qubot
}

func newEventService(q *Qubot) *eventService {
	es := eventService{
		q:        q,
		shutdown: make(chan struct{}),
	}
	es.Listen()
	return &es
}

func (es *eventService) Listen() {
	es.wg.Add(1)
	go func() {
		es.wg.Done()
		for {
			select {
			case event := <-es.q.adapter.Events():
				es.receiveEvent(event)
			case <-es.shutdown:
				return
			}
		}
	}()
}

// receiveEvent processes the event in a new goroutine. The goroutine will be
// abandoned if the timeout is reached. It will be reported so the code that runs
// in the goroutine can be eventually fixed. Remember: a goroutine cannot be
// stopped but you can try to make it return.
func (es *eventService) receiveEvent(e *event) {
	es.wg.Add(1)
	go func() {
		defer es.wg.Done()

		ctx, cancel := context.WithCancel(context.Background())
		errc := make(chan error, 1)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer es.recover()
			errc <- es.handleEvent(ctx, e)
		}()

		select {
		case err := <-errc:
			if err != nil {
				logger.Warn("qubot", "receiveEvent", "error", err)
			}
		case <-time.After(eventHandlerTimeout):
			logger.Warn("qubot", "receiveEvent", "Event handler timed out", e)
		case <-es.shutdown:
			cancel()
			wg.Wait()
		}
	}()
}

func (es *eventService) handleEvent(_ context.Context, e *event) error {
	var err error
	switch e.inner.(type) {
	case *unknownEvent:
		logger.Debug("qubot", "Unknown event received", "type", e.kind)
	case *errorEvent:
		logger.Warn("qubot", "Error event received")
	case *connectingEvent:
		logger.Debug("qubot", "Connection attempt")
	case *connectedEvent:
		logger.Info("qubot", "Connected to Slack!")
	case *helloEvent:
		logger.Info("qubot", "Slack sent greetings!")
	case *latencyReportEvent:
		logger.Debug("qubot", "Latency report")
	case *messageEvent:
		logger.Debug("qubot", "Message received")
	}
	return err
}

func (es *eventService) recover() {
	p := recover()
	if p != nil {
		logger.Warn("qubot", "eventService", "Panic! Regained control.", p)
	}
}

func (es *eventService) Close() {
	close(es.shutdown)
	es.wg.Wait()
}

type event struct {
	kind  string      // String with the name of the event.
	data  interface{} // The original event if any.
	inner interface{} // Our internal event.
}

type unknownEvent struct{}
type errorEvent struct{}
type connectingEvent struct{}
type connectedEvent struct{}
type helloEvent struct{}
type latencyReportEvent struct{}

type messageEvent struct {
	message *IncomingMessage
}
