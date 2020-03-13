package splunk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func retryUntilError(timeout time.Duration, f func() error) error {
	start := time.Now()
	for timeout > time.Since(start) {
		err := f()
		if err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func TestHookFireCanReturnErrorFromPriorFailedAttemptsToLogToSplunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprintln(w, "what you want: splunk. what you got: teapot")
	}))
	writer := Writer{
		Client: NewClient(server.Client(), server.URL, "", "", "", ""),
		// Configure the writer so it will "flush" (i.e. actually send to splunk) after every `Fire`.
		// Hack: we actually want to set this to zero, but the code
		// conflates a FlushThreshold of 0 with a FlushThreshold of defaultThreshold = 10 .
		FlushThreshold: -1,
		// Don't let the flush interval cause raciness
		FlushInterval: 5 * time.Minute,
	}

	hook := &Hook{
		Client: writer.Client,
		levels: []logrus.Level{logrus.ErrorLevel},
		writer: &writer,
	}

	firstError := hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel, Message: "application is broken"})
	assert.NoError(t, firstError)

	// HACK: keep waiting and calling Fire n the hope that the async error from the above Fire will end up
	// buffered in the error channel, so a subsequent Fire call can read it. This simulates time passing
	// between log calls while the real application would be doing useful work or waiting the next request or
	// response. It'd be neater to have an event-driven way to be notified about this, but if we wait for an
	// error using writer.Errors(), then we will consume the error, preventing it from being returned by
	// this second call to Fire.
	// Note that there is no guarantee about exactly when we will get a non-nil value returned from Fire,
	// it depends upon the vagaries of goroutine scheduling and how fast the machine is and load on the machine.

	f := func() error {
		return hook.Fire(&logrus.Entry{Level: logrus.ErrorLevel, Message: "application is still broken"})
	}

	secondError := retryUntilError(5.0*time.Second, f)

	assert.Error(t, secondError)
	remoteSplunkError := secondError.(*RemoteSplunkError)

	// Now, we expect that remoteSplunkError holds the error from the first Fire call that was sent asyncronously.
	assert.Regexp(t, regexp.MustCompile(`http://127\.0\.0\.1:.*`), remoteSplunkError.URL)
	assert.Equal(t, http.StatusTeapot, remoteSplunkError.Status)
	assert.Equal(t, "text/plain; charset=utf-8", remoteSplunkError.ContentType)
	assert.Equal(t, int64(44), remoteSplunkError.ContentLength)
	assert.Equal(t, "what you want: splunk. what you got: teapot\n", remoteSplunkError.BodySnippet)
}
