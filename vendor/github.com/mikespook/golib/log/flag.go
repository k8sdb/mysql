// Copyright 2012 Xing Xing <mikespook@gmail.com>.
// All rights reserved.
// Use of this source code is governed by a commercial
// license that can be found in the LICENSE file.

package log

import (
	"flag"
	"strings"
)

var (
	LogFile, LogLevel string
)

func StrToLevel(str string) int {
	level := LogNone
	levels := strings.SplitN(str, "|", -1)
	for _, v := range levels {
		switch v {
		case "error":
			level = level | LogError
		case "warning":
			level = level | LogWarning
		case "message":
			level = level | LogMessage
		case "debug":
			level = level | LogDebug
		case "all":
			level = LogAll
		case "none":
			fallthrough
		default:
			level = level | LogNone
		}
	}
	return level
}

func InitWithFlag() {
	flag.StringVar(&LogFile, "log", "", "log to write (empty for STDOUT)")
	flag.StringVar(&LogLevel, "log-level", "all", "log level "+
		"('error', 'warning', 'message', 'debug', 'all' and 'none'"+
		" are combined with '|')")

	if err := Init(LogFile, StrToLevel(LogLevel), DefaultCallDepth); err != nil {
		Error(err)
	}
}
