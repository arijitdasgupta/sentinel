package config

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Target struct {
	URL  string
	Host string
}

type Config struct {
	Interval    time.Duration `yaml:"-"`
	RawInterval string        `yaml:"interval"`
	Timeout     time.Duration `yaml:"-"`
	RawTimeout  string        `yaml:"timeout"`
	MetricsAddr string        `yaml:"metrics_addr"`
	RawTargets  []string      `yaml:"targets"`
	Targets     []Target      `yaml:"-"`
}

func Default() *Config {
	return &Config{
		Interval:    5 * time.Minute,
		Timeout:     10 * time.Second,
		MetricsAddr: ":9090",
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.RawInterval == "" {
		cfg.RawInterval = "5m"
	}
	d, err := time.ParseDuration(cfg.RawInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval %q: %w", cfg.RawInterval, err)
	}
	cfg.Interval = d

	if cfg.RawTimeout == "" {
		cfg.RawTimeout = "10s"
	}
	d, err = time.ParseDuration(cfg.RawTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout %q: %w", cfg.RawTimeout, err)
	}
	cfg.Timeout = d

	if cfg.MetricsAddr == "" {
		cfg.MetricsAddr = ":9090"
	}

	for i, raw := range cfg.RawTargets {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("target %d: invalid url %q: %w", i, raw, err)
		}
		if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("target %d: url must include scheme and host: %q", i, raw)
		}
		cfg.Targets = append(cfg.Targets, Target{
			URL:  raw,
			Host: u.Hostname(),
		})
	}

	return &cfg, nil
}
