package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Logging struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"logging"`
}

type Manager struct {
	v    *viper.Viper
	mu   sync.RWMutex
	cfg  *Config
	subs []chan *Config
}

func NewManager(path string) (*Manager, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	m := &Manager{
		v:   v,
		cfg: &cfg,
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		var nc Config
		if err := v.Unmarshal(&nc); err != nil {
			fmt.Printf("[config] reload failed: %v\n", err)
			return
		}

		m.mu.Lock()
		m.cfg = &nc
		for _, ch := range m.subs {
			select {
			case ch <- &nc:
			default:
			}
		}
		m.mu.Unlock()
		fmt.Printf("[config] reloaded from %s\n", e.Name)
	})

	return m, nil
}

func (m *Manager) Get() *Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

func (m *Manager) Subscribe() <-chan *Config {
	ch := make(chan *Config, 1)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	ch <- m.cfg
	m.mu.Unlock()
	return ch
}
