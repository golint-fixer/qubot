package qubot

import "github.com/nlopes/slack"

// A Message represents a message received from Slack.
type Message struct {
	Msg *slack.Msg
}

// NewMessage returns a new Message that wraps a Slack message.
func NewMessage(msg *slack.Msg) *Message {
	return &Message{msg}
}
