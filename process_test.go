package main

import (
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
)

func TestProcessLine(t *testing.T) {
	_verbose = true
	transportMock := &TransportMock{}
	err := sentry.Init(sentry.ClientOptions{Debug: true, Transport: transportMock})
	if err != nil {
		t.Fatal(err)
	}

	g := initGrokProcessor()

	processLine(
		"",
		[]string{"%{COMMONAPACHELOG}"},
		g,
		sentry.CurrentHub().Clone(),
	)
	// We expect it to not send anything to Sentry
	if transportMock.lastEvent != nil {
		t.Errorf("expecting nil, got %v", transportMock.lastEvent)
	}

	processLine(
		`127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] "GET /index.php HTTP/1.1" 404 207`,
		[]string{"%{COMMONAPACHELOG}"},
		g,
		sentry.CurrentHub().Clone(),
	)
	// We expect it to send something to Sentry
	expectMessage := "127.0.0.1 - - [23/Apr/2014:22:58:32 +0200] \"GET /index.php HTTP/1.1\" 404 207"
	if transportMock.lastEvent.Message != expectMessage {
		t.Errorf("expecting lastEvent.Message to be %q, instead got %q", expectMessage, transportMock.lastEvent.Message)
	}

	if transportMock.lastEvent.Level != sentry.LevelError {
		t.Errorf("expecting lastEvent.Level to be %q, instead got %q", sentry.LevelError, transportMock.lastEvent.Level)
	}

	if value, ok := transportMock.lastEvent.Extra["log_entry"]; ok && value != expectMessage {
		t.Errorf("expecting transportMock.lastEvent.Extra[\"log_entry\"] to be %q, instead got %q", expectMessage, value)
	}

	if value, ok := transportMock.lastEvent.Extra["pattern"]; ok && value != "%{COMMONAPACHELOG}" {
		t.Errorf("expecting transportMock.lastEvent.Extra[\"log_entry\"] to be %q, instead got %q", "%{COMMONAPACHELOG}", value)
	}
}

type TransportMock struct {
	mu        sync.Mutex
	events    []*sentry.Event
	lastEvent *sentry.Event
}

func (t *TransportMock) Configure(options sentry.ClientOptions) {}
func (t *TransportMock) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
	t.lastEvent = event
}
func (t *TransportMock) Flush(timeout time.Duration) bool {
	return true
}
func (t *TransportMock) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}
