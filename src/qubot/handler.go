package qubot

import (
	"fmt"
	"regexp"
)

// Handler is an interface for objects to implement in order to respond to
// messages.
type Handler interface {
	Handle(resp *Response) error
}

// NewHandler checks whether h implements the handler interface, wrapping it in a FullHandler
func NewHandler(h interface{}) (Handler, error) {
	switch v := h.(type) {
	case fullHandler:
		return &FullHandler{handler: v}, nil
	case Handler:
		return v, nil
	default:
		return nil, fmt.Errorf("%v does not implement the handler interface", v)
	}
}

// fullHandler is an interface for objects that wish to supply their own define
// methods
type fullHandler interface {
	Usage() string
	Pattern() string
	Run(resp *Response) error
}

// FullHandler declares common functions shared by all handlers
type FullHandler struct {
	handler fullHandler
}

// Regexp func
func (h *FullHandler) Regexp() *regexp.Regexp {
	return handlerRegexp(h.handler.Pattern())
}

// Match func
func (h *FullHandler) Match(resp *Response) bool {
	return handlerMatch(h.Regexp(), resp.Msg.Text)
}

// Handle implements the FullHandler interface
func (h *FullHandler) Handle(resp *Response) error {
	switch {
	// Handle the response without matching
	case h.handler.Pattern() == "":
		return h.handler.Run(resp)
		// Handle the response after finding matches
	case h.Match(resp):
		resp.Match = h.Regexp().FindAllStringSubmatch(resp.Text(), -1)[0]
		return h.handler.Run(resp)
	default:
		return nil
	}
}

func handlerMatch(r *regexp.Regexp, text string) bool {
	if !r.MatchString(text) {
		return false
	}
	return true
}

func handlerRegexp(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
