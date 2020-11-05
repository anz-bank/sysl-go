package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/anz-bank/sysl-go/common"
	"github.com/sirupsen/logrus"
)

var cipherSuites = map[string]uint16{
	"TLS_RSA_WITH_RC4_128_SHA":                tls.TLS_RSA_WITH_RC4_128_SHA,
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA":           tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_RSA_WITH_AES_128_CBC_SHA,
	"TLS_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	"TLS_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":        tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":    tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_RC4_128_SHA":          tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":     tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":      tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":   tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256": tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":   tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384": tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":    tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
	"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":  tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	"TLS_AES_128_GCM_SHA256":                  tls.TLS_AES_128_GCM_SHA256,
	"TLS_AES_256_GCM_SHA384":                  tls.TLS_AES_256_GCM_SHA384,
	"TLS_CHACHA20_POLY1305_SHA256":            tls.TLS_CHACHA20_POLY1305_SHA256,
	"TLS_FALLBACK_SCSV":                       tls.TLS_FALLBACK_SCSV,
}

var tlsVersions = map[string]uint16{
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

var clientAuthTypes = map[string]tls.ClientAuthType{
	"NoClientCert":               tls.NoClientCert,
	"RequestClientCert":          tls.RequestClientCert,
	"RequireAnyClientCert":       tls.RequireAnyClientCert,
	"VerifyClientCertIfGiven":    tls.VerifyClientCertIfGiven,
	"RequireAndVerifyClientCert": tls.RequireAndVerifyClientCert,
}

type TLSConfig struct {
	MinVersion         *string                `yaml:"min" mapstructure:"min"`
	MaxVersion         *string                `yaml:"max" mapstructure:"max"`
	ClientAuth         *string                `yaml:"clientAuth" mapstructure:"clientAuth"`
	Ciphers            []string               `yaml:"ciphers" mapstructure:"ciphers"`
	ServerIdentity     *ServerIdentityConfig  `yaml:"serverIdentity" mapstructure:"serverIdentity"`
	TrustedCertPool    *TrustedCertPoolConfig `yaml:"trustedCertPool" mapstructure:"trustedCertPool"`
	InsecureSkipVerify bool                   `yaml:"insecureSkipVerify" mapstructure:"insecureSkipVerify"`
	SelfSigned         bool                   `yaml:"selfSigned" mapstructure:"selfSigned"`
}

type TrustedCertPoolConfig struct {
	Mode     *string                 `yaml:"mode" mapstructure:"mode"`
	Encoding *string                 `yaml:"encoding" mapstructure:"encoding"`
	Path     *string                 `yaml:"path" mapstructure:"path"`
	Password *common.SensitiveString `yaml:"password" mapstructure:"password"`
}

type ServerIdentityConfig struct {
	CertKeyPair *CertKeyPair `yaml:"certKeyPair" mapstructure:"certKeyPair"`
}

type CertKeyPair struct {
	CertPath *string `yaml:"certPath" mapstructure:"certPath"`
	KeyPath  *string `yaml:"keyPath" mapstructure:"keyPath"`
}

// Cert path modes.
const (
	DIRMODE  = "directory"
	FILEMODE = "file"
	SYSMODE  = "system"
)

// Cert encoding types.
const (
	PEM = "pem"
)

var CertPoolEncodingTypes = map[string]func(cfg *TrustedCertPoolConfig) (pool *x509.CertPool, err error){
	PEM: buildPoolFromPEM,
}

func TLSVersions(cfg *TLSConfig) (min, max uint16, err error) {
	if cfg.MinVersion != nil {
		var has bool
		if min, has = tlsVersions[*cfg.MinVersion]; !has {
			return 0, 0, fmt.Errorf("invalid TLSMin config: %s", *cfg.MinVersion)
		}
	}
	if cfg.MaxVersion != nil {
		var has bool
		if max, has = tlsVersions[*cfg.MaxVersion]; !has {
			return 0, 0, fmt.Errorf("invalid TLSMax config: %s", *cfg.MaxVersion)
		}
	}

	if min > max {
		return 0, 0, fmt.Errorf("invalid TLS version config")
	}

	return min, max, nil
}

func TLSCiphers(cfg *TLSConfig) (ciphers []uint16, err error) {
	tlsCipherCount := len(cipherSuites)
	cfgCipherCount := len(cfg.Ciphers)
	if cfgCipherCount > tlsCipherCount {
		return nil, fmt.Errorf("TLS cipher suite configuration contains more ciphers than the number of known ciphers")
	}

	ciphers = make([]uint16, cfgCipherCount)

	for i, cipher := range cfg.Ciphers {
		c, ok := cipherSuites[cipher]
		if !ok {
			return nil, fmt.Errorf("unknown TLS cipher suite: %s", cipher)
		}
		ciphers[i] = c
	}

	return ciphers, nil
}

func TLSClientAuth(cfg *TLSConfig) (*tls.ClientAuthType, error) {
	policy, ok := clientAuthTypes[*cfg.ClientAuth]
	if !ok {
		return nil, fmt.Errorf("invalid client authentication policy: %s", *cfg.ClientAuth)
	}

	return &policy, nil
}

func OurIdentityCertificates(cfg *TLSConfig) ([]tls.Certificate, error) {
	cs := make([]tls.Certificate, 0)
	if cfg.ServerIdentity.CertKeyPair != nil {
		pair := cfg.ServerIdentity.CertKeyPair
		cert, err := tls.LoadX509KeyPair(*pair.CertPath, *pair.KeyPath)
		if err != nil {
			return nil, err
		}
		cs = append(cs, cert)
		return cs, nil
	}

	return nil, nil
}

func findCertsFromPath(cfg *TrustedCertPoolConfig) ([]string, error) {
	var files []string

	switch strings.ToLower(*cfg.Mode) {
	case DIRMODE:
		var err error
		fileInfo, err := ioutil.ReadDir(*cfg.Path)
		if err != nil {
			return nil, err
		}

		for _, file := range fileInfo {
			if file.IsDir() {
				continue
			}
			files = append(files, filepath.Join(*cfg.Path, file.Name()))
		}
	case FILEMODE:
		files = append(files, *cfg.Path)
	default:
		return nil, fmt.Errorf("unknown or missing mode: %v. Valid modes are %s & %s", *cfg.Mode, DIRMODE, FILEMODE)
	}

	return files, nil
}

func buildPoolFromPEM(cfg *TrustedCertPoolConfig) (*x509.CertPool, error) {
	pool := x509.NewCertPool()

	files, err := findCertsFromPath(cfg)
	if err != nil {
		return nil, err
	}

	var failedCerts []string
	addedCerts := false
	for _, file := range files {
		cert, err := ioutil.ReadFile(file)
		if err != nil {
			failedCerts = append(failedCerts, file)
			continue
		}

		if ok := pool.AppendCertsFromPEM(cert); !ok {
			failedCerts = append(failedCerts, file)
			continue
		}
		addedCerts = true
	}

	if failedCerts != nil {
		if !addedCerts {
			return nil, fmt.Errorf("failed to append any certificates to the RootCA. The following certs failed: %v", failedCerts)
		}
		log.Printf("failed to append the following certs to RootCAs: %v", failedCerts)
	}

	return pool, nil
}

func buildPool(cfg *TrustedCertPoolConfig) (*x509.CertPool, error) {
	var pool *x509.CertPool
	buildPoolFn, ok := CertPoolEncodingTypes[strings.ToLower(*cfg.Encoding)]
	if !ok {
		keys := make([]string, len(CertPoolEncodingTypes))
		for key := range CertPoolEncodingTypes {
			keys = append(keys, key)
		}
		return nil, fmt.Errorf("unrecognised certificate encoding: %s. Valid encodings are: %v in either upper or lower case", *cfg.Encoding, keys)
	}

	pool, err := buildPoolFn(cfg)

	if err != nil {
		return nil, err
	}

	return pool, nil
}

func GetTrustedCAs(cfg *TLSConfig) (*x509.CertPool, error) {
	// certificates exchanged are self signed mostly applicable for dev env
	// skip setting for rootcas, clientcas and cipher suites
	if cfg.TrustedCertPool != nil {
		if *cfg.TrustedCertPool.Mode != SYSMODE {
			pool, err := buildPool(cfg.TrustedCertPool)
			if err != nil {
				return nil, err
			}
			return pool, nil
		}
	}

	if runtime.GOOS == "windows" {
		// crypto/x509: system root pool is not available on Windows
		// hack: try returning nil and see what happens by default.
		return nil, nil
	}
	return x509.SystemCertPool()
}

func makeSelfSignedTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	tlsMin, tlsMax, err := TLSVersions(cfg)
	if err != nil {
		return nil, err
	}
	ourIdentityCertificates, err := OurIdentityCertificates(cfg)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MinVersion:   tlsMin,
		MaxVersion:   tlsMax,
		Certificates: ourIdentityCertificates,
		ClientAuth:   tls.NoClientCert,
	}, nil
}

func MakeTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	if cfg.InsecureSkipVerify {
		//nolint:gosec // This is configured by the user
		return &tls.Config{InsecureSkipVerify: true}, nil
	}

	if cfg.SelfSigned {
		return makeSelfSignedTLSConfig(cfg)
	}

	trustedCAs, err := GetTrustedCAs(cfg)
	if err != nil {
		return nil, err
	}

	tlsMin, tlsMax, err := TLSVersions(cfg)
	if err != nil {
		return nil, err
	}

	ciphers, err := TLSCiphers(cfg)
	if err != nil {
		return nil, err
	}

	policy, err := TLSClientAuth(cfg)
	if err != nil {
		return nil, err
	}

	ourIdentityCertificates, err := OurIdentityCertificates(cfg)
	if err != nil {
		return nil, err
	}

	settings := &tls.Config{
		MinVersion:               tlsMin,
		MaxVersion:               tlsMax,
		CipherSuites:             ciphers,
		PreferServerCipherSuites: true,

		// Certificates must contain one or more certificate chains to present to the
		// other side of the connection, to establish our server's identity.
		// This is intended for use both when we are serving TLS and when we are the
		// client establishing TLS connections with other servers.
		Certificates: ourIdentityCertificates,

		// Upstream (End-user) configuration
		// Certificate authorities to trust when receiving certs from end users (i.e. acting as server)
		ClientCAs:  trustedCAs,
		ClientAuth: *policy,
		// Certificate authorities to trust when receiving certs from other servers (making requests, i.e acting as client)
		RootCAs: trustedCAs,
	}

	return settings, nil
}

//nolint:funlen // TODO: Break this into smaller functions
func (t *TLSConfig) Validate() error {
	if t == nil {
		return nil
	}

	emptyCfg := &TLSConfig{}
	if reflect.DeepEqual(t, emptyCfg) {
		return fmt.Errorf("config missing")
	}

	if t.ClientAuth == nil {
		return fmt.Errorf("clientAuth config missing")
	}

	_, ok := clientAuthTypes[*t.ClientAuth]
	if !ok {
		return fmt.Errorf("clientAuth: client authentication policy must be set if TLS is in use")
	}

	if t.MinVersion == nil {
		return fmt.Errorf("min config missing")
	}

	_, ok = tlsVersions[*t.MinVersion]
	if !ok {
		return fmt.Errorf("min: TLS version not recognized")
	}

	if t.MaxVersion == nil {
		return fmt.Errorf("max config missing")
	}

	_, ok = tlsVersions[*t.MaxVersion]
	if !ok {
		return fmt.Errorf("max: TLS version not recognized")
	}

	if t.Ciphers == nil || len(t.Ciphers) == 0 {
		logrus.Println("ciphers config missing")
	}

	var failedCiphers []string
	for _, cs := range t.Ciphers {
		_, ok := cipherSuites[cs]
		if !ok {
			failedCiphers = append(failedCiphers, cs)
		}
	}
	if len(failedCiphers) > 0 {
		return fmt.Errorf("ciphers: %v are not valid", failedCiphers)
	}

	if t.ServerIdentity == nil {
		return fmt.Errorf("serverIdentity config missing")
	}

	if err := t.ServerIdentity.validate(); err != nil {
		return fmt.Errorf("serverIdentity.%v", err)
	}

	if t.TrustedCertPool == nil {
		return fmt.Errorf("trustedCertPool config missing")
	}

	if err := t.TrustedCertPool.validate(); err != nil {
		return fmt.Errorf("trustedCertPool.%v", err)
	}

	return nil
}

func (b *ServerIdentityConfig) validate() error {
	if b.CertKeyPair != nil {
		if err := b.CertKeyPair.validate(); err != nil {
			return fmt.Errorf("certKeyPair.%v", err)
		}
	}

	return nil
}

func (c *CertKeyPair) validate() error {
	if c.KeyPath == nil {
		return fmt.Errorf("keyPath must be set")
	}

	if c.CertPath == nil {
		return fmt.Errorf("certPath must be set")
	}

	return nil
}

func (t *TrustedCertPoolConfig) validate() error {
	if t.Mode == nil {
		return fmt.Errorf("mode config missing")
	}

	lowMode := strings.ToLower(*t.Mode)
	t.Mode = &lowMode

	if *t.Mode != SYSMODE {
		if t.Path == nil {
			return fmt.Errorf("path config missing")
		}

		if t.Encoding != nil {
			lowEncoding := strings.ToLower(*t.Encoding)
			t.Encoding = &lowEncoding
		} else {
			return fmt.Errorf("encoding missing")
		}

		var validEncodings []string
		for encoding := range CertPoolEncodingTypes {
			validEncodings = append(validEncodings, encoding)
		}

		_, ok := CertPoolEncodingTypes[*t.Encoding]
		if !ok {
			return fmt.Errorf("encoding \"%s\" is not of valid encoding types: %v", *t.Encoding, validEncodings)
		}
	}

	return nil
}
