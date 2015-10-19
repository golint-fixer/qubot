package qubot

import (
	"logger"
	"sync"
	"time"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/juju/ratelimit"
)

const msnRateLimit = 1.0
const msnPollWaitTime = 500 * time.Millisecond

// Messenger interface
type Messenger interface {
	Send(msg *OutgoingMessage) error
	Close()
}

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
type messenger struct {
	wg       sync.WaitGroup
	shutdown chan struct{}
	adapter  adapter
	chq      *chqueue
}

// InitMessenger returns a new Messenger object.
func InitMessenger(adapter adapter) Messenger {
	m := messenger{
		shutdown: make(chan struct{}),
		adapter:  adapter,
		chq:      &chqueue{q: make(map[string]*queue.Queue)},
	}

	return &m
}

// Send puts the message in its corresponding queue.
func (m *messenger) Send(msg *OutgoingMessage) error {
	q, new, err := m.chq.add(msg)
	if err != nil || !new {
		return err
	}

	// When the queue is new we start a goroutine that will be responsible
	// of delivering the messages to its addressee.
	if new {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.startPoller(q)
		}()
	}

	return nil
}

// startPoller creates a new goroutine for a channel.
// TODO: confirm delivery or retry instead (circuitbreaker?)
func (m *messenger) startPoller(q *queue.Queue) {
	logger.Debug("messenger", "Starting new poller goroutine")
	tb := ratelimit.NewBucketWithRate(msnRateLimit, 1)
	for {
		select {
		case <-m.shutdown:
			logger.Debug("messenger", "Closing poller")
			return
		default:
			res, err := q.Poll(1, msnPollWaitTime)
			if err != nil {
				if err != queue.ErrTimeout {
					logger.Warn("messenger", "startPoller", "error", err)
				}
				continue
			}
			msg := res[0].(*OutgoingMessage)
			m.adapter.Send(msg)
			tb.Wait(1) // and relax for a bit!
		}
	}
}

// Close signals all the goroutines and waits until they are all done.
func (m *messenger) Close() {
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
func (chq *chqueue) add(msg *OutgoingMessage) (*queue.Queue, bool, error) {
	q := chq.get(msg.channel)

	created := false
	if q == nil {
		q = queue.New(10)
		created = true

		chq.mux.Lock()
		chq.q[msg.channel] = q
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
