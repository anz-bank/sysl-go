package restlib

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anz-bank/sysl-go/convert"
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

func GetQueryParamForInt(r *http.Request, key string) (int64, error) {
	result, err := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func GetQueryParamForBool(r *http.Request, key string) (bool, error) {
	result, err := strconv.ParseBool(r.URL.Query().Get(key))
	if err != nil {
		return false, err
	}
	return result, nil
}

func GetQueryParamForTime(r *http.Request, key string) (convert.JSONTime, error) {
	result, err := time.Parse("2006-01-02T15:04:05.000-0700", key)
	if err != nil {
		if result, err = time.Parse(time.RFC3339, key); err != nil {
			return convert.JSONTime{Time: time.Time{}}, err
		}
	}
	return convert.JSONTime{Time: result}, nil
}

func GetHeaderParam(r *http.Request, key string) string {
	return r.Header.Get(key)
}
