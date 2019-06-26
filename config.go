package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	SentryDsn    string   `yaml:"sentry_dsn"`
	PatternFiles []string `yaml:"pattern_files"`
	Inputs       []struct {
		File           string
		Follow         *bool // Optional value
		HelperPatterns []struct {
			Name  string
			Value string
		} `yaml:"helper_patterns"`
		Patterns []string
		Tags     map[string]string
	}
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
