package main

import (
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CmdArgs struct {
	file           *string
	pattern        *string
	dryRun         *bool
	noFollow       *bool
	verbose        *bool
	fromLineNumber *int
	config         *string
	debug          *bool
}

var (
	_isDryRun bool
	// _verbose if enabled, will print every debug level log.
	_verbose bool
	// _killed is used to specifying whether the program received
	// a signal to stop and restart the sentlog instance.
	_killed bool
)

func isDryRun() bool {
	return _isDryRun
}

func initSentry(config *Config) {
	if isDryRun() {
		log.Info().Msg("Dry-run mode enabled, not initializing Sentry client")
		return
	}

	dsn := ""

	if config.SentryDsn != "" {
		dsn = config.SentryDsn
	} else {
		dsn = os.Getenv("SENTLOG_SENTRY_DSN")
	}

	err := sentry.Init(sentry.ClientOptions{Dsn: dsn, DebugWriter: log.Logger})
	if err != nil {
		log.Fatal().Err(err).Msg("Sentry initialization failed")
	}
}

// initLogging will initiate logging and convert every call made to log standard library
// to be diverted to zerolog's io.Writer. This function requires _verbose global variable
// to be set beforehand.
func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	logLevelEnv := strings.ToLower(os.Getenv("SENTLOG_LOG_LEVEL"))
	if _verbose {
		logLevelEnv = "debug"
	}

	switch logLevelEnv {
	case "trace":
		log.Logger = log.Sample(zerolog.LevelSampler{
			TraceSampler: &logSamplerEnable{},
			DebugSampler: &logSamplerEnable{},
			InfoSampler:  &logSamplerEnable{},
			WarnSampler:  &logSamplerEnable{},
			ErrorSampler: &logSamplerEnable{},
		})
	case "debug":
		log.Logger = log.Sample(zerolog.LevelSampler{
			TraceSampler: &logSamplerDisable{},
			DebugSampler: &logSamplerEnable{},
			InfoSampler:  &logSamplerEnable{},
			WarnSampler:  &logSamplerEnable{},
			ErrorSampler: &logSamplerEnable{},
		})
	case "warn":
		log.Logger = log.Sample(zerolog.LevelSampler{
			TraceSampler: &logSamplerDisable{},
			DebugSampler: &logSamplerDisable{},
			InfoSampler:  &logSamplerDisable{},
			WarnSampler:  &logSamplerEnable{},
			ErrorSampler: &logSamplerEnable{},
		})
	case "error":
		log.Logger = log.Sample(zerolog.LevelSampler{
			TraceSampler: &logSamplerDisable{},
			DebugSampler: &logSamplerDisable{},
			InfoSampler:  &logSamplerDisable{},
			WarnSampler:  &logSamplerDisable{},
			ErrorSampler: &logSamplerEnable{},
		})
	case "info":
		fallthrough
	default:
		log.Logger = log.Sample(zerolog.LevelSampler{
			TraceSampler: &logSamplerDisable{},
			DebugSampler: &logSamplerDisable{},
			InfoSampler:  &logSamplerEnable{},
			WarnSampler:  &logSamplerEnable{},
			ErrorSampler: &logSamplerEnable{},
		})
	}
}

func main() {
	args := CmdArgs{
		file:           kingpin.Arg("file", "File to parse").String(),
		pattern:        kingpin.Flag("pattern", "Pattern to look for").Short('p').String(),
		dryRun:         kingpin.Flag("dry-run", "Dry-run mode").Default("false").Bool(),
		noFollow:       kingpin.Flag("no-follow", "Do not wait for the new data").Bool(),
		fromLineNumber: kingpin.Flag("from-line", "Start reading from this line number").Default("-1").Int(),
		config:         kingpin.Flag("config", "Path to the configuration").Short('c').String(),
		verbose:        kingpin.Flag("verbose", "Print every match").Short('v').Default("false").Bool(),
	}
	kingpin.Parse()

	// Assign every global variable first
	_isDryRun = *args.dryRun
	_verbose = *args.verbose

	// Initiate logging
	initLogging()

	var config *Config

	if *args.config == "" {
		if *args.pattern == "" || *args.file == "" {
			log.Fatal().Msg("Both file and pattern have to be specified, aborting")
		}

		log.Info().Msg("Using parameters from the command line")
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
		log.Info().Msg("Using parameters from the configuration file")
		if *args.pattern != "" || *args.file != "" {
			log.Fatal().Msg("No pattern/file allowed when configuration file is provided, aborting")
		}

		configPath := *args.config
		parsedConfig, err := ReadConfigFromFile(configPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed reading configuration from file")
		}
		config = parsedConfig
		log.Info().Msgf("Configuration file loaded: \"%s\"\n", configPath)
	}

	initSentry(config)

	// Create global condition
	cond := sync.NewCond(&sync.Mutex{})

	// Listen to interrupt/termination signal.
	// Once received, this will exit the program.
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGTERM, os.Interrupt, os.Kill)

	// Start the main program
	log.Debug().Msg("Starting sentlog")
	go runWithConfig(config, cond)

	// Listen to SIGHUP and SIGUSR1 for handling graceful restarts.
	// This is most useful for re-opening log files sentlog's tailing.
	// It's an unbuffered channel, you can send SIGHUP multiple times.
	usrStopSignal := make(chan os.Signal)
	signal.Notify(usrStopSignal, syscall.SIGHUP, syscall.SIGUSR1)
	go func() {
		for {
			log.Debug().Msg("Waiting for SIGHUP/SIGUSR1")
			<-usrStopSignal

			log.Debug().Msg("Received SIGHUP/SIGUSR1")
			cond.L.Lock()
			_killed = true
			cond.Broadcast()
			cond.L.Unlock()

			cond.L.Lock()
			_killed = false
			cond.L.Unlock()

			// Spawn new instance of the main application
			log.Debug().Msg("Starting sentlog")
			go runWithConfig(config, cond)
		}
	}()

	log.Debug().Msg("Waiting for Interrupt signal")
	<-exitSignal

	log.Debug().Msg("Interrupt received")
	// Handle graceful exit
	cond.L.Lock()
	_killed = true
	cond.Broadcast()
	cond.L.Unlock()

	log.Debug().Msg("Cleaning up...")
	// Sleep for 1 second to allow other goroutine to flush and exit properly
	time.Sleep(time.Second)
	sentry.Flush(5 * time.Second)
	log.Debug().Msg("Shutting down sentlog")
}
