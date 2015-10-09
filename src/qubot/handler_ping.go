package qubot

import "logger"

type pingHandler struct{}

func (h *pingHandler) Usage() string {
	return `ping - responds with "PONG"`
}

func (h *pingHandler) Pattern() string {
	return `(?i)ping`
}

func (h *pingHandler) Run(res *Response) error {
	logger.Debug("handler", "ping")
	return nil
}

// PingHandler ...
var PingHandler = &pingHandler{}
