package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/vjeantet/grok"
)

func printMap(m map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%+15s: %s\n", k, m[k])

	}
}

func parseNginxEntry(entry string) {
	entries := []string{
		`2019/06/23 12:04:09 [error] 19246#0: *39023663608 client intended to send too large body: 1780877 bytes, client: 123.234.123.234, server: app.getsentry.com, request: "POST /api/1/store/ HTTP/1.1", host: "test.com", referrer: "https://test.com/bla"`,
		`2019/06/24 12:06:01 [error] 19234#0: *39022971737 upstream prematurely closed connection while reading response header from upstream, client: 123.243.123.198, server: sentry.io, request: "POST /api/123/store/ HTTP/1.1", upstream: "http://unix:/var/run/haproxy-api-store.sock:/api/123/store/", host: "sentry.io:443"`}

	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		fmt.Printf("grok initialization failed: %v\n", err)
	}

	g.AddPattern("NGINX_TIMESTAMP", `%{DATE} %{TIME}`)
	g.AddPattern("NGINX_ERROR_LOG", `%{NGINX_TIMESTAMP:timestamp} \[%{DATA:err_severity}\] (%{NUMBER:pid:int}#%{NUMBER}: \*%{NUMBER}|\*%{NUMBER}) %{DATA:err_message}(?:, client: "?%{IPORHOST:client}"?)(?:, server: %{IPORHOST:server})(?:, request: "%{WORD:verb} %{URIPATHPARAM:request} HTTP/%{NUMBER:httpversion}")?(?:, upstream: "%{DATA:upstream}")?(?:, host: "%{URIHOST:host}")?(?:, referrer: "%{URI:referrer}")?`)

	for index, entry := range entries {
		fmt.Printf("\n\n--- Entry number %d\n", index)
		values, err := g.Parse("%{NGINX_ERROR_LOG}", entry)
		if err != nil {
			fmt.Printf("grok parsing failed: %v\n", err)
		}

		if len(values) == 0 {
			fmt.Printf("Pattern %d: matching error!\n", index)
			os.Exit(1)
		}

		printMap(values)

		sentry.WithScope(func(scope *sentry.Scope) {
			for key, value := range values {
				if value == "" {
					continue
				}
				scope.SetTag(key, value)
			}

			// Original log line
			sentry.AddBreadcrumb(&sentry.Breadcrumb{
				Message: entry,
				Level:   sentry.LevelInfo,
			})
			sentry.CaptureMessage(values["err_message"])
		})
		sentry.Flush(5 * time.Second)
	}
}

func captureTest() {
	_, err := os.Open("INVALID")
	if err != nil {
		fmt.Println("Capturing...")
		eventID := sentry.CaptureException(err)
		fmt.Println(*eventID)
	}

	sentry.Flush(time.Second * 5)
}

func main() {
	fmt.Println("starting...")

	err := sentry.Init(sentry.ClientOptions{})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	parseNginxEntry("")
}
