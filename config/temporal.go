package config

type CommonTemporalDownstreamData struct {
	HostPort  string `yaml:"hostPort" mapstructure:"hostPort"`
	Identity  string `yaml:"identity" mapstructure:"identity"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`
}

type TemporalServerConfig struct {
	HostPort  string `yaml:"hostPort" mapstructure:"hostPort"`
	Namespace string `yaml:"namespace" mapstructure:"namespace"`
}
