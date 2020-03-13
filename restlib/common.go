package restlib

import (
	"net/http"

	"github.com/go-chi/chi"
)

func GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func GetHeaderParam(r *http.Request, key string) string {
	return r.Header.Get(key)
}
