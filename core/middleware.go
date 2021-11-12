package core

import (
	"context"
	"net/http"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type middlewareCollection struct {
	admin  []func(handler http.Handler) http.Handler
	public []func(handler http.Handler) http.Handler
}

func prepareMiddleware(name string, promRegistry *prometheus.Registry, contextTimeout time.Duration) middlewareCollection {
	result := middlewareCollection{}
	result.addToBoth(Recoverer)
	result.addToBoth(common.Timeout(contextTimeout, http.HandlerFunc(timeoutHandler)))

	result.public = append(result.public, common.TraceabilityMiddleware)
	result.addToBoth(common.CoreRequestContextMiddleware)

	if promRegistry != nil {
		metricsMiddleware := metrics.NewHTTPServerMetricsMiddleware(promRegistry, name, metrics.GetChiPathPattern)
		result.addToBoth(metricsMiddleware)
	}

	return result
}

func (m *middlewareCollection) addToBoth(h ...func(handler http.Handler) http.Handler) {
	m.admin = append(m.admin, h...)
	m.public = append(m.public, h...)
}

func timeoutHandler(w http.ResponseWriter, r *http.Request) {
	common.HandleError(
		r.Context(),
		w,
		common.InternalError,
		"timeout expired while processing response",
		r.Context().Err(),
		func(context.Context, error) *common.HTTPError {
			return &common.HTTPError{
				HTTPCode:    http.StatusInternalServerError,
				Description: "timeout expired while processing response",
			}
		},
		nil,
	)
}
