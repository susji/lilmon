package main

import (
	"time"

	"github.com/susji/tinyini"
)

type config struct {
	sections map[string]tinyini.Section
}

type config_serve struct {
	width, height                                 int
	pad_left, pad_right, pad_up, pad_down         int
	label_max_y0, label_shift_x                   int
	default_period, autorefresh_period, bin_width time.Duration
}

type config_measure struct {
	retention_time, prune_db_period, measure_period time.Duration
}

type config_paths struct {
	db_path, config_path string
}

type params_measure struct {
	config_paths
	shell  string
	period time.Duration
}

type params_serve struct {
	config_paths
	template_path string
	addr          string
}

type metric struct {
	name, description, command string
	options                    graph_options
}

type graph_options struct {
	differentiate bool
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

	prune_metric_name      string
	prune_retention_period time.Duration

	insert_measurement *measurement
}
