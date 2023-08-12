package main

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getsentry/sentry-go"
	"github.com/hpcloud/tail"
	"github.com/rs/zerolog/log"
	"github.com/vjeantet/grok"
)

const MessageField = "message"
const TimeStampField = "timestamp"

func captureEvent(line string, values map[string]string, hub *sentry.Hub) {
	if isDryRun() {
		return
	}

	message := values[MessageField]
	if message == "" {
		message = line
	}

	// Attempt to parse the timestamp
	timestamp := parseTimestamp(values[TimeStampField])

	scope := hub.Scope()

	for key, value := range values {
		if value == "" {
			continue
		}
		scope.SetTag(key, value)
	}

	if timestamp != 0 {
		scope.SetTag("parsed_timestamp", strconv.FormatInt(timestamp, 10))
	}

	scope.SetLevel(sentry.LevelError)

	scope.SetExtra("log_entry", line)

	hub.CaptureMessage(message)
}

func parseTimestamp(str string) int64 {
	fallback := int64(0)
	if str == "" {
		return fallback
	}

	localTime, err := dateparse.ParseLocal(str)
	if err != nil {
		return fallback
	}

	return localTime.Unix()
}

func processLine(line string, patterns []string, g *grok.Grok, hub *sentry.Hub) {
	var parsedValues map[string]string

	// Try all patterns
	for _, pattern := range patterns {
		values, err := g.Parse(pattern, line)
		if err != nil {
			log.Fatal().Err(err).Msg("grok parsing failed")
		}

		if len(values) != 0 {
			parsedValues = values
			hub.Scope().SetExtra("pattern", pattern)
			break
		}
	}

	if len(parsedValues) == 0 {
		return
	}

	captureEvent(line, parsedValues, hub)

	log.Debug().Any("parsed_values", parsedValues).Msg("Entry found")
}

func initGrokProcessor() *grok.Grok {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		log.Fatal().Err(err).Msg("Grok engine initialization failed")
	}

	if err := AddDefaultPatterns(g); err != nil {
		log.Fatal().Err(err).Msg("Processing default patterns")
	}

	return g
}

func getSeekInfo(file *os.File, fromLineNumber int) tail.SeekInfo {
	if fromLineNumber < 0 {
		// By default: from the end
		return tail.SeekInfo{
			Offset: 0,
			Whence: io.SeekEnd,
		}
	} else {
		// Seek to the line number
		scanner := bufio.NewScanner(file)
		pos := int64(0)
		scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			advance, token, err = bufio.ScanLines(data, atEOF)
			pos += int64(advance)
			return
		}
		scanner.Split(scanLines)
		for i := 0; i < fromLineNumber; i++ {
			dataAvailable := scanner.Scan()
			if !dataAvailable {
				break
			}
		}
		return tail.SeekInfo{
			Offset: pos,
			Whence: io.SeekStart,
		}
	}
}

func processFile(fileInput *FileInputConfig, g *grok.Grok, cond *sync.Cond) {
	absFilePath, err := filepath.Abs(fileInput.File)
	if err != nil {
		log.Fatal().Err(err).Msgf("Getting absolute file path of %q", fileInput.File)
	}

	log.Debug().Str("path", absFilePath).Msg("Opening file")

	file, err := os.Open(absFilePath)
	if err != nil {
		log.Fatal().Err(err).Msg("Opening file")
	}
	defer func() {
		log.Debug().Str("path", absFilePath).Msg("Closing file")

		err := file.Close()
		if err != nil {
			log.Error().Err(err).Str("path", absFilePath).Msg("Closing file")
		}
	}()

	info, err := file.Stat()
	if err != nil {
		log.Fatal().Err(err).Msg("Executing file stat operation")
	}

	if info.IsDir() {
		log.Fatal().Msg("Directory paths are not allowed, exiting")
	}

	log.Info().Msgf("Reading input from file %q", absFilePath)

	// One hub per file/goroutine
	hub := sentry.CurrentHub().Clone()
	scope := hub.PushScope()
	scope.SetTag("file_input_path", absFilePath)
	scope.SetTags(fileInput.Tags)

	fromLineNumber := -1
	if fileInput.FromLineNumber != nil {
		fromLineNumber = *fileInput.FromLineNumber
	}

	seekInfo := getSeekInfo(file, fromLineNumber)

	follow := true
	if fileInput.Follow != nil {
		follow = *fileInput.Follow
	}

	tailFile, err := tail.TailFile(
		absFilePath,
		tail.Config{
			Follow:   follow,
			Location: &seekInfo,
			ReOpen:   follow,
		})
	defer func() {
		log.Debug().Str("path", absFilePath).Msg("Stop tailing")
		err := tailFile.Stop()
		if err != nil {
			log.Error().Err(err).Str("path", absFilePath).Msg("Stop tailing file")
		}

		tailFile.Cleanup()
	}()

	// Run the tail function on a separate goroutine,
	// so current processFile can wait for a kill signal broadcast.
	go func() {
		log.Debug().Str("file", fileInput.File).Msg("Start tailing file")
		for line := range tailFile.Lines {
			hub.WithScope(func(_ *sentry.Scope) {
				processLine(line.Text, fileInput.Patterns, g, hub)
			})

			if !isDryRun() {
				hub.AddBreadcrumb(&sentry.Breadcrumb{
					Message: line.Text,
					Level:   sentry.LevelInfo,
				}, nil)
			}
		}
	}()

	// Wait for killl signal broadcast.
	cond.L.Lock()
	for !_killed {
		cond.Wait()
	}
	cond.L.Unlock()

	// Gracefully handles Sentry flush and close file descriptors.
	log.Debug().Msgf("Finished reading from %q, flushing events...", absFilePath)
	hub.Flush(10 * time.Second)
}

func runWithConfig(config *Config, cond *sync.Cond) {
	log.Debug().Msg("Verbose mode enabled, printing every match")

	g := initGrokProcessor()

	// Load patterns
	for _, filename := range config.PatternFiles {
		err := ReadPatternsFromFile(g, filename)
		if err != nil {
			log.Fatal().Err(err).Msg("Reading patterns from file")
		}
		log.Info().Msgf("Loaded additional patterns from \"%s\"\n", filename)
	}

	if len(config.Inputs) == 0 {
		log.Fatal().Msg("No file inputs specified, aborting")
	}

	// Process file inputs
	for _, fileInput := range config.Inputs {
		go processFile(&fileInput, g, cond)
	}

	// Wait for kill signal broadcast
	cond.L.Lock()
	for !_killed {
		cond.Wait()
	}
	cond.L.Unlock()
}
