//nolint:goconst
package jsontime

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// NOTE: We do not test against yaml as that would introduce a third part dependency

func TestDurationUnmarshalJSONInvalidDuration(t *testing.T) {
	input := []byte(`{"duration": "BAD"}`)
	var target struct {
		Duration Duration `json:"duration"`
	}
	require.Error(t, json.Unmarshal(input, &target))
}

func TestDurationUnmarshalJSON(t *testing.T) {
	input := []byte(`{"duration": "1h"}`)
	var target struct {
		Duration Duration `json:"duration"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, time.Hour, target.Duration.Duration())
}

func TestDurationUnmarshalJSONNull(t *testing.T) {
	input := []byte(`{"duration": null}`)
	var target struct {
		Duration Duration `json:"duration"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, time.Duration(0), target.Duration.Duration())
}

func TestDurationUnmarshalJSONPointer(t *testing.T) {
	input := []byte(`{"duration": "1h"}`)
	var target struct {
		Duration *Duration `json:"duration"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, time.Hour, target.Duration.Duration())
}

func TestDurationUnmarshalJSONInvalidDurationPointer(t *testing.T) {
	input := []byte(`{"duration": "BAD"}`)
	var target struct {
		Duration *Duration `json:"duration"`
	}
	require.Error(t, json.Unmarshal(input, &target))
}

func TestDurationUnmarshalJSONPointerNull(t *testing.T) {
	input := []byte(`{"duration": null}`)
	var target struct {
		Duration *Duration `json:"duration"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Nil(t, target.Duration)
}

func TestDurationUnmarshalJSONPointerOmitEmpty(t *testing.T) {
	input := []byte(`{}`)
	var target struct {
		Duration *Duration `json:"duration,omitempty"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Nil(t, target.Duration)
}

func TestDurationMarshalJSON(t *testing.T) {
	d := struct {
		Duration Duration `json:"duration"`
	}{
		Duration: Duration(time.Hour),
	}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"duration":"1h0m0s"}`, string(marshalled))
}

func TestDurationMarshalJSONEmpty(t *testing.T) {
	d := struct {
		Duration Duration `json:"duration"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"duration":"0s"}`, string(marshalled))
}

func TestDurationMarshalJSONOmitEmpty(t *testing.T) {
	d := struct {
		Duration Duration `json:"duration,omitempty"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{}`, string(marshalled))
}

func TestDurationMarshalJSONPointer(t *testing.T) {
	dur := Duration(time.Hour)
	d := struct {
		Duration *Duration `json:"duration"`
	}{
		Duration: &dur,
	}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"duration":"1h0m0s"}`, string(marshalled))
}

func TestDurationMarshalJSONPointerEmpty(t *testing.T) {
	d := struct {
		Duration *Duration `json:"duration"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"duration":null}`, string(marshalled))
}

func TestDurationMarshalJSONPointerOmitEmpty(t *testing.T) {
	d := struct {
		Duration *Duration `json:"duration,omitempty"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{}`, string(marshalled))
}

func TestDurationMapstructureDecodeHookFunc(t *testing.T) {
	in := reflect.TypeOf("")
	out := reflect.TypeOf(Duration(0))
	data := "10m"

	result, err := DurationMapstructureDecodeHookFunc(in, out, data)
	require.NoError(t, err)

	// It is ok if the output is of type time.Duration
	// go will convert to jsontime.Duration when it goes to populate the target struct
	require.Equal(t, 10*time.Minute, result)
}

func TestDurationMapstructureDecodeHookFuncWithTimeDuration(t *testing.T) {
	in := reflect.TypeOf("")
	out := reflect.TypeOf(time.Duration(0))
	data := "10m"

	result, err := DurationMapstructureDecodeHookFunc(in, out, data)
	require.NoError(t, err)

	// It is ok if the output is of type time.Duration
	// go will convert to jsontime.Duration when it goes to populate the target struct
	require.Equal(t, 10*time.Minute, result)
}

func TestDurationMapstructureDecodeHookFuncWrongInputType(t *testing.T) {
	in := reflect.TypeOf(1)
	out := reflect.TypeOf(Duration(0))
	data := 10

	result, err := DurationMapstructureDecodeHookFunc(in, out, data)

	// Should not return error and pass straight through
	require.NoError(t, err)
	require.Equal(t, 10, result)
}

func TestDurationMapstructureDecodeHookFuncWrongOutputType(t *testing.T) {
	in := reflect.TypeOf("")
	out := reflect.TypeOf(1)
	data := "10m"

	result, err := DurationMapstructureDecodeHookFunc(in, out, data)

	// Should not return error and pass straight through
	require.NoError(t, err)
	require.Equal(t, "10m", result)
}

func TestYaml(t *testing.T) {
	raw := `
delay: 10m
`

	var target struct {
		Delay Duration `yaml:"delay"`
	}

	require.NoError(t, yaml.Unmarshal([]byte(raw), &target))
	assert.Equal(t, Duration(10*time.Minute), target.Delay)
}

func TestYamlNull(t *testing.T) {
	raw := `
delay: null
`
	var target struct {
		Delay Duration `yaml:"delay"`
	}

	require.NoError(t, yaml.Unmarshal([]byte(raw), &target))
	assert.Equal(t, Duration(0), target.Delay)
}
