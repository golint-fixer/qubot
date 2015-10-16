package qubot

import (
	"fmt"

	"github.com/nlopes/slack"
)

// slackClient is the interface of the Slack client.
type slackClient interface {
	NewRTM() slackRTMClient
	AuthTest() (*slack.AuthTestResponse, error)
}

// slackRTMClient is the interface of the Slack RTM client.
type slackRTMClient interface {
	ManageConnection()
	Disconnect() error
	GetInfo() *slack.Info
	SendMessage(msg *slack.OutgoingMessage)
	Events() chan slack.RTMEvent
}

type slackClientStruct struct {
	*slack.Client
}

func newSlackClient(key string) slackClient {
	return &slackClientStruct{slack.New(key)}
}

func (c slackClientStruct) NewRTM() slackRTMClient {
	return &slackRTMClientStruct{}
}

func (c slackClientStruct) AuthTest() (*slack.AuthTestResponse, error) {
	resp, err := c.AuthTest()
	return resp, err
}

type slackRTMClientStruct struct {
}

func (c *slackRTMClientStruct) ManageConnection() {
	fmt.Println("asdf")
	c.ManageConnection()
}

func (c *slackRTMClientStruct) Disconnect() error {
	return c.Disconnect()
}

func (c *slackRTMClientStruct) GetInfo() *slack.Info {
	return c.GetInfo()
}

func (c *slackRTMClientStruct) SendMessage(msg *slack.OutgoingMessage) {
	c.SendMessage(msg)
}

func (c *slackRTMClientStruct) Events() chan slack.RTMEvent {
	ch := make(chan slack.RTMEvent)
	return ch
}
