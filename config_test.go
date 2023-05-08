package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfigFromFile(t *testing.T) {
	temporaryDirectory, err := os.MkdirTemp(os.TempDir(), "sentlog-*")
	if err != nil {
		t.Fatalf("creating temporary directory: %v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(temporaryDirectory)
		if err != nil {
			t.Log(err)
		}
	})

	// Write temporary file
	payload := []byte(`---
# Sentry DSN (also can be configured via environment)
sentry_dsn: https://XXX@sentry.io/YYY
# Additional Grok pattern files
pattern_files:
  - ./patterns1.txt
  - ../patterns2.txt

# List of files that we want to watch
inputs:
  - file: /var/log/nginx/error.log
    # Patterns to find and report
    patterns:
      - "%{NGINX_ERROR_LOG}"
    # Additional tags that will be added to the Sentry event
    tags:
      pattern: nginx_error
      custom: tag`)
	fileName := path.Join(temporaryDirectory, "config.yml")
	err = os.WriteFile(fileName, payload, 0644)
	if err != nil {
		t.Fatalf("writing temporary file: %v", err)
	}

	config, err := ReadConfigFromFile(fileName)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, &Config{
		SentryDsn:    "https://XXX@sentry.io/YYY",
		PatternFiles: []string{"./patterns1.txt", "../patterns2.txt"},
		Inputs: []FileInputConfig{
			{
				File:           "/var/log/nginx/error.log",
				Follow:         nil,
				FromLineNumber: nil,
				Patterns:       []string{"%{NGINX_ERROR_LOG}"},
				Tags: map[string]string{
					"pattern": "nginx_error",
					"custom":  "tag",
				},
			},
		},
	}, config)
}
