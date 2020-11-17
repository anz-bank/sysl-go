package core

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestServiceInterface struct{}

type TestAppConfig struct{}

func TestNewServerReturnsErrorIfNewManagerReturnsError(t *testing.T) {
	// Override sysl-go app command line interface to directly pass in app config
	ctx := WithConfigFile(context.Background(), []byte(""))

	srv, err := NewServer(
		ctx,
		struct{}{},
		func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, *Hooks, error) {
			return &TestServiceInterface{}, nil, nil
		},
		&TestServiceInterface{},
		func(ctx context.Context, cfg *config.DefaultConfig, serviceIntf interface{}, _ *Hooks) (Manager, *GrpcServerManager, error) {
			return nil, nil, fmt.Errorf("not happening")
		},
	)
	assert.Nil(t, srv)
	assert.Error(t, err)
}

func TestDescribeYAMLForType(t *testing.T) {
	t.Parallel()

	// for logrus.Level
	w := bytes.Buffer{}
	describeYAMLForType(&w, reflect.TypeOf(logrus.Level(0)), map[reflect.Type]string{}, 0)
	assert.Equal(t, " \x1b[1minfo\x1b[0m", w.String())
}
