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

type TestAppConfig struct {
	testFn func()
	field1 int `mapstructure:"field1"`
	field2 int `yaml:"field2"`
	Field3 int
}

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

func TestDescribeYAMLForTypeContainsFuncs(t *testing.T) {
	t.Parallel()

	w := bytes.Buffer{}
	describeYAMLForType(&w, reflect.TypeOf(TestAppConfig{
		testFn: func() {},
		field1: 0,
		field2: 1,
	}), map[reflect.Type]string{}, 0)
	assert.Equal(t, "\nfield1: \x1b[1m0\x1b[0m\nfield2: \x1b[1m0\x1b[0m\nField3: \x1b[1m0\x1b[0m",
		w.String())
}
