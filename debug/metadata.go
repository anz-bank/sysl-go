package debug

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Entry records metadata for a single interaction.
type Entry struct {
	Request  http.Header   `json:"request,omitempty"`
	Response string        `json:"response,omitempty"`
	Status   int           `json:"status,omitempty"`
	Latency  time.Duration `json:"latency,omitempty"`
}

// Metadata records all interaction entries.
type Metadata struct {
	Entries []Entry
}

// Record adds the metadata for a call to the Metadata store.
func (m *Metadata) Record(req *http.Request, res string, status int, latency time.Duration) {
	entry := Entry{Request: req.Header, Response: res, Status: status, Latency: latency}
	if entry.TraceId() != "" {
		m.Entries = append(m.Entries, entry)
	} else {
		logrus.Infof("missing trace ID")
	}
}

// GetEntryByTrace returns the metadata entry with the given trace ID, or an empty entry if there's
// no match.
func (m *Metadata) GetEntryByTrace(traceId string) Entry {
	for _, entry := range m.Entries {
		if entry.Request.Get("traceId") == traceId {
			return entry
		}
	}
	return Entry{}
}

// TraceId returns the trace ID from the request header.
func (e Entry) TraceId() string {
	return e.Request.Get("traceId")
}
