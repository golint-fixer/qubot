package qubot

import "golang.org/x/net/context"

// testHandler implements the Handler interface.
type testHandler struct {
	done chan struct{}
}

func (h *testHandler) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
		case <-h.done:
			return nil
		}
	}
}

func (h *testHandler) Handle(msg Message, r Response) {}

func (h *testHandler) Stop() {
	close(h.done)
}
