package qubot

import (
	"testing"

	"testutil"
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

// Ensure that Qubot does not block when it's started.
func TestQubot_Start(t *testing.T) {
	q := Init(testConfig)
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
	h := &testHandler{}
	q := Init(testConfig)
	testutil.Assert(t, len(q.handlers) == 0, "len(q.handlers) should be 0")
	q.Handle(h)
	testutil.Assert(t, len(q.handlers) == 1, "len(q.handlers) should be 1")
}

// Ensure that Qubot notifies external receivers when the service shuts down.
func TestQubot_Done(t *testing.T) {
	q := Init(testConfig)
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
