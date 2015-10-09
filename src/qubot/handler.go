package qubot

import "golang.org/x/net/context"

// Handler is the interface that wraps the basic methods of Qubot handlers.
//
// It is up to the implementor to block in the Start method. However the method
// is expected to return when a cancellation signal is emitted via the context
// object. The implementor is signaled via the channel returned by
// context.Done().
type Handler interface {
	Start(context.Context) error
	Handle(Response)
}

// A HandlerMatcher is implemented by handlers that want to avoid
type HandlerMatcher interface {
	Match(Response) bool
}
