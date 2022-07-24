package main

import (
	"image/color"
	"regexp"
	"time"
)

const (
	FLAG_DB_PATH    = "db-path"
	DEFAULT_DB_PATH = "/tmp/lilmon.sqlite"
	HELP_DB_PATH    = "Filepath to lilmon SQLite database"

	FLAG_SHELL    = "shell"
	DEFAULT_SHELL = "/bin/sh"
	HELP_SHELL    = "Filepath for shell to use when measuring metrics"

	FLAG_PERIOD    = "period"
	DEFAULT_PERIOD = 20 * time.Second
	HELP_PERIOD    = "How often to take new measurements"

	FLAG_ADDR    = "addr"
	DEFAULT_ADDR = "localhost:15515"
	HELP_ADDR    = "HTTP listening address"
)

const (
	DEFAULT_GRAPH_PERIOD          = 1 * time.Hour
	DEFAULT_GRAPH_WIDTH           = 600
	DEFAULT_GRAPH_HEIGHT          = 200
	DEFAULT_GRAPH_PAD_WIDTH_LEFT  = 10
	DEFAULT_GRAPH_PAD_HEIGHT_UP   = 25
	DEFAULT_GRAPH_PAD_WIDTH_RIGHT = 70
	DEFAULT_GRAPH_PAD_HEIGHT_DOWN = 25
	DEFAULT_GRAPH_BINS            = DEFAULT_GRAPH_WIDTH / 10
	DEFAULT_LABEL_MAX_Y0          = 10
	DEFAULT_LABEL_SHIFT_X         = DEFAULT_GRAPH_PAD_WIDTH_RIGHT * 2.5
	DEFAULT_RETENTION_TIME        = 7 * 24 * time.Hour
	DEFAULT_REFRESH_PERIOD        = 60
	DEFAULT_PRUNE_PERIOD          = 15 * time.Minute
)

var (
	COLOR_BG    = color.RGBA{230, 230, 230, 255}
	COLOR_FG    = color.RGBA{255, 0, 0, 255}
	COLOR_LABEL = color.RGBA{0, 0, 0, 255}
	// 01/02 03:04:05PM '06 -0700
	TIMESTAMP_FORMAT = "2006-01-02 03:04 MST"
)

var (
	RE_NAME = regexp.MustCompile("^[-_a-zA-Z0-9]{1,512}$")
)
