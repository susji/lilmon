package main

import (
	"time"

	"github.com/susji/tinyini"
)

type config struct {
	sections map[string]tinyini.Section
}

type config_serve struct {
	width, height, max_bins, downsampling_scale                   int
	default_period, autorefresh_period, measure_period, bin_width time.Duration
	graph_format, graph_mimetype                                  string
	listen_addr                                                   string
	path_template, path_db                                        string
}

type config_measure struct {
	retention_time, prune_db_period, measure_period time.Duration
	path_db                                         string
	shell                                           string
}

type metric struct {
	name, description, command string
	options                    graph_options
}

type graph_options struct {
	differentiate bool
	kibi, kilo    bool
	no_downsample bool
	y_min, y_max  *float64
}

type measurement struct {
	metric *metric
	value  float64
}

type datapoint struct {
	ts    time.Time
	value float64
}

type bin_op func(i int, vals []float64, times []time.Time) float64

const (
	DB_TASK_PRUNE_TABLE = iota
	DB_TASK_INSERT
)

type db_task struct {
	kind int

	prune_metric           *metric
	prune_retention_period time.Duration

	insert_measurement *measurement
}
