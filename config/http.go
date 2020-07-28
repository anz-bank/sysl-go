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
	}
}

// CommonDownstreamData collects all the client http configuration.
type CommonDownstreamData struct {
	ServiceURL      string        `yaml:"serviceURL"`
	ClientTransport Transport     `yaml:"clientTransport"`
	ClientTimeout   time.Duration `yaml:"clientTimeout" validate:"timeout=1ms:60s"`
}

// Transport is used to initialise DefaultHTTPTransport.
type Transport struct {
	Dialer                Dialer `yaml:"dialer"`
	MaxIdleConns          int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	ClientTLS             *TLSConfig `yaml:"tls"`
	ProxyURL              string     `yaml:"proxyURL"`
	UseProxy              bool       `yaml:"useProxy"`
}

// Dialer is part of the Transport struct.
type Dialer struct {
	Timeout   time.Duration `yaml:"timeout"`
	KeepAlive time.Duration `yaml:"keepAlive"`
	DualStack bool          `yaml:"dualStack"`
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
	HostName string     `yaml:"hostName"`
	Port     int        `yaml:"port" validate:"min=0,max=65534"`
	TLS      *TLSConfig `yaml:"tls"`
}

type CommonHTTPServerConfig struct {
	Common       CommonServerConfig `yaml:"common"`
	BasePath     string             `yaml:"basePath" validate:"startswith=/"`
	ReadTimeout  time.Duration      `yaml:"readTimeout" validate:"nonnil"`
	WriteTimeout time.Duration      `yaml:"writeTimeout" validate:"nonnil"`
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
