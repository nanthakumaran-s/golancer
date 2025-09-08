package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Manager struct {
	v    *viper.Viper
	mu   sync.RWMutex
	cfg  *MutableConfig
	subs []chan *MutableConfig
}

func NewManager() (*Manager, error) {
	config := viper.GetString("config")
	viper.SetConfigFile(config)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg MutableConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	m := &Manager{
		v:   viper.GetViper(),
		cfg: &cfg,
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		var nc MutableConfig
		if err := viper.Unmarshal(&nc); err != nil {
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

func (m *Manager) Get() *MutableConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

func (m *Manager) Subscribe() <-chan *MutableConfig {
	ch := make(chan *MutableConfig, 1)
	m.mu.Lock()
	m.subs = append(m.subs, ch)
	ch <- m.cfg
	m.mu.Unlock()
	return ch
}
