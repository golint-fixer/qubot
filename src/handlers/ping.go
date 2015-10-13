package handlers

import (
	"fmt"
	"logger"
	"qubot"

	"golang.org/x/net/context"
)

// PingHandler is...
var PingHandler *pingHandler

func init() {
	PingHandler = &pingHandler{
		done: make(chan struct{}),
	}
}

// pingHandler implements the Handler interface.
type pingHandler struct {
	ctx  context.Context
	done chan struct{}
}

func (h *pingHandler) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("pingHandler: Cancelled!")
			return ctx.Err()
		case <-h.done:
			fmt.Println("pingHandler: Closed!")
			return nil
		}
	}
}

func (h *pingHandler) Handle(r qubot.Response, msg *qubot.Message) {
	logger.Debug("pingHandler", "Received a message!")
}
