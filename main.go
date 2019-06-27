package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"gopkg.in/alecthomas/kingpin.v2"
)

type CmdArgs struct {
	file     *string
	pattern  *string
	dryRun   *bool
	noFollow *bool
	verbose  *bool
	// maxEvents      *int
	fromLineNumber *int
	config         *string
}

var (
	_isDryRun bool
	_verbose  bool
)

func isDryRun() bool {
	return _isDryRun
}

func isVerbose() bool {
	return _verbose
}

func initSentry(config *Config) {
	if isDryRun() {
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

// Catches Ctrl-C
func catchInterrupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func() {
		<-c
		log.Println("Cleaning up...")
		sentry.Flush(5 * time.Second)
		os.Exit(1)
	}()
}

func main() {
	args := CmdArgs{
		file:     kingpin.Arg("file", "File to parse").String(),
		pattern:  kingpin.Flag("pattern", "Pattern to look for").String(),
		dryRun:   kingpin.Flag("dry-run", "Dry-run mode").Default("false").Bool(),
		noFollow: kingpin.Flag("no-follow", "Do not wait for the new data").Bool(),
		// maxEvents:      kingpin.Flag("max-events", "Exit after the given number of events are processed").Int(),
		fromLineNumber: kingpin.Flag("from-line", "Start reading from this line number").Default("-1").Int(),
		config:         kingpin.Flag("config", "Path to the configuration").Short('c').String(),
		verbose:        kingpin.Flag("verbose", "Print every match").Short('v').Default("false").Bool(),
	}

	kingpin.Parse()
	_isDryRun = *args.dryRun
	_verbose = *args.verbose

	var config *Config

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

	initSentry(config)
	catchInterrupt()
	runWithConfig(config)
}
