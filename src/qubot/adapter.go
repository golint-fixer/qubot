package qubot

import (
	"fmt"
	"sync"

	"github.com/nlopes/slack"
)

type adapter interface {
	Connect() error
	Send(*OutgoingMessage) error
	Events() chan *event
	Close()
}

type slackAdapter struct {
	rtm      *slack.RTM
	wg       sync.WaitGroup
	ch       chan *event
	shutdown chan struct{}
}

func newSlackAdapter(key string) adapter {
	client := slack.New(key)
	a := slackAdapter{
		rtm:      client.NewRTM(),
		ch:       make(chan *event),
		shutdown: make(chan struct{}),
	}
	return &a
}

func (a *slackAdapter) Connect() error {
	_, err := a.rtm.Client.AuthTest()
	if err != nil {
		return fmt.Errorf("Authentication test failed: %s", err)
	}

	a.wg.Add(2)

	go func() {
		defer a.wg.Done()
		// This function returns after calling Disconnect().
		a.rtm.ManageConnection()
	}()

	go func() {
		defer a.wg.Done()
		for {
			select {
			case event := <-a.rtm.IncomingEvents:
				a.ch <- a.wrapEvent(&event)
			case <-a.shutdown:
				return
			}
		}
	}()

	return nil
}

func (a *slackAdapter) Send(*OutgoingMessage) error {
	return nil
}

func (a *slackAdapter) Events() chan *event {
	return a.ch
}

func (a *slackAdapter) Close() {
	a.rtm.Disconnect()
	close(a.shutdown)
	a.wg.Wait()
}

// wrapEvent wraps a *slack.RTMEvent with a known internal event.
func (a *slackAdapter) wrapEvent(ev *slack.RTMEvent) *event {
	var e *event
	switch evt := ev.Data.(type) {
	default:
		e = newSlackUnknownEvent(ev)
	case *slack.MessageEvent:
		e = newSlackMessageEvent(ev, evt)
	case *slack.ConnectingEvent:
		e = &event{
			kind: ev.Type,
			data: ev.Data,
		}
	case *slack.ConnectedEvent:
		e = &event{
			kind: ev.Type,
			data: ev.Data,
		}
	case *slack.HelloEvent:
		e = &event{
			kind: ev.Type,
			data: ev.Data,
		}
	case *slack.LatencyReport:
		e = &event{
			kind: ev.Type,
			data: ev.Data,
		}
	case *slack.InvalidAuthEvent: // TODO: this should panic unless you can recover
	case *slack.RTMError:
	case *slack.AckErrorEvent:
	case *slack.ConnectionErrorEvent:
	case *slack.DisconnectedEvent:
	case *slack.MessageTooLongEvent:
	case *slack.OutgoingErrorEvent:
		e = newSlackErrorEvent(ev)
	}
	return e
}

func newSlackMessageEvent(ev *slack.RTMEvent, evt *slack.MessageEvent) *event {
	event := event{
		kind: ev.Type,
		data: ev.Data,
		inner: &messageEvent{
			message: &IncomingMessage{
				Type:    evt.Type,
				User:    evt.User,
				Channel: evt.Channel,
				Text:    evt.Text,
			},
		},
	}
	return &event
}

func newSlackUnknownEvent(ev *slack.RTMEvent) *event {
	event := event{
		kind:  ev.Type,
		data:  ev.Data,
		inner: &unknownEvent{},
	}
	return &event
}

func newSlackErrorEvent(ev *slack.RTMEvent) *event {
	event := event{
		kind:  ev.Type,
		data:  ev.Data,
		inner: &errorEvent{},
	}
	return &event
}
