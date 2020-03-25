package config

import (
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func NewTLSConfig(tlsMin, tlsMax string, clientAuth string, ciphers []string, identity ServerIdentityConfig) *TLSConfig {
	return &TLSConfig{
		MinVersion:     &tlsMin,
		MaxVersion:     &tlsMax,
		ClientAuth:     &clientAuth,
		Ciphers:        ciphers,
		ServerIdentity: &identity,
	}
}

func defaultAdminServer() CommonHTTPServerConfig {
	return CommonHTTPServerConfig{
		Common: CommonServerConfig{
			HostName: "admin host",
			Port:     3333,
			TLS:      nil,
		},
		BasePath:     "/admin",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

func defaultPublicServer() CommonHTTPServerConfig {
	return CommonHTTPServerConfig{
		Common: CommonServerConfig{
			HostName: "public host",
			Port:     3000,
			TLS:      nil,
		},
		BasePath:     "/public",
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

func TestValidateGlobalConfigLoPort(t *testing.T) {
	config := defaultAdminServer()
	config.Common.Port = -1
	err := config.Validate()
	require.Error(t, err)
	ErrorDueToFields(t, err, "Port")
}

func TestValidateGlobalConfigHiPort(t *testing.T) {
	config := defaultAdminServer()
	config.Common.Port = 65535
	err := config.Validate()
	require.Error(t, err)
	ErrorDueToFields(t, err, "Port")
}

func TestValidateGlobalConfigBadBasePath(t *testing.T) {
	// regard missing base path as "/" base path
	config := defaultAdminServer()
	config.BasePath = "basepath"
	err := config.Validate()
	require.Error(t, err)
	ErrorDueToFields(t, err, "BasePath")
}

func TestValidateGlobalConfigEmptyBasePath(t *testing.T) {
	config := defaultAdminServer()
	config.BasePath = ""
	err := config.Validate()
	require.Error(t, err)
	ErrorDueToFields(t, err, "BasePath")
}

func TestValidateGlobalConfigSlashBasePath(t *testing.T) {
	config := defaultAdminServer()
	config.BasePath = "/"
	err := config.Validate()
	require.NoError(t, err)
}

func TestProxyHandlerFromConfig(t *testing.T) {
	dummyReq, _ := http.NewRequest("", "", nil)
	testTransport := Transport{
		ProxyURL: "https://localhost:3128",
		UseProxy: true,
	}
	testURL, _ := url.Parse(testTransport.ProxyURL)
	fn := proxyHandlerFromConfig(&testTransport)
	requestURL, err := fn(dummyReq)
	require.NoError(t, err)
	require.Equal(t, requestURL, testURL)
}

func TestProxyHandlerFromConfigDefaultProxy(t *testing.T) {
	os.Setenv(`http_proxy`, `http://localhost:3128`)
	os.Setenv(`https_proxy`, `http://localhost:3128`)
	dummyReq, _ := http.NewRequest("", "", nil)
	testTransport := Transport{
		UseProxy: true,
	}
	fn := proxyHandlerFromConfig(&testTransport)
	requestURL, err := fn(dummyReq)
	require.NoError(t, err)
	require.Equal(t, `http://localhost:3128`, requestURL.String())
}

func TestProxyHandlerFromConfigNoProxy(t *testing.T) {
	testTransport := Transport{}
	fn := proxyHandlerFromConfig(&testTransport)
	require.Nil(t, fn)
}
