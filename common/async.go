package common

import (
	"context"
	"fmt"
	"runtime/debug"
)

type AsyncFunc func(ctx context.Context) (interface{}, error)

type AsyncResult struct {
	Obj interface{}
	Err error
}

// A Future is an object that can retrieve a value from an Async call.
type Future interface {
	// Calling future.Get() will block until the Async function either returns a value or panics or the context is done
	Get() (interface{}, error)

	// GetChan will return a channel that the AsyncResult can be read from.
	// Panic and context handling is already handled by Async so there is no need to use a select with a context to cancel
	// the blocking call
	GetChan() <-chan AsyncResult
}

// Async executes the passed in fn in a new go routine and returns a Future which can be called to retrieve the
// result from the go routine.
// This will capture any panic() in the Future as well as manages channels to pass the returned error and result
// back to the caller.
func Async(ctx context.Context, fn AsyncFunc) Future {
	resultChan := make(chan AsyncResult, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- AsyncResult{Obj: nil,
					Err: &ServerError{
						Kind:    InternalError,
						Message: string(debug.Stack()),
						Cause:   fmt.Errorf("%+v", r),
					}}
			}
		}()

		res, err := fn(ctx)
		resultChan <- AsyncResult{
			Obj: res,
			Err: err,
		}
	}()

	return &futureImp{func() (interface{}, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case r := <-resultChan:
			return r.Obj, r.Err
		}
	}}
}

type futureImp struct {
	get func() (interface{}, error)
}

func (future *futureImp) Get() (interface{}, error) {
	return future.get()
}

func (future *futureImp) GetChan() <-chan AsyncResult {
	r := make(chan AsyncResult, 1)

	go func() {
		obj, err := future.get()

		r <- AsyncResult{obj, err}
	}()

	return r
}
