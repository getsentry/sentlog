package main

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestProcessLine(t *testing.T) {
	_verbose = true
	err := sentry.Init(sentry.ClientOptions{Debug: true})
	if err != nil {
		t.Fatal(err)
	}

	g := initGrokProcessor()

	processLine(
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
		[]string{"%{COMMONAPACHELOG}"},
		g,
		sentry.CurrentHub(),
	)

	processLine(
		"",
		[]string{"%{COMMONAPACHELOG}"},
		g,
		sentry.CurrentHub(),
	)
}
