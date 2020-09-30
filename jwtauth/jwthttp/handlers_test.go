package jwthttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testerr int

func (testerr) Error() string     { return "error" }
func (t testerr) HTTPStatus() int { return int(t) }

func TestDefaultHandlerResponds403ForGenericError(t *testing.T) {
	rec := httptest.NewRecorder()
	DefaultUnauthHandler(rec, nil, errors.New("error"))
	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDefaultHandlerUsesErrorStatusIfDefined(t *testing.T) {
	rec := httptest.NewRecorder()
	DefaultUnauthHandler(rec, nil, testerr(http.StatusBadGateway))
	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestHiddenEndpointHandlerReturns404(t *testing.T) {
	rec := httptest.NewRecorder()
	HiddenEndpoint(rec, nil, nil)
	resp := rec.Result()
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
