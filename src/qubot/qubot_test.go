package qubot

import (
	"testing"
	"testutil"

	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

var testConfig = &Config{
	Database: &DatabaseConfig{
		Location: "/tmp/db",
	},
	Slack: &SlackConfig{
		Key: "12345",
	},
	Redmine: &RedmineConfig{
		URL:           "http://foobar.com",
		Key:           "12345",
		User:          "foobar",
		VerifyTLSCert: true,
	},
}

// InitTestQubot creates the Qubot object with the fake Slack client and returns
// a pointer to it.
func InitTestQubot() *Qubot {
	q := Qubot{
		config: testConfig,
		ready:  make(chan struct{}),
		done:   make(chan struct{}),
		users:  make(map[string]*slack.User),
	}
	q.client = newFakeSlackClient()
	q.rtm = q.client.NewRTM()

	q.db = &DB{}
	err := q.db.Open(testutil.Tempfile(), 0600)
	if err != nil {
		panic(err)
	}

	root := context.Background()
	q.ctx, q.cancel = context.WithCancel(root)
	return &q
}

// Ensures that Qubot starts properly.
func TestQubot_Start(t *testing.T) {
	q := InitTestQubot()
	done := make(chan struct{})
	reached := false
	go func() {
		q.Start()
		defer q.Close()
		reached = true
		close(done)
	}()
	select {
	case <-q.ready:
	case <-done:
		testutil.Assert(t, reached == true, "reached should be true")
	}
}

// Ensure that handlers can be registered with Qubot.
func TestQubot_Handle(t *testing.T) {
	q := InitTestQubot()
	h := &testHandler{}
	testutil.Assert(t, len(q.handlers) == 0, "len(q.handlers) should be 0")
	q.Handle(h)
	testutil.Assert(t, len(q.handlers) == 1, "len(q.handlers) should be 1")
}

// Ensure that Qubot notifies external receivers when the service shuts down.
func TestQubot_Done(t *testing.T) {
	q := InitTestQubot()
	done := make(chan struct{})
	closed := false
	go func() {
		q.Start()
		<-q.Done()
		closed = true
		close(done)
	}()
	testutil.Assert(t, closed == false, "closed should be false")
	q.Close()
	<-done // block until the goroutine is done.
	testutil.Assert(t, closed == true, "closed should be true")
}

func TestQubot_connect(t *testing.T) {
	q := InitTestQubot()
	client := q.client.(*fakeSlackClient)
	rtm := q.rtm.(*fakeSlackRTMClient)
	testutil.Assert(t, client.authTestCalled == false, "q.authTestCalled should be false")
	testutil.Assert(t, rtm.manageConnectionCalled == false, "q.manageConnectionCalled should be false")
	q.connect()
	q.wg.Wait()
	testutil.Assert(t, client.authTestCalled == true, "q.authTestCalled should be false")
	testutil.Assert(t, rtm.manageConnectionCalled == true, "q.manageConnectionCalled should be true")
}
