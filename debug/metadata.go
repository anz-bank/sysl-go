package debug

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var MetadataStore = Metadata{}

// Request captures request metadata.
type Request struct {
	Method  string      `json:"method,omitempty"`
	Route   string      `json:"route,omitempty"`
	Headers http.Header `json:"request,omitempty"`
	Body    string      `json:"reqBody,omitempty"`
	URL     string      `json:"url,omitempty"`
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
	ServiceName string   `json:"serviceName,omitempty"`
	Request     Request  `json:"request,omitempty"`
	Response    Response `json:"response,omitempty"`
}

// Metadata records all interaction entries.
type Metadata struct {
	Entries []Entry
}

func (m *Metadata) RecordEntry(e Entry) {
	if e.TraceId() != "" {
		m.Entries = append(m.Entries, e)
	} else {
		logrus.Infof("missing trace ID")
	}
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
	m.RecordEntry(entry)
}

// GetEntriesByTrace returns an array of metadata entries with the given trace ID.
// Each entry represents a single request/response pair, upstream or downstream.
func (m *Metadata) GetEntriesByTrace(traceId string) []Entry {
	es := []Entry{}
	for _, e := range m.Entries {
		if e.TraceId() == traceId {
			es = append(es, e)
		}
	}
	return es
}

// TraceId returns the trace ID from the request header.
func (e Entry) TraceId() string {
	return e.Request.Headers.Get("traceId")
}
