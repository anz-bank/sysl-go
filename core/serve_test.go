package core

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/anz-bank/sysl-go/config"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServiceInterface struct{}

type TestAppConfig struct{}

func TestServe(t *testing.T) {
	args := os.Args
	defer func() { os.Args = args }()
	os.Args = append(os.Args[:1], "config.yaml")

	memFs := afero.NewMemMapFs()
	err := afero.Afero{Fs: memFs}.WriteFile("config.yaml", []byte(""), 0777)
	require.NoError(t, err)
	assert.Error(t, Serve(
		ConfigFileSystemOnto(context.Background(), memFs),
		struct{}{},
		func(ctx context.Context, config TestAppConfig) (*TestServiceInterface, error) {
			return &TestServiceInterface{}, nil
		},
		&TestServiceInterface{},
		func(cfg *config.DefaultConfig, serviceIntf interface{}) (interface{}, error) {
			return nil, fmt.Errorf("not happening")
		},
	))
}
