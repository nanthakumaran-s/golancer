package config

import "time"

type ServerDefaults struct {
	Port   int
	UseTLS bool
}

type Proxy struct {
	DefaultTimeout        time.Duration `mapstructure:"default_timeout"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	IdleConnectionTimeout time.Duration `mapstructure:"idle_connection_timeout"`
}

type Match struct {
	Hosts      []string `mapstructure:"hosts"`
	PathPrefix string   `mapstructure:"path_prefix"`
}

type Route struct {
	Name      string   `mapstructure:"name"`
	Match     Match    `mapstructure:"match"`
	Upstreams []string `mapstructure:"upstreams"`
	LB        string   `mapstructure:"lb"`
}

type Config struct {
	Proxy  Proxy   `mapstructure:"proxy"`
	Routes []Route `mapstructure:"routes"`
}
