# sentlog <!-- omit in toc -->

## This is a Sentry Hackweek project! Development may stop anytime. You've been warned.<!-- omit in toc -->

`sentlog` is a command-line tool that can read arbitrary text files (e.g., webserver or database logs), search for specific user-defined patterns, and report the findings to Sentry.

## Table of Contents <!-- omit in toc -->
- [Introduction](#Introduction)
- [Downloads](#Downloads)
- [Command Line Arguments](#Command-Line-Arguments)
- [Example](#Example)
- [Configuration File](#Configuration-File)

## Introduction

Sentry provides SDKs for a lot of different [platforms and frameworks](https://docs.sentry.io/). However, you might also want to use Sentry for parts of your infrastructure that were not developed by you, or don't have an integration with Sentry (yet): databases, web servers, and even operating system kernels. What do these tools have in common? They normally have some sort of output (i.e. logs), where both regular events and errors are usually logged. So why not parsing those logs and look for entries that look like errors? We can do that. And what platform do we usually use for error management? Sentry, of course!

And this is when `sentlog` steps in.

## Downloads

`sentlog` can be downloaded from [GitHub releases](https://github.com/getsentry/sentlog/releases).

## Command Line Arguments

```sh
usage: sentlog [<flags>] [<file>]

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
  -p, --pattern=PATTERN  Pattern to look for
      --dry-run          Dry-run mode
      --no-follow        Do not wait for the new data
      --from-line=-1     Start reading from this line number
  -c, --config=CONFIG    Path to the configuration
  -v, --verbose          Print every match

Args:
  [<file>]  File to parse
```

`sentlog` can operate in two modes:

1. Basic: filename and pattern are specified on the command line
2. Advanced: using the configuration file provided by `--config` argument

## Example

The following example shows how you can run `sentlog` in Basic mode.

```sh
export SENTLOG_SENTRY_DSN="https://XXX@sentry.io/YYY"   # Your Sentry DSN
sentlog /var/log/postgresql/postgresql-9.6.log \
        -p '^%{DATESTAMP:timestamp}.*FATAL:.*host "%{IP:host}", user "%{USERNAME:user}", database "%{WORD:database}"'
```

...will watch the PostgreSQL log (`/var/log/postgresql/postgresql-9.6.log`) for events that look like this:

```
2019-05-21 08:51:09 GMT [11212]: FATAL: no pg_hba.conf entry for host "123.123.123.123", user "postgres", database "testdb"
```

`sentlog` will extract the timestamp, IP address, username, and database from the entry, and will add them as tags to the Sentry event.

## Configuration File

```yaml
---
# Sentry DSN (also can be configured via environment)
sentry_dsn: https://XXX@sentry.io/YYY
# Additional Grok pattern files
pattern_files:
  - ./patterns1.txt
  - ../patterns2.txt

# List of files that we want to watch
inputs:
  - file: /var/log/nginx/error.log
    # Patterns to find and report
    patterns:
      - "%{NGINX_ERROR_LOG}"
    # Additional tags that will be added to the Sentry event
    tags:
      pattern: nginx_error
      custom: tag
```
