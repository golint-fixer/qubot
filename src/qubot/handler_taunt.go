package qubot

import "fmt"

type tauntHandler struct{}

func (h *tauntHandler) Usage() string {
	return ``
}

func (h *tauntHandler) Pattern() string {
	return `qubot`
}

func (h *tauntHandler) Run(resp *Response) error {
	// return res.Send("PONG")

	user, err := resp.Qubot.client.GetUserInfo(resp.Msg.User)
	if err != nil {
		return err
	}
	if user.Name == "slackbot" || user.Name == "qubot" {
		return nil
	}
	msg := fmt.Sprintf("@%s: don't taunt qubot!", user.Name)
	resp.Qubot.rtm.SendMessage(resp.Qubot.rtm.NewOutgoingMessage(msg, resp.Msg.Channel))

	return nil
}

// TauntHandler ...
var TauntHandler = &tauntHandler{}