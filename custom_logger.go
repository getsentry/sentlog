package main

import "github.com/rs/zerolog"

type logSamplerEnable struct{}

func (l logSamplerEnable) Sample(lvl zerolog.Level) bool {
	return true
}

type logSamplerDisable struct{}

func (l logSamplerDisable) Sample(lvl zerolog.Level) bool {
	return false
}
