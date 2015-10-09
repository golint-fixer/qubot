package qubot

import "github.com/nlopes/slack"

// Response represents the response from Qubot. It is meant to simplify the
// construction of responses from our handlers.
type Response struct {
	Qubot *Qubot

	// The message that was sent to obtain this Response.
	Msg *slack.Msg

	// The strings that
	Match []string
}

// Text returns the text property of the message
func (r *Response) Text() string {
	return r.Msg.Text
}
