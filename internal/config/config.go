package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Owner       string     `yaml:"owner"`
	Repo        string     `yaml:"repo"`
	CachePeriod string     `yaml:"cachePeriod"`
	Artifacts   []Artifact `yaml:"artifacts"`
}

type Artifact struct {
	Name   string `yaml:"name"`
	Regexp string `yaml:"regexp"`
}

// New creates new Config from config data
func New(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// FromFile reads filename and creates Config from it
func FromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg, err := New(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	return cfg, nil
}
