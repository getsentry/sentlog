package main

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	"gopkg.in/alecthomas/kingpin.v2"
)

type CmdArgs struct {
	file     *string
	pattern  *string
	dryRun   *bool
	noFollow *bool
	// maxEvents      *int
	fromLineNumber *int
	config         *string
}

var (
	args = CmdArgs{
		file:     kingpin.Arg("file", "File to parse").String(),
		pattern:  kingpin.Flag("pattern", "Pattern to look for").String(),
		dryRun:   kingpin.Flag("dry-run", "Dry-run mode").Bool(),
		noFollow: kingpin.Flag("no-follow", "Do not wait for the new data").Bool(),
		// maxEvents:      kingpin.Flag("max-events", "Exit after the given number of events are processed").Int(),
		fromLineNumber: kingpin.Flag("from-line", "Start reading from this line number").Default("-1").Int(),
		config:         kingpin.Flag("config", "Path to the configuration").String(),
	}
)

func IsDryRun() bool {
	return *args.dryRun
}

func InitSentry(config *Config) {
	if IsDryRun() {
		log.Println("Dry-run mode enabled, not initializing Sentry client")
		return
	}

	dsn := ""

	if config.SentryDsn != "" {
		dsn = config.SentryDsn
	} else {
		dsn = os.Getenv("SENTLOG_SENTRY_DSN")
	}

	if dsn == "" {
		log.Fatal("No DSN found\n")
	}

	err := sentry.Init(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		log.Fatalf("Sentry initialization failed: %v\n", err)
	}
}

func main() {
	kingpin.Parse()

	config := &Config{}

	if *args.config == "" {
		log.Println("Using parameters from the command line")
		follow := !*args.noFollow
		config = &Config{
			SentryDsn: "",
			Inputs: []FileInputConfig{
				FileInputConfig{
					File:           *args.file,
					Follow:         &follow,
					FromLineNumber: args.fromLineNumber,
					Patterns:       []string{*args.pattern},
				},
			},
		}
	} else {
		log.Println("Using parameters from the configuration file")
		if *args.pattern != "" || *args.file != "" {
			log.Fatalln("No pattern/file allowed when configuration file is provided, exiting.")
		}

		configPath := *args.config
		parsedConfig, err := ReadConfigFromFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
		config = parsedConfig
		log.Printf("Configuration file loaded: \"%s\"\n", configPath)
	}

	InitSentry(config)
	RunWithConfig(config)
}
