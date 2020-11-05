package config

// DevelopmentConfig struct.
type DevelopmentConfig struct {
	// disableAllAuthorizationRules can be used to disable all authorization rule logic
	// guarding calls to endpoints or RPC methods, and instead unconditionally grant access.
	// This option is insecure and should not be enabled in production.
	DisableAllAuthorizationRules bool `yaml:"disableAllAuthorizationRules" mapstructure:"disableAllAuthorizationRules"`
}
