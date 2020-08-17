package debug

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
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

// NewEntry returns a new Entry constructed from a request, response, name and latency.
func NewEntry(req *http.Request, res *http.Response, serviceName string, latency time.Duration) Entry {
	request := Request{
		Method:  req.Method,
		Route:   req.URL.Path,
		URL:     req.URL.String(),
		Headers: req.Header,
		Body:    readBody(req),
	}

	var response Response
	if res != nil {
		response = Response{
			Status:  res.StatusCode,
			Headers: res.Header,
			Body:    readBody(res),
			Latency: latency,
		}
	} else {
		response = Response{
			Status:  500,
			Latency: latency,
		}
	}

	return Entry{
		ServiceName: serviceName,
		Request:     request,
		Response:    response,
	}
}

func (m *Metadata) RecordEntry(e Entry) {
	if e.TraceId() != "" {
		m.Entries = append(m.Entries, e)
	} else {
		logrus.Infof("missing trace ID")
	}
}

// TraceIds returns an array of all distinct trace IDs.
func (m *Metadata) TraceIds() []string {
	ids := map[string]bool{}
	out := []string{}
	for _, e := range m.Entries {
		id := e.TraceId()
		if _, ok := ids[id]; !ok {
			ids[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
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

// GetEntriesByTrace returns an array of metadata entries with the given trace ID.
// Each entry represents a single request/response pair, upstream or downstream.
func (m *Metadata) GetBaseEntryByTrace(traceId string) Entry {
	for _, e := range m.Entries {
		if e.TraceId() == traceId && e.ServiceName == "PaymentServer" {
			return e
		}
	}
	return Entry{}
}

// TraceId returns the trace ID from the request header.
func (e Entry) TraceId() string {
	return e.Request.Headers.Get("traceId")
}

// Id returns the ID from the entry header.
func (e Entry) Id() string {
	return strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s_%s_%s", e.ServiceName, e.Request.Method, e.Request.Route)), "/", "_")
}

// readBody returns the content of a request/response body as a string, and reassigns the source's
// Body field to an unread copy of itself.
func readBody(r interface{}) string {
	switch t := r.(type) {
	case *http.Request:
	case *http.Response:
		if t.Body != nil {
			// TODO: Replace with some kind of tee-like caching passthrough mechanism.
			bodyBytes, _ := ioutil.ReadAll(t.Body)
			_ = t.Body.Close() //  must close
			t.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			return string(bodyBytes)
		}
	default:
		panic(fmt.Sprintf("input must be *http.Request or *http.Response, not %v", t))
	}
	return ""
}
