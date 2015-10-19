package qubot

// IncomingMessage is a message coming from the adapter. It is public so handler
// implementors can access to its values.
type IncomingMessage struct {
	Type    string
	User    string
	Channel string
	Text    string
}

// OutgoingMessage is a message that Qubot is sending to an adapter.
type OutgoingMessage struct {
	text    string
	channel string
}

// NewOutgoingMessage returns a pointer to a new OutgoingMessage object.
func NewOutgoingMessage(text string) *OutgoingMessage {
	m := OutgoingMessage{}
	if text != "" {
		m.text = text
	}
	return &m
}
