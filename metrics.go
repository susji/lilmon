package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/susji/tinyini"
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

func metrics_parse_options(options string) (graph_options, []error) {
	ret := graph_options{}
	errs := []error{}
	for _, option := range strings.Split(strings.TrimSpace(options), ",") {
		vals := strings.SplitN(option, "=", 2)
		key := strings.TrimSpace(strings.ToLower(vals[0]))
		if len(key) == 0 {
			continue
		}
		switch key {
		case "deriv":
			ret.differentiate = true
		default:
			errs = append(errs, fmt.Errorf("unrecognized graph option: %s", key))
		}
	}
	return ret, errs
}

func metrics_parse_line(line string) (*metric, error) {
	vals := strings.SplitN(line, CONFIG_DELIM, 4)
	if len(vals) < 4 {
		return nil, fmt.Errorf(
			"line does not contain four %s-separated values, got %d",
			CONFIG_DELIM, len(vals))
	}
	options, errs := metrics_parse_options(vals[2])
	if len(errs) > 0 {
		return nil, fmt.Errorf(
			"%s: invalid graph options: %v", vals[0], errs)

	}

	m := &metric{
		name:        vals[0],
		description: vals[1],
		options:     options,
		command:     vals[3],
	}

	return m, nil
}

func metrics_load(filepath string) ([]*metric, error) {
	log.Println("attempting to read metrics from ", filepath)
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("cannot open configuration file for reading: ", err)
		return nil, err
	}
	ini, errs := tinyini.Parse(f)
	if len(errs) != 0 {
		log.Println("errors when reading configuration file: ", filepath)
		for n, err := range errs {
			log.Printf("[%d] %v\n", n+1, err)
		}
		return nil, errors.New("invalid configuration file")
	}
	metrics := []*metric{}

	pairs, ok := ini[""]["metric"]
	if !ok {
		return nil, errors.New("no metrics defined in configuration file")
	}

	parse_in_err := false
	for _, pair := range pairs {
		metric, err := metrics_parse_line(pair.Value)
		if err != nil {
			log.Printf("%d: parsing metric line failed: %v\n", pair.Lineno, err)
			parse_in_err = true
			continue
		}
		metrics = append(metrics, metric)
	}
	if err := validate_metrics(metrics); err != nil {
		log.Println("metrics validation failed: ", err)
		return nil, err
	}
	if parse_in_err {
		return nil, errors.New("metrics parsing failed")
	}
	return metrics, nil
}

func metric_find(metrics []*metric, name string) *metric {
	for _, cur := range metrics {
		if cur.name == name {
			return cur
		}
	}
	return nil
}
