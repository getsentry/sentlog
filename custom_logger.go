package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// logSamplerEnable enables zerolog logging for current level
// specified on the Sample() implementation that will always
// return true.
type logSamplerEnable struct{}

func (l logSamplerEnable) Sample(lvl zerolog.Level) bool {
	return true
}

// logSamplerDisable enables zerolog logging for current level
// specified on the Sample() implementation that will always
// return false.
type logSamplerDisable struct{}

func (l logSamplerDisable) Sample(lvl zerolog.Level) bool {
	return false
}

// tailLogger is a zerolog wrapper for tail package Sentlog is using.
type tailLogger struct{}

func (t tailLogger) Fatal(v ...interface{}) {
	log.Fatal().Msgf("%v", v)
}

func (t tailLogger) Fatalf(format string, v ...interface{}) {
	log.Fatal().Msgf(format, v...)
}

func (t tailLogger) Fatalln(v ...interface{}) {
	log.Fatal().Msgf("%v", v)
}

func (t tailLogger) Panic(v ...interface{}) {
	log.Panic().Msgf("%v", v)
}

func (t tailLogger) Panicf(format string, v ...interface{}) {
	log.Panic().Msgf(format, v...)
}

func (t tailLogger) Panicln(v ...interface{}) {
	log.Panic().Msgf("%v", v)
}

func (t tailLogger) Print(v ...interface{}) {
	log.Debug().Msgf("%v", v)
}

func (t tailLogger) Printf(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}

func (t tailLogger) Println(v ...interface{}) {
	log.Debug().Msgf("%v", v)
}
