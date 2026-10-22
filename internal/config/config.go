package config

import "github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"

type Config struct {
	UpdateInterval int
	OutputPath     string
}

func NewConfig(updateInterval int, outputPath string) *Config {
	if updateInterval <= 0 {
		updateInterval = consts.DefaultLabelUpdateIntervalSeconds
	}
	if outputPath == "" {
		outputPath = consts.DefaultOutputPath
	}

	return &Config{
		UpdateInterval: updateInterval,
		OutputPath:     outputPath,
	}
}
