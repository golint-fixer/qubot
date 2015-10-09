package qubot

import "github.com/nlopes/slack"

// Response ...
type Response interface{}

type response struct {
	msg *slack.Msg
}

// NewResponse returns a new Response.
func NewResponse(msg *slack.Msg) Response {
	return &response{msg}
}
