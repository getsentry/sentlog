package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/vjeantet/grok"
	"gopkg.in/alecthomas/kingpin.v2"
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

func IsDryRun() bool {
	return *dryRun
}

func InitSentry() {
	if IsDryRun() {
		log.Println("Dry-run mode enabled, not initializing Sentry client")
		return
	}

	dsn := os.Getenv("SENTLOG_SENTRY_DSN")
	if dsn == "" {
		log.Fatal("No DSN found\n")
	}
	err := sentry.Init(sentry.ClientOptions{})

	if err != nil {
		log.Fatal("Sentry initialization failed: %v\n", err)
	}
}

func CaptureEvent(line string, values map[string]string) {
	message := values["err_message"]
	if message == "" {
		message = line
	}

	if IsDryRun() {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range values {
			if value == "" {
				continue
			}
			scope.SetTag(key, value)
		}

		scope.SetLevel(sentry.LevelError)

		scope.SetExtra("log_entry", line)

		sentry.CaptureMessage(message)
	})
}

func ProcessLine(line string, pattern string, g *grok.Grok) {
	values, err := g.Parse(pattern, line)
	if err != nil {
		log.Printf("grok parsing failed: %v\n", err)
		os.Exit(1)
	}

	if !IsDryRun() {
		// Original log line
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Message: line,
			Level:   sentry.LevelInfo,
		})
	}

	if len(values) == 0 {
		return
	}

	CaptureEvent(line, values)

	log.Println(">>> Entry:")
	printMap(values)
}

func ProcessFile(filename string, pattern string) {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		log.Fatalf("grok initialization failed: %v\n", err)
	}
	AddDefaultPatterns(g)

	file, err := os.Open(filename) // For read access.
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()

	log.Printf("Opened \"%s\"", filename)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		ProcessLine(line, pattern, g)
	}

	if !IsDryRun() {
		sentry.Flush(5 * time.Second)
	}
}

var (
	file    = kingpin.Arg("file", "File to parse").Required().String()
	pattern = kingpin.Flag("pattern", "Pattern to look for").Required().String()
	dryRun  = kingpin.Flag("dry-run", "Dry-run mode").Bool()
)

func main() {
	kingpin.Parse()

	InitSentry()
	ProcessFile(*file, *pattern)
}
