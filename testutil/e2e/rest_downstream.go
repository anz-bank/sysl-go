package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/anz-bank/sysl-go/syslgo"
	"github.com/go-chi/chi"
)

const (
	// DownstreamTimeout is only used for mocked downstreams, and is as small as possible to make the timeout tests
	// quicker, but give them time to actually succeed.
	DownstreamTimeout = time.Millisecond * 400
)

type restDownstream struct {
	t        syslgo.TestingT
	server   *httptest.Server
	r        chi.Router
	hostname string
	handlers map[string]*restOrderedHandlers
}

func newBackEnd(t syslgo.TestingT, hostname string) *restDownstream {
	be := &restDownstream{
		t:        t,
		r:        chi.NewRouter(),
		handlers: map[string]*restOrderedHandlers{},
		hostname: hostname,
	}
	be.server = httptest.NewServer(be.r)
	be.server.Client().Timeout = DownstreamTimeout

	return be
}

func (d *restDownstream) init(method, path string) *restOrderedHandlers {
	h := &restOrderedHandlers{
		Mutex:    sync.Mutex{},
		t:        d.t,
		handlers: nil,
	}
	d.handlers[fmt.Sprintf("%s %s", method, path)] = h
	d.r.Method(method, path, h)

	return h
}

func (d *restDownstream) getClient() (*http.Client, string, error) {
	return d.server.Client(), d.server.URL, nil
}

func (d *restDownstream) close() {
	for methodAndPath, h := range d.handlers {
		h.assertCompleted(d.hostname, methodAndPath)
	}
	d.server.Close()
}
