package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func exec_metric(ctx context.Context, m *metric, shell string, ord int, tasks chan<- db_task) {
	cmd := exec.CommandContext(ctx, shell, "-c", m.command)
	out, err := cmd.Output()
	if err != nil {
		log.Printf(
			"{%d}... run failed: %v\n",
			ord, err)
		return
	}
	cleaned := strings.TrimSpace(string(out))
	log.Printf(
		"{%d}... run worked and returned: %q\n",
		ord, cleaned)

	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		log.Printf(
			"{%d}... but it's not floaty: %v\n",
			ord, err)
		return
	}

	tasks <- db_task{
		kind: DB_TASK_INSERT,
		insert_measurement: &measurement{
			metric: m,
			value:  val,
		}}
}

func run_metrics(ctx context.Context, db *sql.DB, period time.Duration, shell string,
	metrics []*metric, tasks chan<- db_task) {

	log.Println("Entering measurement loop with period of ", period, "...")
	ord := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(period):
			for n, m := range metrics {
				ord++
				sctx, cf := context.WithTimeout(ctx, period/2+1)
				defer cf()
				log.Printf(
					"{%d} Running command %d/%d: %q\n",
					ord, n+1, n, m.command)
				go exec_metric(sctx, m, shell, ord, tasks)
			}
		}
	}
}

func is_metric_name_valid(name string) bool {
	return RE_NAME.MatchString(name)
}

func validate_metrics(metrics []*metric) error {
	in_err := false
	for n, m := range metrics {
		log.Printf("Validating metric name %d/%d: %s\n", n+1, len(metrics), m.name)
		if !is_metric_name_valid(m.name) {
			log.Println("... and the name is not valid.")
			in_err = true
		}
	}
	if in_err {
		return errors.New("one or more metrics did not validate")
	}
	return nil
}

func metric_find(metrics []*metric, name string) *metric {
	for _, cur := range metrics {
		if cur.name == name {
			return cur
		}
	}
	return nil
}
