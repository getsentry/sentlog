package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/vjeantet/grok"
)

func AddDefaultPatterns(g *grok.Grok) (err error) {
	// Nginx
	err = g.AddPattern("NGINX_ERROR_DATESTAMP", `\d{4}/\d{2}/\d{2}[- ]%{TIME}`)
	if err != nil {
		return err
	}

	err = g.AddPattern("NGINX_ERROR_LOG", `%{NGINX_ERROR_DATESTAMP:timestamp} \[%{DATA:err_severity}\] (%{NUMBER:pid:int}#%{NUMBER}: \*%{NUMBER}|\*%{NUMBER}) %{DATA:message}(?:, client: "?%{IPORHOST:client}"?)(?:, server: %{IPORHOST:server})(?:, request: "%{WORD:verb} %{URIPATHPARAM:request} HTTP/%{NUMBER:httpversion}")?(?:, upstream: "%{DATA:upstream}")?(?:, host: "%{URIHOST:host}")?(?:, referrer: "%{URI:referrer}")?`)
	if err != nil {
		return err
	}

	return nil
}

func ReadPatternsFromFile(g *grok.Grok, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Printf("Adding grok patterns from \"%s\"", filename)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Cannot parse patterns in \"%s\"", filename)
		}

		patternName, pattern := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if patternName == "" || pattern == "" {
			return fmt.Errorf("Empty pattern definition in \"%s\"", filename)
		}

		err := g.AddPattern(patternName, pattern)
		if err != nil {
			return err
		}
	}

	return nil
}
