package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/vjeantet/grok"
)

func AddDefaultPatterns(g *grok.Grok) (err error) {
	patterns := map[string]string{
		// Nginx
		"NGINX_ERROR_DATESTAMP": `\d{4}/\d{2}/\d{2}[- ]%{TIME}`,
		"NGINX_ERROR_LOG":       `%{NGINX_ERROR_DATESTAMP:timestamp} \[%{DATA:err_severity}\] (%{NUMBER:pid:int}#%{NUMBER}: \*%{NUMBER}|\*%{NUMBER}) %{DATA:message}(?:, client: "?%{IPORHOST:client}"?)(?:, server: %{IPORHOST:server})(?:, request: "%{WORD:verb} %{URIPATHPARAM:request} HTTP/%{NUMBER:httpversion}")?(?:, upstream: "%{DATA:upstream}")?(?:, host: "%{URIHOST:host}")?(?:, referrer: "%{URI:referrer}")?`,
		// Rsyslog
		"RSYSLOGMESSAGE":      `(?:%{GREEDYDATA:syslog_message})`,
		"RSYSLOGCUSTOMHEADER": `(?:(?:<%{NONNEGINT:syslog_abspri}>(?:%{NONNEGINT:syslog_version} )?)?(?<syslog_timestamp>%{TIMESTAMP_ISO8601:}|%{SYSLOGTIMESTAMP}))`,
		"RSYSLOGPREFIX":       `%{RSYSLOGCUSTOMHEADER} %{IPORHOST:syslog_hostname} (?:%{PROG:program}(?:\[%{POSINT:pid}\])?: )?`,
		"RSYSLOGCUSTOM":       `%{RSYSLOGPREFIX}%{RSYSLOGMESSAGE}`,
		// Postgres
		"PGPREFIX":     `\[%{NUMBER}-1\] user=%{DATA:postgres_username}-%{IPORHOST:postgres_client},db=%{DATA:postgres_dbname}`,
		"PGLOGPREFIX":  `%{PGPREFIX}%{SPACE}LOG:`,
		"PGQUERY":      `%{PGLOGPREFIX}%{SPACE}statement: %{GREEDYDATA:postgres_query}`,
		"PGDURATION":   `%{PGLOGPREFIX}%{SPACE}duration: %{DATA:postgres_duration} ms`,
		"PGDISCONNECT": `%{PGLOGPREFIX}%{SPACE}disconnection: session time: %{DATA:postgres_sessiontime} user=%{DATA:postgres_username} database=%{DATA:postgres_dbname} host=%{IPORHOST:postgres_client} port=%{NUMBER:postgres_clientport}`,
		"PGMESSAGE":    `%{PGLOGPREFIX} %{SPACE}%{GREEDYDATA:postgres_message}`,
		"PGCONNECT":    `\[%{NUMBER}-1\] %{SPACE}user=\[unknown\]-,db=\[unknown\] %{SPACE}LOG: %{SPACE}connection %{SPACE}received: %{SPACE}host=%{IPORHOST:postgres_client} %{SPACE}port=%{NUMBER:postgres_clientport}`,
		"PGFATAL":      `%{PGPREFIX} %{SPACE}FATAL: %{SPACE}%{GREEDYDATA:postgres_message}`,
		"POSTGRES":     `%{PGQUERY}|%{PGCONNECT}|%{PGDISCONNECT}|%{PGDURATION}|%{PGFATAL}|%{PGMESSAGE}`,
	}

	return g.AddPatternsFromMap(patterns)
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
