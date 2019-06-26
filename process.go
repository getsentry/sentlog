package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getsentry/sentry-go"
	"github.com/hpcloud/tail"
	"github.com/vjeantet/grok"
)

const MessageField = "message"
const TimeStampField = "timestamp"

func PrintMap(m map[string]string) {
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

func CaptureEvent(line string, values map[string]string) {
	message := values[MessageField]
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

func ParseTimestamp(str string) int64 {
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

func ProcessLine(line string, pattern string, g *grok.Grok) {
	values, err := g.Parse(pattern, line)
	if err != nil {
		log.Printf("grok parsing failed: %v\n", err)
		os.Exit(1)
	}

	if !IsDryRun() {
		// Attempt to parse the timestamp
		timestamp := ParseTimestamp(values[TimeStampField])

		// Original log line
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Message:   line,
			Level:     sentry.LevelInfo,
			Timestamp: timestamp,
		})
	}

	if len(values) == 0 {
		return
	}

	CaptureEvent(line, values)

	log.Println("Entry found:")
	PrintMap(values)
}

func InitGrokProcessor() *grok.Grok {
	g, err := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err != nil {
		log.Fatalf("Grok engine initialization failed: %v\n", err)
	}
	AddDefaultPatterns(g)
	return g
}

func ProcessFile(filename string, pattern string) {
	g := InitGrokProcessor()

	file, err := os.Open(filename)
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

	log.Printf("Reading input from file \"%s\"", filename)

	var seekInfo tail.SeekInfo
	if *args.fromLineNumber < 0 {
		// By default: from the end
		seekInfo = tail.SeekInfo{
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
		for i := 0; i < *args.fromLineNumber; i++ {
			dataAvailable := scanner.Scan()
			if !dataAvailable {
				break
			}
		}
		seekInfo = tail.SeekInfo{
			Offset: pos,
			Whence: io.SeekStart,
		}
	}

	follow := !*args.noFollow
	tailFile, err := tail.TailFile(
		filename,
		tail.Config{
			Follow:   follow,
			Location: &seekInfo,
			ReOpen:   follow,
		})

	for line := range tailFile.Lines {
		ProcessLine(line.Text, pattern, g)

	}

	sentry.Flush(5 * time.Second)
}
