package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
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

	data, err := ioutil.ReadAll(file)
	config := Config{}
	err = yaml.UnmarshalStrict([]byte(data), &config)

	if err != nil {
		return &config, err
	}

	return &config, nil
}
