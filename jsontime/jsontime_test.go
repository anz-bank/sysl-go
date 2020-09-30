package jsontime

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// utility function to produce a random but consistent time struct across machines.
func randomTime() time.Time {
	year := rand.Intn(100) + 1970
	month := time.Month(rand.Intn(12) + 1)
	day := rand.Intn(28) + 1
	hour := rand.Intn(24)
	min := rand.Intn(60)
	sec := rand.Intn(60)
	loc := time.UTC
	return time.Date(year, month, day, hour, min, sec, 0, loc)
}

func TestTimeUnmarshalJSONInvalidDuration(t *testing.T) {
	input := []byte(`{"time": "BAD"}`)
	var target struct {
		Time Time `json:"time"`
	}
	require.Error(t, json.Unmarshal(input, &target))
}

func TestTimeUnmarshalJSON(t *testing.T) {
	now := randomTime()
	input := []byte(fmt.Sprintf(`{"time": "%s"}`, now.Format(time.RFC3339Nano)))
	var target struct {
		Time Time `json:"time"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, now, target.Time.Time())
}

func TestTimeUnmarshalJSONNull(t *testing.T) {
	input := []byte(`{"time": null}`)
	var target struct {
		Time Time `json:"time"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, time.Time{}, target.Time.Time())
}

func TestTimeUnmarshalJSONPointer(t *testing.T) {
	now := randomTime()
	input := []byte(fmt.Sprintf(`{"time": "%s"}`, now.Format(time.RFC3339Nano)))
	var target struct {
		Time *Time `json:"time"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Equal(t, now, target.Time.Time())
}

func TestTimeUnmarshalJSONInvalidDurationPointer(t *testing.T) {
	input := []byte(`{"time": "BAD"}`)
	var target struct {
		Time *Time `json:"time"`
	}
	require.Error(t, json.Unmarshal(input, &target))
}

func TestTimeUnmarshalJSONPointerNull(t *testing.T) {
	input := []byte(`{"time": null}`)
	var target struct {
		Time *Time `json:"time"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Nil(t, target.Time)
}

func TestTimeUnmarshalJSONPointerOmitEmpty(t *testing.T) {
	input := []byte(`{}`)
	var target struct {
		Time *Time `json:"time,omitempty"`
	}
	require.NoError(t, json.Unmarshal(input, &target))
	require.Nil(t, target.Time)
}

func TestTimeMarshalJSON(t *testing.T) {
	now := randomTime()
	d := struct {
		Time Time `json:"time"`
	}{
		Time: Time(now),
	}
	timeMarshalled, err := json.Marshal(now)
	require.NoError(t, err)
	require.NotEmpty(t, timeMarshalled)
	marshalled, err := json.Marshal(d.Time)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, string(timeMarshalled), string(marshalled))
}

func TestTimeMarshalJSONEmpty(t *testing.T) {
	d := struct {
		Time Time `json:"time"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"time":"0001-01-01T00:00:00Z"}`, string(marshalled))
}

func TestTimeMarshalJSONOmitEmpty(t *testing.T) {
	d := struct {
		Time Time `json:"time,omitempty"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"time":"0001-01-01T00:00:00Z"}`, string(marshalled))
}

func TestTimeMarshalJSONPointer(t *testing.T) {
	now := randomTime()
	tim := Time(now)
	d := struct {
		Time *Time `json:"time"`
	}{
		Time: &tim,
	}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	expected := fmt.Sprintf(`{"time":"%s"}`, now.Format(time.RFC3339Nano))
	assert.Equal(t, expected, string(marshalled))
}

func TestTimeMarshalJSONPointerEmpty(t *testing.T) {
	d := struct {
		Time *Time `json:"time"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{"time":null}`, string(marshalled))
}

func TestTimeMarshalJSONPointerOmitEmpty(t *testing.T) {
	d := struct {
		Time *Time `json:"time,omitempty"`
	}{}
	marshalled, err := json.Marshal(d)
	require.NoError(t, err)
	require.NotNil(t, marshalled)
	assert.Equal(t, `{}`, string(marshalled))
}
