package qubot

import (
	"logger"
	"sync"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/juju/ratelimit"
	"github.com/nlopes/slack"
)

const rateLimit = 1.0
const pollWaitTime = 500 * time.Millisecond

// Messenger posts Qubot's messages to Slack respecting their API rate limit
// policy (see https://api.slack.com/docs/rate-limits for more details). We are
// assuming though that Slack is applying the rule per channel and not per
// client.
//
// If we have more than one message waiting to be delivered for a specific
// channel, we will group them together to avoid extra posting.
//
// TODO: allow bursts
// TODO: group messages
type Messenger struct {
	// rtm is the real-time websocket.
	rtm *slack.RTM

	// shutdown is a channel used to coordinate shutting down all the
	// goroutines in this object cleanly.
	shutdown chan struct{}

	// wg helps us to wait for the different channel goroutines.
	wg sync.WaitGroup

	// chq holds the channel queues
	chq *chqueue
}

// NewMessenger returns a new Messenger object.
func NewMessenger(rtm *slack.RTM) *Messenger {
	m := Messenger{
		rtm:      rtm,
		shutdown: make(chan struct{}),
		chq:      &chqueue{q: make(map[string]*queue.Queue)},
	}

	return &m
}

// Send puts the message in its corresponding queue.
func (m *Messenger) Send(msg *slack.OutgoingMessage) error {
	q, new, err := m.chq.add(msg)
	if err != nil || !new {
		return err
	}

	// When the queue is new we start a goroutine that will be responsible
	// of delivering the messages to its addressee.
	if new {
		m.startPoller(q)
	}

	return nil
}

// startPoller creates a new goroutine for a channel.
// TODO: confirm delivery or retry instead (circuitbreaker?)
func (m *Messenger) startPoller(q *queue.Queue) {
	logger.Debug("msg", "Starting new poller goroutine")
	go func() {
		m.wg.Add(1)
		tb := ratelimit.NewBucketWithRate(rateLimit, 1)
		for {
			select {
			case <-m.shutdown:
				m.wg.Done()
				return
			default:
				logger.Debug("msg", "Polling queue...")
				res, err := q.Poll(1, pollWaitTime)
				if err != nil {
					if err != queue.ErrTimeout {
						logger.Warn("Messenger.startPoller", "q.Poll error", "error", err)
					}
					continue
				}
				msg := res[0].(*slack.OutgoingMessage)
				m.rtm.SendMessage(msg)
				tb.Wait(1) // and relax for a bit!
			}
		}
	}()
}

// Close signals all the goroutines and waits until they are all done.
func (m *Messenger) Close() {
	close(m.shutdown)
	m.wg.Wait()
}

// chqueue wraps a map used to map channel IDs to the queue of messages pending
// to be delivered. It may have multiple readers and writers so we'll use a
// sync.RWMutex mutual exclusion lock to synchronize read/write access.
type chqueue struct {
	q   map[string]*queue.Queue
	mux sync.RWMutex
}

// add outgoing message to the corresponding channel queue, returns a pointer
// to the queue, a boolean where true means that the queue had to be created and
// false otherwise and an error if the message could not be added to the queue.
func (chq *chqueue) add(msg *slack.OutgoingMessage) (*queue.Queue, bool, error) {
	q := chq.get(msg.Channel)

	created := false
	if q == nil {
		q = queue.New(10)
		created = true

		chq.mux.Lock()
		chq.q[msg.Channel] = q
		chq.mux.Unlock()
	}

	err := q.Put(msg)

	return q, created, err
}

func (chq *chqueue) get(channel string) *queue.Queue {
	chq.mux.RLock()
	defer chq.mux.RUnlock()

	return chq.q[channel]
}
