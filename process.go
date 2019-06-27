package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getsentry/sentry-go"
	"github.com/hpcloud/tail"
	"github.com/vjeantet/grok"
)

const MessageField = "message"
const TimeStampField = "timestamp"

var wg sync.WaitGroup

var printMutex sync.Mutex

func printMap(m map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%+15s: %s\n", k, m[k])
	}
	fmt.Println()
}

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

	time, err := dateparse.ParseLocal(str)
	if err != nil {
		return fallback
	}

	return time.Unix()
}

func processLine(line string, patterns []string, g *grok.Grok, hub *sentry.Hub) {
	var parsedValues map[string]string

	// Try all patterns
	for _, pattern := range patterns {
		values, err := g.Parse(pattern, line)
		if err != nil {
			log.Fatalf("grok parsing failed: %v\n", err)
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

	if isVerbose() {
		printMutex.Lock()
		log.Println("Entry found:")
		printMap(parsedValues)
		printMutex.Unlock()
	}
}

func initGrokProcessor() *grok.Grok {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		log.Fatalf("Grok engine initialization failed: %v\n", err)
	}
	AddDefaultPatterns(g)
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

func processFile(fileInput *FileInputConfig, g *grok.Grok) {
	defer wg.Done()

	absFilePath, err := filepath.Abs(fileInput.File)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Open(absFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		log.Fatal("Directory paths are not allowed, exiting")
	}

	log.Printf("Reading input from file \"%s\"", absFilePath)

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

	log.Printf("Finished reading from \"%s\", flushing events...\n", absFilePath)
	hub.Flush(10 * time.Second)
}

func runWithConfig(config *Config) {
	if isVerbose() {
		log.Println("Verbose mode enabled, printing every match")
	}

	g := initGrokProcessor()

	// Load patterns
	for _, filename := range config.PatternFiles {
		err := ReadPatternsFromFile(g, filename)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Loaded additional patterns from \"%s\"\n", filename)
	}

	if len(config.Inputs) == 0 {
		log.Fatalln("No file inputs specified, aborting")
	}

	// Process file inputs
	for _, fileInput := range config.Inputs {
		wg.Add(1)
		go processFile(&fileInput, g)
	}

	wg.Wait()
}
