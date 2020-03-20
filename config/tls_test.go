package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	tls2 "crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TLS
func TestTLSCiphers(t *testing.T) {
	req := require.New(t)

	ciphers := []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"}
	tlsVer := "1.2"
	tlsCfg := &TLSConfig{
		MinVersion: &tlsVer,
		MaxVersion: &tlsVer,
		Ciphers:    ciphers,
	}

	val, err := TLSCiphers(tlsCfg)
	req.NoError(err)

	expected := []uint16{
		tls2.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls2.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}
	req.Equal(expected, val)
}

func TestTLSCiphersFail(t *testing.T) {
	req := require.New(t)

	ciphers := []string{"IM_NOT_A_CIPHER_SUITE"}
	tlsVer := "1.2"
	tlsCfg := &TLSConfig{
		MinVersion: &tlsVer,
		MaxVersion: &tlsVer,
		Ciphers:    ciphers,
	}

	val, err := TLSCiphers(tlsCfg)
	req.Error(err)

	var expected []uint16
	req.Equal(expected, val)
}

func TestTLSVersions(t *testing.T) {
	req := require.New(t)

	var ciphers []string
	tlsVer := "1.2"
	tlsCfg := &TLSConfig{
		MinVersion: &tlsVer,
		MaxVersion: &tlsVer,
		Ciphers:    ciphers,
	}

	min, max, err := TLSVersions(tlsCfg)
	req.NoError(err)

	req.Equal(min, uint16(tls2.VersionTLS12))
	req.Equal(max, uint16(tls2.VersionTLS12))
}

func TestTLSVersionsFail(t *testing.T) {
	req := require.New(t)

	var ciphers []string
	badVersion := "1.4"
	tlsCfg := &TLSConfig{
		MinVersion: &badVersion,
		MaxVersion: &badVersion,
		Ciphers:    ciphers,
	}

	_, _, err := TLSVersions(tlsCfg)
	req.Error(err)

	tlsCfg.MinVersion = NewString("1.1")
	_, _, err = TLSVersions(tlsCfg)
	req.Error(err)
}

var tlsConfigSetupTests = []struct {
	in   TLSConfig
	out  error
	name string
}{
	{TLSConfig{
		NewString("1.4"),
		NewString("1.4"),
		nil,
		[]string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		nil,
		nil,
		false,
	},
		fmt.Errorf("invalid TLSMin config: 1.4"), "TEST: tlsConfigSetupTests #1"},
	{TLSConfig{
		NewString("1.2"),
		NewString("1.2"),
		NewString("this_is_not_a_valid_policy"),
		[]string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		nil,
		nil,
		false,
	},
		fmt.Errorf("invalid client authentication policy: this_is_not_a_valid_policy"), "TEST: tlsConfigSetupTests #2"},
	{TLSConfig{
		NewString("1.2"),
		NewString("1.2"),
		NewString("this_is_not_a_valid_policy"),
		[]string{"T", "L", "S", "E", "C", "D", "H", "E", "E", "C", "D", "S", "A", "W",
			"I", "T", "H", "A", "E", "S", "L", "S", "T", "L", "S", "E", "C", "D", "H", "E"},
		nil,
		nil,
		false,
	}, fmt.Errorf("TLS cipher suite configuration contains more ciphers than the number of known ciphers"), "TEST: tlsConfigSetupTests #3"},
	{TLSConfig{
		NewString("1.3"),
		NewString("1.2"),
		nil,
		[]string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"},
		nil,
		nil,
		false,
	},
		fmt.Errorf("invalid TLS version config"), "TEST: tlsConfigSetupTests #4"},
}

func TestConfigureTLSInvalidConfig(t *testing.T) {
	for _, tt := range tlsConfigSetupTests {
		_, err := MakeTLSConfig(&tt.in)
		assert.Error(t, err, tt.name)
		assert.Equal(t, tt.out, err, tt.name)
	}
}

func TestConfigureTLS(t *testing.T) {
	req := require.New(t)

	dir, err := ioutil.TempDir("", "TestConfigureTLS")
	req.NoError(err, "error during test setup: failed to create temp dir")
	defer func() {
		err = os.RemoveAll(dir)
		require.NoError(t, err, "warning: failed to remove temp dir: %+v", err)
	}()

	certFilename := filepath.Join(dir, "cert.pem")
	keyFilename := filepath.Join(dir, "key.pem")

	err = generateSelfSignedCert([]string{""}, "banana.example.com", certFilename, keyFilename)
	req.NoError(err, "failed to generate cert & key for test scenario")

	expectedIdentityCert, err := tls2.LoadX509KeyPair(certFilename, keyFilename)
	req.NoError(err)

	identity := ServerIdentityConfig{
		CertKeyPair: &CertKeyPair{
			CertPath: &certFilename,
			KeyPath:  &keyFilename,
		},
	}
	cfg := NewTLSConfig("1.2", "1.2", "RequireAndVerifyClientCert",
		[]string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256"}, identity)

	tlsCfg, err := MakeTLSConfig(cfg)
	req.NoError(err)

	tempCAs, err := GetTrustedCAs(cfg)
	req.NoError(err)

	expectedTLS := &tls2.Config{
		MinVersion:               tls2.VersionTLS12,
		MaxVersion:               tls2.VersionTLS12,
		CipherSuites:             []uint16{tls2.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, tls2.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
		Certificates:             []tls2.Certificate{expectedIdentityCert},
		ClientCAs:                tempCAs,
		ClientAuth:               tls2.RequireAndVerifyClientCert,
		RootCAs:                  tempCAs,
	}

	req.Equal(expectedTLS, tlsCfg)
}

func TestTLSClientAuth(t *testing.T) {
	req := require.New(t)

	var ciphers []string
	tlsMin := "1.3"
	tlsMax := "1.3"
	clientAuth := "RequireAndVerifyClientCert"
	cfg := TLSConfig{
		MinVersion: &tlsMin,
		MaxVersion: &tlsMax,
		ClientAuth: &clientAuth,
		Ciphers:    ciphers,
	}

	expectedPolicy := tls2.RequireAndVerifyClientCert
	res, err := TLSClientAuth(&cfg)
	req.NoError(err)
	req.Equal(&expectedPolicy, res)
}

func TestTLSClientAuthFail(t *testing.T) {
	req := require.New(t)

	var ciphers []string
	tlsMin := "1.3"
	tlsMax := "1.3"
	ClientAuth := "NON_EXISTENT_POLICY"
	cfg := TLSConfig{
		ClientAuth: &ClientAuth,
		MinVersion: &tlsMin,
		MaxVersion: &tlsMax,
		Ciphers:    ciphers,
	}

	var expected *tls2.ClientAuthType
	res, err := TLSClientAuth(&cfg)
	req.Error(err)
	req.Equal(expected, res)
}

var tlsInvalidConfigTests = []struct {
	in   TLSConfig
	out  error
	name string
}{
	{TLSConfig{}, fmt.Errorf("config missing"), "TEST: tlsInvalidConfigTests #1"},
	{TLSConfig{MinVersion: NewString("1.2")}, fmt.Errorf("clientAuth config missing"), "TEST: tlsInvalidConfigTests #2"},
	{TLSConfig{
		Ciphers:    []string{},
		MinVersion: NewString(""),
		MaxVersion: NewString(""),
		ClientAuth: NewString(""),
	}, fmt.Errorf("clientAuth: client authentication policy must be set if TLS is in use"), "TEST: tlsInvalidConfigTests #3"},
	{TLSConfig{
		Ciphers:    []string{"TLS_BANANA_RAMA"},
		MinVersion: NewString("1.2"),
		MaxVersion: NewString("1.2"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
	}, fmt.Errorf("ciphers: [TLS_BANANA_RAMA] are not valid"), "TEST: tlsInvalidConfigTests #4"},
	{TLSConfig{
		Ciphers:    []string{},
		MinVersion: NewString("1.2"),
		MaxVersion: nil,
		ClientAuth: NewString("RequireAndVerifyClientCert"),
	}, fmt.Errorf("max config missing"), "TEST: tlsInvalidConfigTests #5"},
	{TLSConfig{
		Ciphers:    []string{},
		MinVersion: NewString("1.5"),
		MaxVersion: NewString("1.2"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
	}, fmt.Errorf("min: TLS version not recognized"), "TEST: tlsInvalidConfigTests #6"},
	{TLSConfig{
		Ciphers:    []string{},
		MinVersion: NewString("1.2"),
		MaxVersion: NewString("1.5"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
	}, fmt.Errorf("max: TLS version not recognized"), "TEST: tlsInvalidConfigTests #7"},
}

func TestValidateInvalidTlsConfigs(t *testing.T) {
	for _, tt := range tlsInvalidConfigTests {
		err := tt.in.Validate()
		assert.Error(t, err, tt.name)
		assert.Equal(t, tt.out, err, tt.name)
	}
}

// *** helper functions ***

func makeX509Template(organisation string) (*x509.Certificate, error) {
	notBefore := time.Now()

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{organisation},
		},
		NotBefore: notBefore,
		NotAfter:  notBefore.Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}, nil
}

//nolint:funlen
func generateSelfSignedCert(hosts []string, organisation string, certFilename, keyFilename string) error {
	pemBlockForKey := func(priv *ecdsa.PrivateKey) (*pem.Block, error) {
		b, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return nil, err
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	template, err := makeX509Template(organisation)
	if err != nil {
		return err
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	isCA := true
	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFilename)
	if err != nil {
		return err
	}
	if err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	if err = certOut.Close(); err != nil {
		return err
	}

	keyOut, err := os.OpenFile(keyFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	block, err := pemBlockForKey(priv)
	if err != nil {
		return err
	}
	if err = pem.Encode(keyOut, block); err != nil {
		return err
	}
	if err = keyOut.Close(); err != nil {
		return err
	}
	return nil
}

func TestGetTrustedCAsFromPEMByDir(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestGetTrustedCAsByDir")
	require.NoError(t, err, "error during test setup: failed to create temp dir")
	defer func() {
		err = os.RemoveAll(tmpDir)
		require.NoError(t, err, "warning: failed to remove temp dir: %+v", err)
	}()

	cfg := &TLSConfig{
		Ciphers:    []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		MinVersion: NewString("1.2"),
		MaxVersion: NewString("1.2"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
		TrustedCertPool: &TrustedCertPoolConfig{
			Mode:     NewString("directory"),
			Encoding: NewString("pem"),
			Path:     NewString(tmpDir),
		},
	}

	var pool *x509.CertPool
	certPath := filepath.Join(*cfg.TrustedCertPool.Path, "cert")
	keyPath := filepath.Join(*cfg.TrustedCertPool.Path, "key")
	err = generateSelfSignedCert(nil, "", certPath, keyPath)
	require.NoError(t, err)

	pool, err = GetTrustedCAs(cfg)
	assert.NotNil(t, pool)
	assert.NoError(t, err)
}

var tlsGetTrustedCAsByFileTests = []struct {
	in   *TLSConfig
	name string
}{
	{&TLSConfig{
		Ciphers:    []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		MinVersion: NewString("1.2"),
		MaxVersion: NewString("1.2"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
		TrustedCertPool: &TrustedCertPoolConfig{
			Mode:     NewString("file"),
			Encoding: NewString("pem"),
			Path:     NewString("cert"),
		}}, "TEST: tlsGetTrustedCAsByFileTests #1"},
}

func TestGetTrustedCAsByFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestGetTrustedCAsByFile")
	require.NoError(t, err, "error during test setup: failed to create temp dir")
	defer func() {
		err = os.RemoveAll(tmpDir)
		require.NoError(t, err, "warning: failed to remove temp dir: %+v", err)
	}()
	for _, tt := range tlsGetTrustedCAsByFileTests {
		pathToFile := filepath.Join(tmpDir, *tt.in.TrustedCertPool.Path)
		tt.in.TrustedCertPool.Path = &pathToFile
		var pool *x509.CertPool

		err := generateSelfSignedCert(nil, "", *tt.in.TrustedCertPool.Path, filepath.Join(tmpDir, "key"))
		require.NoError(t, err, tt.name)

		pool, err = GetTrustedCAs(tt.in)
		assert.NotNil(t, pool, tt.name)
		assert.NoError(t, err, tt.name)
	}
}

func TestGetTrustedCAsFromSystem(t *testing.T) {
	cfg := &TLSConfig{
		Ciphers:    []string{"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
		MinVersion: NewString("1.2"),
		MaxVersion: NewString("1.2"),
		ClientAuth: NewString("RequireAndVerifyClientCert"),
		TrustedCertPool: &TrustedCertPoolConfig{
			Mode: NewString("system"),
		},
	}
	res, err := GetTrustedCAs(cfg)
	assert.NoError(t, err)
	if runtime.GOOS == "windows" {
		assert.Nil(t, res)
	} else {
		assert.NotNil(t, res)
	}
}

var tlsValidTLSTrustedCertPoolConfigTests = []struct {
	in   *TrustedCertPoolConfig
	name string
}{
	{&TrustedCertPoolConfig{
		Mode:     NewString("directory"),
		Encoding: NewString("PEM"),
		Path:     NewString(""),
	}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #1"},
	{&TrustedCertPoolConfig{
		Mode:     NewString("File"),
		Encoding: NewString("PEM"),
		Path:     NewString(""),
	}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #2"},
}

var tlsInvalidTLSTrustedCertPoolConfigTests = []struct {
	in   *TrustedCertPoolConfig
	name string
}{
	{&TrustedCertPoolConfig{}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #1"},
	{&TrustedCertPoolConfig{
		Encoding: NewString("PEM"),
		Path:     NewString(""),
	}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #2"},
	{&TrustedCertPoolConfig{
		Mode:     NewString("directory"),
		Encoding: NewString("UNKNOWN_ENCODING"),
		Path:     NewString(""),
	}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #4"},
	{&TrustedCertPoolConfig{
		Mode: NewString("directory"),
		Path: NewString(""),
	}, "TEST: tlsValidTlsTrustedCertPoolConfigTests #5"},
}

func TestInvalidateTlsTrustedCertPoolConfig(t *testing.T) {
	for _, tt := range tlsInvalidTLSTrustedCertPoolConfigTests {
		err := tt.in.validate()
		assert.Error(t, err, tt.name)
	}
}

func TestInvalidBuildPoolEncodingTypes(t *testing.T) {
	cfg := &TrustedCertPoolConfig{
		Mode:     NewString("directory"),
		Encoding: NewString("UNKNOWN_CERT_ENCODING"),
		Path:     NewString("."),
	}
	res, err := buildPool(cfg)
	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestInsecureSkipVerify(t *testing.T) {
	cfg := &TLSConfig{
		InsecureSkipVerify: true,
	}

	tlsConfig, err := MakeTLSConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, true, tlsConfig.InsecureSkipVerify)
}
