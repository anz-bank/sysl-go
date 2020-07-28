package splunk

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// Hook is a logrus hook for splunk.
type Hook struct {
	Client *Client
	levels []logrus.Level
	writer *Writer
}

// NewHook creates a new hook.
// client - splunk client instance (use NewClient)
// level - log level.
func NewHook(client *Client, levels []logrus.Level) *Hook {
	return &Hook{client, levels, client.Writer()}
}

// Fire triggers a splunk event.
func (h *Hook) Fire(entry *logrus.Entry) error {
	formatter := logrus.JSONFormatter{}

	jsontext, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	value := map[string]interface{}{}

	if err := json.Unmarshal(jsontext, &value); err != nil {
		return err
	}

	if err := h.writer.Write(value); err != nil {
		return err
	}
	// We don't know if we have hit any errors from the log entry we queued for a send above, as
	// that will be processed asynchronously. But, it is possible there were errors from previous
	// attempts to send events to splunk, during previous Fire calls.
	// Check the error buffer. If there are any errors, pop one and return it.

	// Note: when Logrus sees that this hook Fire call has returned an error, it wont log it
	// normally, it will just write the error message to stderr.
	select {
	case err := <-h.writer.Errors():
		return err
	default:
		return nil
	}
}

// Levels returns the levels equired for logrus hook implementation.
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}
