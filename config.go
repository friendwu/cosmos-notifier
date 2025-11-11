package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel      string      `yaml:"log_level"`
	FetchInterval string      `yaml:"fetch_interval"`
	Slack         SlackConfig `yaml:"slack"`
	Chains        []ChainConfig `yaml:"chains"`
}

type SlackConfig struct {
	WebhookURL string `yaml:"webhook_url"`
}

type ChainConfig struct {
	ChainID       string `yaml:"chain_id"`
	Endpoint      string `yaml:"endpoint"`
	ExplorerURL   string `yaml:"explorer_url,omitempty"`
}

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "warn"
	}
	if config.FetchInterval == "" {
		config.FetchInterval = "1m"
	}

	// Validate config
	if config.Slack.WebhookURL == "" {
		return nil, fmt.Errorf("slack.webhook_url is required")
	}
	if len(config.Chains) == 0 {
		return nil, fmt.Errorf("at least one chain must be configured")
	}

	for i, chain := range config.Chains {
		if chain.ChainID == "" {
			return nil, fmt.Errorf("chain[%d].chain_id is required", i)
		}
		if chain.Endpoint == "" {
			return nil, fmt.Errorf("chain[%d].endpoint is required", i)
		}
	}

	return &config, nil
}

func setLogLevelFromConfig(level string) {
	parsedLevel, err := log.ParseLevel(strings.ToLower(level))
	if err != nil {
		log.Warnf("Invalid log_level '%s', defaulting to warning", level)
		log.SetLevel(log.WarnLevel)
		return
	}

	log.SetLevel(parsedLevel)
	log.Infof("Logging level set to %s", parsedLevel)
}

func parseInterval(intervalStr string) time.Duration {
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Errorf("Error parsing fetch_interval: %s, defaulting to 1m", err)
		return time.Minute
	}
	return interval
}

