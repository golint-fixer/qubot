package qubot

import (
	"fmt"
	"io"
)

// A Response interface is used by a handler to construct a response.
type Response interface {
	Write(io.Reader)
}

// response
type response struct {
	msn  Messenger
	text []byte
}

// NewResponse returns a new response.
func NewResponse(msn Messenger) Response {
	r := response{msn, make([]byte, 100)}
	return &r
}

func (r *response) Write(reader io.Reader) {
	fmt.Println(reader)
}
