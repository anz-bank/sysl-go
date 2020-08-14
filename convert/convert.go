package convert

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anz-bank/sysl-go/common"
)

type JSONTime struct {
	time.Time
}

func (i JSONTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", i.Format("2006-01-02T15:04:05.000-0700"))
	return []byte(stamp), nil
}

func (i *JSONTime) UnmarshalJSON(data []byte) error {
	var t time.Time
	var err error
	str := strings.ReplaceAll((string(data)), "\"", "")
	if t, err = time.Parse("2006-01-02T15:04:05.000-0700", str); err != nil {
		if t, err = time.Parse(time.RFC3339, string(data)); err != nil {
			return err
		}
	}
	*i = JSONTime{t}
	return nil
}

// StringToIntPtr takes a string and converts it to an integer pointer.
func StringToIntPtr(ctx context.Context, input string) (*int64, error) {
	if input == "" {
		return nil, nil
	}

	result, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return nil, common.CreateError(ctx, common.InternalError, "invalid integer format: "+input, nil)
	}

	return &result, nil
}

// StringToBoolPtr takes a string and converts it to a bool pointer.
func StringToBoolPtr(ctx context.Context, input string) (*bool, error) {
	if input == "" {
		return nil, nil
	}

	result := strings.EqualFold(input, "true")
	if !result {
		result = strings.EqualFold(input, "false")
		if !result {
			return nil, common.CreateError(ctx, common.InternalError, "invalid boolean format: "+input, nil)
		}
		result = !result
	}
	return &result, nil
}

// StringToStringPtr takes a string and converts it to a string pointer.
func StringToStringPtr(ctx context.Context, input string) (*string, error) {
	if input == "" {
		return nil, nil
	}
	result := input

	return &result, nil
}

// StringToTimePtr takes a string and converts it to a time.Time pointer.
func StringToTimePtr(ctx context.Context, input string) (*JSONTime, error) {
	if input == "" {
		return nil, nil
	}

	result, err := time.Parse("2006-01-02T15:04:05.000-0700", input)
	if err != nil {
		if result, err = time.Parse(time.RFC3339, input); err != nil {
			return nil, common.CreateError(ctx, common.InternalError, "invalid time format: "+input, err)
		}
	}
	var jsonTime = JSONTime{result}
	return &jsonTime, nil
}
