package common

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
)

type temporalClientMapKeyType int

const temporalClientMapKey temporalClientMapKeyType = iota

func WithTemporalClientMap(ctx context.Context, m map[string]client.Client) context.Context {
	return context.WithValue(ctx, temporalClientMapKey, m)
}

func TemporalClientFrom(ctx context.Context, taskqueue string) (client.Client, error) {
	m, is := ctx.Value(temporalClientMapKey).(map[string]client.Client)
	if !is {
		return nil, fmt.Errorf("no temporal client exists in this handler")
	}
	c, has := m[taskqueue]
	if !has {
		return nil, fmt.Errorf("no temporal client for TaskQueue %q exists in this handler", taskqueue)
	}
	return c, nil
}
