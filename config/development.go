package config

// DevelopmentConfig struct.
type DevelopmentConfig struct {
	// insecureDisableAllAuthorisationRules can be used to disable all authorisation rule logic
	// guarding calls to endpoints or RPC methods, and instead unconditionally grant access.
	// This option is insecure and should not be enabled in production.
	InsecureDisableAllAuthorisationRules bool `yaml:"insecureDisableAllAuthorisationRules"`
}
