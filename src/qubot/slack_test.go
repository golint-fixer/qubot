package qubot

import "github.com/nlopes/slack"

type fakeSlackClient struct {
	authTestCalled bool
}

func newFakeSlackClient() slackClient {
	return &fakeSlackClient{}
}

func (c *fakeSlackClient) NewRTM() slackRTMClient {
	return &fakeSlackRTMClient{}
}

func (c *fakeSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	c.authTestCalled = true
	return &slack.AuthTestResponse{
		URL: "http://foobar.com",
	}, nil
}

type fakeSlackRTMClient struct {
	manageConnectionCalled bool
}

func (c *fakeSlackRTMClient) ManageConnection() {
	c.manageConnectionCalled = true
}

func (c *fakeSlackRTMClient) Disconnect() error {
	return nil
}

func (c *fakeSlackRTMClient) GetInfo() *slack.Info {
	return nil
}

func (c *fakeSlackRTMClient) SendMessage(msg *slack.OutgoingMessage) {}

func (c *fakeSlackRTMClient) Events() chan slack.RTMEvent {
	ch := make(chan slack.RTMEvent, 1)
	return ch
}
