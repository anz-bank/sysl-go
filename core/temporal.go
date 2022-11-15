package core

import (
	"context"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

type TemporalServiceSpec interface {
	worker.Worker
	client.Client
	Register()
}

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

func ExecuteWorkflow[T any](ctx context.Context, option client.StartWorkflowOptions, c client.Client, tq, name string, args ...any) (*Run[T], error) {
	// TaskQueue name must be App while name is just Endpoint name
	option.TaskQueue = tq
	w, err := c.ExecuteWorkflow(ctx, option, name, args...)
	if err != nil {
		return nil, err
	}
	return &Run[T]{w}, nil
}

// GetOptionFromClientIntf takes in the user provided options which is an array, check for length,
// and return an option.
func GetOptionFromClientIntf(options []client.StartWorkflowOptions) client.StartWorkflowOptions {
	switch len(options) {
	case 1:
		return options[0]
	case 0:
		return client.StartWorkflowOptions{}
	}
	panic("more than one option is defined")
}

type Future[T any] struct {
	workflow.Future
}

func (f *Future[T]) Get(ctx workflow.Context) (T, error) {
	var t T
	// FIXME: should this wait until it's ready?
	err := f.Future.Get(ctx, &t)
	return t, err
}

func ExecuteActivity[T any](ctx workflow.Context, tq, name string, args ...any) *Future[T] {
	ctx = workflow.WithTaskQueue(ctx, tq)
	if workflow.GetActivityOptions(ctx).StartToCloseTimeout == 0 {
		ctx = workflow.WithStartToCloseTimeout(ctx, 5*time.Second)
	}
	return &Future[T]{workflow.ExecuteActivity(ctx, name, args...)}
}
