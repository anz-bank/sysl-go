package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServiceInterface struct{}

type TestAppConfig struct{}

func TestNewServerReturnsErrorIfNewManagerReturnsError(t *testing.T) {
	args := os.Args
	defer func() { os.Args = args }()
	os.Args = append(os.Args[:1], "config.yaml")

	memFs := afero.NewMemMapFs()
	err := afero.Afero{Fs: memFs}.WriteFile("config.yaml", []byte(""), 0777)
	require.NoError(t, err)
	srv, err := NewServer(
		ConfigFileSystemOnto(context.Background(), memFs),
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
