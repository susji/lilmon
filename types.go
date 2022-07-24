package main

import "time"

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
	addr string
}

type metric struct {
	name, description, command string
}

type measurement struct {
	metric *metric
	value  float64
}

type datapoint struct {
	ts    time.Time
	value float64
}

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
