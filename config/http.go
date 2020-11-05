package config

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/anz-bank/sysl-go/validator"
)

func DefaultCommonDownstreamData() *CommonDownstreamData {
	return &CommonDownstreamData{
		ServiceURL: "",
		ClientTransport: Transport{
			Dialer: Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		ClientTimeout: 60 * time.Second,
		Headers:       make(map[string][]string),
	}
}

// CommonDownstreamData collects all the client http configuration.
type CommonDownstreamData struct {
	ServiceURL      string              `yaml:"serviceURL" mapstructure:"serviceURL"`
	ClientTransport Transport           `yaml:"clientTransport" mapstructure:"clientTransport"`
	ClientTimeout   time.Duration       `yaml:"clientTimeout" mapstructure:"clientTimeout" validate:"timeout=1ms:60s"`
	Headers         map[string][]string `yaml:"headers" mapstructure:"headers"`
}

// Transport is used to initialise DefaultHTTPTransport.
type Transport struct {
	Dialer                Dialer        `yaml:"dialer" mapstructure:"dialer"`
	MaxIdleConns          int           `yaml:"maxIdleConns" mapstructure:"maxIdleConns"`
	IdleConnTimeout       time.Duration `yaml:"idleConnTimeout" mapstructure:"idleConnTimeout"`
	TLSHandshakeTimeout   time.Duration `yaml:"tLSHandshakeTimeout" mapstructure:"tLSHandshakeTimeout"`
	ExpectContinueTimeout time.Duration `yaml:"expectContinueTimeout" mapstructure:"expectContinueTimeout"`
	ClientTLS             *TLSConfig    `yaml:"tls" mapstructure:"tls"`
	ProxyURL              string        `yaml:"proxyURL" mapstructure:"proxyURL"`
	UseProxy              bool          `yaml:"useProxy" mapstructure:"useProxy"`
}

// Dialer is part of the Transport struct.
type Dialer struct {
	Timeout   time.Duration `yaml:"timeout" mapstructure:"timeout"`
	KeepAlive time.Duration `yaml:"keepAlive" mapstructure:"keepAlive"`
	DualStack bool          `yaml:"dualStack" mapstructure:"dualStack"`
}

func (g *CommonDownstreamData) Validate() error {
	if err := validator.Validate(g); err != nil {
		return err
	}

	if g.ClientTransport.ClientTLS != nil {
		if err := g.ClientTransport.ClientTLS.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CommonServerConfig struct {
	HostName string     `yaml:"hostName" mapstructure:"hostName"`
	Port     int        `yaml:"port" mapstructure:"port" validate:"min=0,max=65534"`
	TLS      *TLSConfig `yaml:"tls" mapstructure:"tls"`
}

type CommonHTTPServerConfig struct {
	Common       CommonServerConfig `yaml:"common" mapstructure:"common"`
	BasePath     string             `yaml:"basePath" mapstructure:"basePath" validate:"startswith=/"`
	ReadTimeout  time.Duration      `yaml:"readTimeout" mapstructure:"readTimeout" validate:"nonnil"`
	WriteTimeout time.Duration      `yaml:"writeTimeout" mapstructure:"writeTimeout" validate:"nonnil"`
}

func (c *CommonHTTPServerConfig) Validate() error {
	// existing validation
	if err := validator.Validate(c); err != nil {
		return err
	}

	return nil
}

func proxyHandlerFromConfig(cfg *Transport) func(req *http.Request) (*url.URL, error) {
	if cfg.UseProxy {
		if len(cfg.ProxyURL) > 0 {
			return func(req *http.Request) (*url.URL, error) {
				proxyURL, err := url.Parse(cfg.ProxyURL)
				if err != nil {
					return http.ProxyFromEnvironment(req)
				}
				return proxyURL, err
			}
		}
		return http.ProxyFromEnvironment
	}
	return nil
}

// defaultHTTPTransport returns a new *http.Transport with the same configuration as http.DefaultTransport.
func defaultHTTPTransport(cfg *Transport) (*http.Transport, error) {
	// Finalise the handler loading
	tlsConfig, err := MakeTLSConfig(cfg.ClientTLS)
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		Proxy: proxyHandlerFromConfig(cfg),
		DialContext: (&net.Dialer{
			Timeout:   cfg.Dialer.Timeout,
			KeepAlive: cfg.Dialer.KeepAlive,
			DualStack: cfg.Dialer.DualStack,
		}).DialContext,
		MaxIdleConns:          cfg.MaxIdleConns,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
		TLSClientConfig:       tlsConfig,
	}, nil
}

// DefaultHTTPClient returns a new *http.Client with sensible defaults, in particular it has a timeout set.
func DefaultHTTPClient(cfg *CommonDownstreamData) (*http.Client, error) {
	if cfg == nil {
		cfg = DefaultCommonDownstreamData()
	}

	transport, err := defaultHTTPTransport(&cfg.ClientTransport)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: transport,
	}, nil
}
