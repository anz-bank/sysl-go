package convert

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/stretchr/testify/require"
)

func getContext() context.Context {
	ctx, _ := common.NewTestContextWithLoggerHook()
	return ctx
}

func TestStringToIntPtr_Invalid(t *testing.T) {
	result, err := StringToIntPtr(getContext(), "integer")
	require.Equal(t, "ServerError(Kind=Internal Server Error, Message=invalid integer format: integer, Cause=%!s(<nil>))", err.Error())
	require.Nil(t, result)
}

func TestStringToIntPtr_Empty(t *testing.T) {
	result, err := StringToIntPtr(getContext(), "")
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestStringToIntPtr_Valid(t *testing.T) {
	result, err := StringToIntPtr(getContext(), "42")
	require.NoError(t, err)
	require.Equal(t, *result, int64(42))
}

func TestStringToBoolPtr_Invalid(t *testing.T) {
	result, err := StringToBoolPtr(getContext(), "falte")
	require.Equal(t, "ServerError(Kind=Internal Server Error, Message=invalid boolean format: falte, Cause=%!s(<nil>))", err.Error())
	require.Nil(t, result)
}

func TestStringToBoolPtr_Empty(t *testing.T) {
	result, err := StringToBoolPtr(getContext(), "")
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestStringToTimePtr_Empty(t *testing.T) {
	result, err := StringToTimePtr(getContext(), "")
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestStringToTimePtr_Invalid(t *testing.T) {
	result, err := StringToTimePtr(getContext(), "2012-11-01T22;08:41+00:00")
	require.Equal(t, "ServerError(Kind=Internal Server Error, Message=invalid time format: 2012-11-01T22;08:41+00:00, Cause=parsing time \"2012-11-01T22;08:41+00:00\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \";08:41+00:00\" as \":\")", err.Error())
	require.Nil(t, result)
}

func TestStringToTimePtr_Valid(t *testing.T) {
	result, err := StringToTimePtr(getContext(), "2012-11-01T22:08:41.000+00:00")
	require.NoError(t, err)
	expected, _ := time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")
	require.Equal(t, JSONTime{expected}, *result)
}

func TestStringToTimePtr_ISOValid(t *testing.T) {
	result, err := StringToTimePtr(getContext(), "2012-11-01T22:08:41.000+0000")
	require.NoError(t, err)
	expected, _ := time.Parse("2006-01-02T15:04:05.000-0700", "2012-11-01T22:08:41.000+0000")
	require.Equal(t, JSONTime{expected}, *result)
}

func TestJSONTimeEncode(t *testing.T) {
	var b bytes.Buffer
	bytesWriter := bufio.NewWriter(&b)
	pt, _ := time.Parse("2006-01-02T15:04:05.000-0700", "2012-11-01T22:08:41.000+0000")
	ut := JSONTime{pt}
	_ = json.NewEncoder(bytesWriter).Encode(ut)
	bytesWriter.Flush()
	require.Equal(t, "\"2012-11-01T22:08:41.000+0000\"\n", b.String())
}

func TestJSONTimeDecode(t *testing.T) {
	timeStr := "\"2012-11-01T22:08:41.000+0000\"\n"
	pt, _ := time.Parse("2006-01-02T15:04:05.000-0700", "2012-11-01T22:08:41.000+0000")
	ut := JSONTime{time.Now()}
	_ = json.NewDecoder(strings.NewReader(timeStr)).Decode(&ut)
	require.Equal(t, JSONTime{pt}, ut)
}
