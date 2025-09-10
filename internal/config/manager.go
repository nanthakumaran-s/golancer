package config

import (
	"fmt"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/nanthakumaran-s/golancer/internal/utils"
	"github.com/spf13/viper"
)

type Manager struct {
	v      *viper.Viper
	mu     sync.RWMutex
	cfg    *Config
	logger *utils.Logger
	subs   []chan *Config
}

func NewManager(lg *utils.Logger) (*Manager, error) {
	config := viper.GetString(utils.CONFIG)
	viper.SetConfigFile(config)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	m := &Manager{
		v:      viper.GetViper(),
		cfg:    &cfg,
		logger: lg,
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		var nc Config
		if err := viper.Unmarshal(&nc); err != nil {
			m.logger.Warn(utils.CONFIG_MANAGER, fmt.Sprintf("reload failed: %v\n", err))
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
		m.logger.Info(utils.CONFIG_MANAGER, fmt.Sprintf("reloaded from %s\n", e.Name))
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
