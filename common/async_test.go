package common

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAsyncSuccess(t *testing.T) {
	fn := func(_ context.Context) (interface{}, error) {
		return 2, nil
	}
	ctx := context.Background()

	future := Async(ctx, fn)
	assert.NotNil(t, future)

	result, err := future.Get()

	assert.NoError(t, err)
	assert.Equal(t, 2, result)
}

func TestAsyncError(t *testing.T) {
	fn := func(_ context.Context) (interface{}, error) {
		return nil, fmt.Errorf("downstream error")
	}
	ctx := context.Background()

	future := Async(ctx, fn)
	assert.NotNil(t, future)

	result, err := future.Get()
	assert.Nil(t, result)

	assert.EqualError(t, err, "downstream error")
}

func TestAsyncPanic(t *testing.T) {
	fn := func(_ context.Context) (interface{}, error) {
		panic("panic error")
	}
	ctx := context.Background()

	future := Async(ctx, fn)
	assert.NotNil(t, future)

	result, err := future.Get()
	assert.Nil(t, result)

	expectedErr := &ServerError{
		Kind:  InternalError,
		Cause: fmt.Errorf("panic error"),
	}
	assert.IsType(t, &ServerError{}, err)
	got := err.(*ServerError)
	assert.Equal(t, expectedErr.Kind, got.Kind)
	assert.Contains(t, got.Message, "common.TestAsyncPanic.func1")
	assert.Equal(t, expectedErr.Cause, got.Cause)
}

func TestAsyncContextTimeout(t *testing.T) {
	fn := func(_ context.Context) (interface{}, error) {
		<-time.After(3 * time.Second)
		return 1, errors.New("hello")
	}
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now())
	defer cancel()

	future := Async(ctx, fn)
	assert.NotNil(t, future)

	result, err := future.Get()
	assert.Nil(t, result)

	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestAsyncGetChan(t *testing.T) {
	fnSuccess := func(_ context.Context) (interface{}, error) {
		return 2, nil
	}
	fnError := func(_ context.Context) (interface{}, error) {
		return nil, fmt.Errorf("downstream error")
	}
	fnPanic := func(_ context.Context) (interface{}, error) {
		panic("panic error")
	}
	fnTimeOut := func(_ context.Context) (interface{}, error) {
		<-time.After(3 * time.Second)
		return 1, errors.New("hello")
	}

	ctx := context.Background()
	ctxWithDeadline, cancel := context.WithDeadline(ctx, time.Now())
	defer cancel()

	futureSuccess := Async(ctx, fnSuccess)
	futureError := Async(ctx, fnError)
	futurePanic := Async(ctx, fnPanic)
	futureTimeOut := Async(ctxWithDeadline, fnTimeOut)
	assert.NotNil(t, futureSuccess)
	assert.NotNil(t, futureError)
	assert.NotNil(t, futurePanic)
	assert.NotNil(t, futureTimeOut)

	resSuccess, resError, resPanic, resTimeOut := <-futureSuccess.GetChan(), <-futureError.GetChan(), <-futurePanic.GetChan(), <-futureTimeOut.GetChan()

	assert.NoError(t, resSuccess.Err)
	assert.Equal(t, 2, resSuccess.Obj)
	assert.Nil(t, resError.Obj)
	assert.EqualError(t, resError.Err, "downstream error")

	assert.Nil(t, resPanic.Obj)
	expectedErr := &ServerError{
		Kind:  InternalError,
		Cause: fmt.Errorf("panic error"),
	}
	assert.IsType(t, &ServerError{}, resPanic.Err)
	got := resPanic.Err.(*ServerError)
	assert.Equal(t, expectedErr.Kind, got.Kind)
	assert.Contains(t, got.Message, "common.TestAsyncGetChan.func3")
	assert.Equal(t, expectedErr.Cause, got.Cause)

	assert.Nil(t, resTimeOut.Obj)
	assert.Equal(t, context.DeadlineExceeded, resTimeOut.Err)
}
