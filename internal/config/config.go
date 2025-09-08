package config

type ImmutableConfig struct {
	Port   int
	UseTLS bool
}

type MutableConfig struct {
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
}
