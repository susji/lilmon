package main

import (
	"image/color"
	"regexp"
	"time"
)

const (
	FLAG_DB_PATH    = "db-path"
	DEFAULT_DB_PATH = "/var/lilmon/lilmon.sqlite"
	HELP_DB_PATH    = "Filepath to lilmon SQLite database"

	FLAG_CONFIG_PATH    = "config-path"
	DEFAULT_CONFIG_PATH = "/etc/lilmon.ini"
	HELP_CONFIG_PATH    = "Filepath to lilmon configuration file"

	FLAG_SHELL    = "shell"
	DEFAULT_SHELL = "/bin/sh"
	HELP_SHELL    = "Filepath for shell to use when measuring metrics"

	FLAG_ADDR    = "addr"
	DEFAULT_ADDR = "localhost:15515"
	HELP_ADDR    = "HTTP listening address"

	FLAG_TEMPLATE_PATH    = "template-path"
	DEFAULT_TEMPLATE_PATH = "/etc/lilmon.template"
	HELP_TEMPLATE_PATH    = "Filepath to monitoring page's HTML template"
)

const (
	DEFAULT_GRAPH_PERIOD       = 1 * time.Hour
	DEFAULT_GRAPH_WIDTH        = 300
	DEFAULT_GRAPH_HEIGHT       = 100
	DEFAULT_RETENTION_TIME     = 90 * 24 * time.Hour
	DEFAULT_REFRESH_PERIOD     = 30
	DEFAULT_PRUNE_PERIOD       = 15 * time.Minute
	DEFAULT_MEASUREMENT_PERIOD = 15 * time.Second
	DEFAULT_BIN_WIDTH          = 1 * time.Minute
	DEFAULT_MAX_BINS           = DEFAULT_GRAPH_WIDTH / 2
	DEFAULT_GRAPH_FORMAT       = "svg"
	DEFAULT_GRAPH_MIMETYPE     = "image/svg+xml"
	CONFIG_DELIM               = "|"
)

var (
	COLOR_BG                = color.RGBA{230, 230, 230, 255}
	COLOR_FG                = color.RGBA{255, 0, 0, 255}
	COLOR_LABEL             = color.RGBA{0, 0, 0, 255}
	TIMESTAMP_FORMAT_YEAR   = time.UnixDate
	TIMESTAMP_FORMAT_MONTH  = "Jan _2"
	TIMESTAMP_FORMAT_DAY    = time.Stamp
	TIMESTAMP_FORMAT_HOUR   = "15:04"
	TIMESTAMP_FORMAT_MINUTE = "15:04:05"
)

var (
	RE_NAME = regexp.MustCompile("^[_a-zA-Z0-9]{1,512}$")
)
