package main

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	"gopkg.in/alecthomas/kingpin.v2"
)

type CmdArgs struct {
	file           *string
	pattern        *string
	dryRun         *bool
	noFollow       *bool
	maxEvents      *int
	fromLineNumber *int
}

var (
	args = CmdArgs{
		file:           kingpin.Arg("file", "File to parse").Required().String(),
		pattern:        kingpin.Flag("pattern", "Pattern to look for").Required().String(),
		dryRun:         kingpin.Flag("dry-run", "Dry-run mode").Bool(),
		noFollow:       kingpin.Flag("no-follow", "Do not wait for the new data").Bool(),
		maxEvents:      kingpin.Flag("max-events", "Exit after the given number of events are processed").Int(),
		fromLineNumber: kingpin.Flag("from-line", "Start reading from this line number").Default("-1").Int(),
	}
)

func IsDryRun() bool {
	return *args.dryRun
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
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}
}

func main() {
	kingpin.Parse()

	InitSentry()
	ProcessFile(*args.file, *args.pattern)
}
