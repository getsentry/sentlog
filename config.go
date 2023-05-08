package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type FileInputConfig struct {
	File           string
	Follow         *bool
	FromLineNumber *int `yaml:"from_line_number"`
	Patterns       []string
	Tags           map[string]string
}

type Config struct {
	SentryDsn    string   `yaml:"sentry_dsn"`
	PatternFiles []string `yaml:"pattern_files"`
	Inputs       []FileInputConfig
}

func ReadConfigFromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := Config{}
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(&config)
	if err != nil {
		return &config, err
	}

	return &config, nil
}
