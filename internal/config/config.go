package config

type ServerDefaults struct {
	Port   int
	UseTLS bool
}

type Config struct {
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
}
