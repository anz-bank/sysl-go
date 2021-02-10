package restlib

import "github.com/google/go-querystring/query"

func urlencode(v interface{}) ([]byte, error) {
	values, err := query.Values(v)
	if err != nil {
		return nil, err
	}
	return []byte(values.Encode()), nil
}
