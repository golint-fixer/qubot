package qubot

import "logger"

type redmineHandler struct{}

func (h *redmineHandler) Usage() string {
	return `redmine - allows to interact with Redmine`
}

func (h *redmineHandler) Pattern() string {
	return `^redmine.*`
}

func (h *redmineHandler) Run(res *Response) error {
	logger.Debug("handler", "redmine")
	return nil
}

// RedmineHandler ...
var RedmineHandler = &redmineHandler{}
