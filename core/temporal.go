package core

import (
	"context"

	"go.temporal.io/sdk/client"
)

type Run[T any] struct {
	client.WorkflowRun
}

func (r *Run[T]) Get(ctx context.Context) (T, error) {
	var t T
	err := r.WorkflowRun.Get(ctx, &t)
	return t, err
}

func (r *Run[T]) GetWithOptions(ctx context.Context, options client.WorkflowRunGetOptions) (T, error) {
	var t T
	err := r.WorkflowRun.GetWithOptions(ctx, &t, options)
	return t, err
}

func ExecuteWorkflow[T any](ctx context.Context, c client.Client, tq, name string, args ...any) (*Run[T], error) {
	// TaskQueue name must be App-Endpoint while name is just Endpoint name
	option := client.StartWorkflowOptions{
		TaskQueue: tq,
	}
	w, err := c.ExecuteWorkflow(ctx, option, name, args...)
	if err != nil {
		return nil, err
	}
	return &Run[T]{w}, nil
}
