package debug

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// Request captures request metadata.
type Request struct {
	Method  string      `json:"method,omitempty"`
	Route   string      `json:"route,omitempty"`
	Headers http.Header `json:"request,omitempty"`
	Body    string      `json:"reqBody,omitempty"`
}

// Response captures response metadata.
type Response struct {
	Status  int           `json:"status,omitempty"`
	Latency time.Duration `json:"latency,omitempty"`
	Headers http.Header   `json:"resHeader,omitempty"`
	Body    string        `json:"response,omitempty"`
}

// Entry records metadata for a single interaction.
type Entry struct {
	Request  Request  `json:"request,omitempty"`
	Response Response `json:"response,omitempty"`
}

// Metadata records all interaction entries.
type Metadata struct {
	Entries []Entry
}

// Record adds the metadata for a call to the Metadata store.
func (m *Metadata) Record(req *http.Request, method string, route string, reqBody string, res string, status int, responseHeader http.Header, latency time.Duration) {
	entry := Entry{
		Request{
			Headers: req.Header,
			Method:  method, Route: route,
			Body: reqBody,
		}, Response{
			Headers: responseHeader,
			Body:    res,
			Status:  status,
			Latency: latency,
		},
	}
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
		if entry.TraceId() == traceId {
			return entry
		}
	}
	return Entry{}
}

// TraceId returns the trace ID from the request header.
func (e Entry) TraceId() string {
	return e.Request.Headers.Get("traceId")
}
