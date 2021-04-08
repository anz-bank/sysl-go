package convert

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/bmizerany/assert"
)

func TestQueryParam(t *testing.T) {
	type args struct {
		params url.Values
		key    string
		value  interface{}
	}
	tests := []struct {
		name string
		args args
		want url.Values
	}{
		{name: "HandleStrings", args: args{params: url.Values{}, key: "Key", value: "1234"}, want: url.Values{"key": []string{"1234"}}},
		{name: "HandleInt64s", args: args{params: url.Values{}, key: "Key", value: int64(1234)}, want: url.Values{"key": []string{"1234"}}},
		{name: "HandleFloat64s", args: args{params: url.Values{}, key: "Key", value: float64(1.234)}, want: url.Values{"key": []string{"1.234"}}},
		{name: "HandleBooleans", args: args{params: url.Values{}, key: "Key", value: "true"}, want: url.Values{"key": []string{"true"}}},
		{name: "HandleSliceStrings", args: args{params: url.Values{}, key: "Key", value: []string{"a", "b"}}, want: url.Values{"key": []string{"a", "b"}}},
		{name: "HandleSliceInt64", args: args{params: url.Values{}, key: "Key", value: []int64{1, 2, 3, 4}}, want: url.Values{"key": []string{"1", "2", "3", "4"}}},
		{name: "HandleNil", args: args{params: url.Values{}, key: "Key", value: nil}, want: nil},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			if got := EncodeQueryParam(tc.args.params, tc.args.key, tc.args.value); reflect.DeepEqual(got, tc.want) {
				t.Errorf("QueryParam() = %v, want %v", got, tc.want)
			}
			if got := EncodeQueryParam(tc.args.params, tc.args.key, tc.args.value).Encode(); reflect.DeepEqual(got, tc.want) {
				t.Errorf("QueryParam() = %v, want %v", got, tc.want)
			}
		})
	}
}

// This test demonstrates the underlying behaviour of net/url when duplicate keys are added.
// Values are consistent with the default OpenAPI 3.0 representation e.g duplicateKey=1&duplicateKey=2.
func TestQueryBuilderDuplicateString(t *testing.T) {
	u, _ := url.Parse("")
	q := u.Query()
	q.Add("duplicateKey", "1")
	q.Add("duplicateKey", "2")
	assert.Equal(t, q.Encode(), "duplicateKey=1&duplicateKey=2")
}

// This test demonstrates that an empty url.Values struct encodes without error.
func TestQueryBuilderEncodeEmpty(t *testing.T) {
	q := url.Values{}
	assert.Equal(t, q.Encode(), "")
}

// This test demonstrates the underlying behaviour of net/url when a comma separated value is used.
// Because the entire string is url encoded, values must be added individually.
func TestQueryBuilderCommaEncoding(t *testing.T) {
	u, _ := url.Parse("")
	q := u.Query()
	q.Add("csv", "1,2,3")
	assert.Equal(t, q.Encode(), "csv=1%2C2%2C3")
}
