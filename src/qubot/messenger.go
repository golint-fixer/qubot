package qubot

import (
	"logger"

	"github.com/juju/ratelimit"
	"github.com/nlopes/slack"
)

const rateLimit = 1.0

// Messenger delivers Qubot's messages to Slack making sure that the rate limit
// rules are respected (see https://api.slack.com/docs/rate-limits).
type Messenger struct {
	// rtm is the real-time websocket.
	rtm *slack.RTM

	// bucket for rate limiting.
	bucket *ratelimit.Bucket

	// router is the synchronization channel between the sender(s) and the
	// deliverer.
	router chan *slack.OutgoingMessage
}

// NewMessenger returns a new Messenger object.
func NewMessenger(rtm *slack.RTM) *Messenger {
	m := Messenger{
		rtm:    rtm,
		bucket: ratelimit.NewBucketWithRate(rateLimit, 1),
		router: make(chan *slack.OutgoingMessage, 0),
	}

	// Poller goroutine.
	go func() {
		for {
			msg := <-m.router
			logger.Info("Sending your message...")
			go m.rtm.SendMessage(msg)
			logger.Info("Message sent!")
			m.bucket.Wait(1)
		}
	}()

	return &m
}

// Send routes a message and locks until it's delivered.
func (m *Messenger) Send(msg *slack.OutgoingMessage) {
	m.router <- msg
}
