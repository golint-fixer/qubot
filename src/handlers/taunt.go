package handlers

import (
	"fmt"
	"logger"
	"qubot"

	"golang.org/x/net/context"
)

// TauntHandler is...
var TauntHandler *tauntHandler

func init() {
	TauntHandler = &tauntHandler{
		done: make(chan struct{}),
	}
}

// tauntHandler implements the Handler interface.
type tauntHandler struct {
	ctx  context.Context
	done chan struct{}
}

func (h *tauntHandler) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("tauntHandler: Cancelled!")
			return ctx.Err()
		case <-h.done:
			fmt.Println("tauntHandler: Closed!")
			return nil
		}
	}
}

func (h *tauntHandler) Handle(r qubot.Response) {
	logger.Debug("tauntHandler", "Received a message!")
}
