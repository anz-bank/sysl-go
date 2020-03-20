package config

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
)

type TestMyConfig struct {
	Server TestServer `yaml:"server"`
}

type TestServer struct {
	AdminServer TestAdminServer `yaml:"adminServer"`
}

type TestAdminServer struct {
	ContextTimeout time.Duration `yaml:"contextTimeout"`
	Http           TestHttp      `yaml:"http"`
}

type TestHttp struct {
	BasePath     string `yaml:"basePath"`
	ReadTimeout  string `yaml:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout"`
}

func TestSReadConfig(t *testing.T) {
	t.Parallel()

	myConfig := TestMyConfig{}
	lib, gen := ReadConfig("testdata/config.yaml", &myConfig)

	assert.Equal(t, time.Duration(2*time.Second), myConfig.Server.AdminServer.ContextTimeout)
	assert.Equal(t, "/admintest", myConfig.Server.AdminServer.Http.BasePath)

	assert.False(t, lib.Log.ReportCaller)

	assert.Equal(t, 8080, gen.Upstream.HTTP.Common.Port)
	assert.Equal(t, 8081, gen.Upstream.GRPC.Port)
	assert.Equal(t, time.Duration(120*time.Second), gen.Downstream.Fenergo.ClientTimeout)
}
