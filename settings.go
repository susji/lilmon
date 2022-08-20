package main

import (
	"image/color"
	"regexp"
	"time"
)

const (
	FLAG_CONFIG_PATH    = "config-path"
	DEFAULT_CONFIG_PATH = "/etc/lilmon/lilmon.ini"
	HELP_CONFIG_PATH    = "Filepath to lilmon configuration file"

	DEFAULT_DB_PATH       = "/var/lilmon/db/lilmon.sqlite"
	DEFAULT_SHELL         = "/bin/sh"
	DEFAULT_ADDR          = "localhost:15515"
	DEFAULT_TEMPLATE_PATH = "/etc/lilmon/lilmon.template"
)

const (
	DEFAULT_GRAPH_PERIOD       = 1 * time.Hour
	DEFAULT_GRAPH_WIDTH        = 300
	DEFAULT_GRAPH_HEIGHT       = 100
	DEFAULT_RETENTION_TIME     = 90 * 24 * time.Hour
	DEFAULT_REFRESH_PERIOD     = 2 * time.Minute
	DEFAULT_PRUNE_PERIOD       = 15 * time.Minute
	DEFAULT_MEASUREMENT_PERIOD = 1 * time.Minute
	DEFAULT_BIN_WIDTH          = 1 * time.Minute
	DEFAULT_MAX_BINS           = DEFAULT_GRAPH_WIDTH / 1
	DEFAULT_GRAPH_FORMAT       = "svg"
	DEFAULT_GRAPH_MIMETYPE     = "image/svg+xml"
	CONFIG_DELIM               = "|"
)

var (
	COLOR_BG                = color.RGBA{255, 255, 255, 255}
	COLOR_FG                = color.RGBA{255, 0, 0, 255}
	COLOR_LABEL             = color.RGBA{0, 0, 0, 255}
	TIMESTAMP_FORMAT_YEAR   = "2006-01-02\n15:04"
	TIMESTAMP_FORMAT_MONTH  = "2006-01-02\n15:04"
	TIMESTAMP_FORMAT_DAY    = "Jan _2\n15:04"
	TIMESTAMP_FORMAT_HOUR   = "15:04"
	TIMESTAMP_FORMAT_MINUTE = "15:04:05"
)

var (
	RE_NAME = regexp.MustCompile("^[_a-zA-Z0-9]{1,512}$")
)
