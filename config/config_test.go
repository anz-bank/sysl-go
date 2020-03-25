package config

import (
	"testing"
	"time"

	"github.com/alecthomas/assert"
	"github.com/sirupsen/logrus"
)

type TestMyConfig struct {
	Server TestServer `yaml:"server"`
}

type TestServer struct {
	AdminServer TestAdminServer `yaml:"adminServer"`
}

type TestAdminServer struct {
	ContextTimeout time.Duration `yaml:"contextTimeout"`
	HTTP           TestHTTP      `yaml:"http"`
}

type TestHTTP struct {
	BasePath     string `yaml:"basePath"`
	ReadTimeout  string `yaml:"readTimeout"`
	WriteTimeout string `yaml:"writeTimeout"`
}

type TestDownstreamConfig struct {
	ContextTimeout time.Duration        `yaml:"contextTimeout"`
	Fenergo        CommonDownstreamData `yaml:"fenergo"`
	Qas            CommonDownstreamData `yaml:"qas"`
}

func TestSReadConfig(t *testing.T) {
	t.Parallel()

	defaultConfig := DefaultConfig{
		Library: LibraryConfig{},
		GenCode: GenCodeConfig{
			Downstream: &TestDownstreamConfig{},
		},
	}
	myConfig := TestMyConfig{}
	err := ReadConfig("testdata/config.yaml", &defaultConfig, &myConfig)

	assert.Nil(t, err)

	assert.Equal(t, 2*time.Second, myConfig.Server.AdminServer.ContextTimeout)
	assert.Equal(t, "/admintest", myConfig.Server.AdminServer.HTTP.BasePath)

	assert.True(t, defaultConfig.Library.Log.ReportCaller)
	assert.Equal(t, logrus.WarnLevel, defaultConfig.Library.Log.Level)

	assert.Equal(t, 8080, defaultConfig.GenCode.Upstream.HTTP.Common.Port)
	assert.Equal(t, 8081, defaultConfig.GenCode.Upstream.GRPC.Port)
	assert.Equal(t, 120*time.Second, defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Fenergo.ClientTimeout)
	assert.Equal(t, "https://stubs.bah.apps.x.gcpnp.anz", defaultConfig.GenCode.Downstream.(*TestDownstreamConfig).Qas.ServiceURL)
}
