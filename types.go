package main

import "time"

type params_measure struct {
	db_path, shell string
	period         time.Duration
}

type params_serve struct {
	db_path string
	addr    string
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
