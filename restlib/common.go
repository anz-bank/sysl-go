package restlib

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func GetURLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func GetURLParamForInt(r *http.Request, key string) int64 {
	result, _ := strconv.ParseInt(chi.URLParam(r, key), 10, 64)
	return result
}

func GetQueryParam(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func GetHeaderParam(r *http.Request, key string) string {
	return r.Header.Get(key)
}
