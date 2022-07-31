package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/susji/tinyini"
)

func config_load(filepath string) (*config, error) {
	log.Println("attempting to read settings from ", filepath)
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("cannot open configuration file for reading: ", err)
		return nil, err
	}
	sections, errs := tinyini.Parse(f)
	if len(errs) != 0 {
		log.Println("errors when reading configuration file: ", filepath)
		for n, err := range errs {
			log.Printf("[%d] %v\n", n+1, err)
		}
		return nil, errors.New("invalid configuration file")
	}
	return &config{
		sections: sections,
	}, nil
}

func config_parse_metric_options(options string) (graph_options, []error) {
	ret := graph_options{}
	errs := []error{}
	for _, option := range strings.Split(strings.TrimSpace(options), ",") {
		split := strings.SplitN(option, "=", 2)
		key := strings.TrimSpace(strings.ToLower(split[0]))

		if len(key) == 0 {
			continue
		}

		var value string
		if len(split) == 2 {
			value = strings.TrimSpace(strings.ToLower(split[1]))
		}

		switch key {
		case "deriv":
			ret.differentiate = true
		case "y_min":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				errs = append(errs, fmt.Errorf("bad y_min value: %w", err))
			}
			ret.y_min = &val
		case "y_max":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				errs = append(errs, fmt.Errorf("bad y_max value: %w", err))
			}
			ret.y_max = &val
		default:
			errs = append(errs, fmt.Errorf("unrecognized graph option: %s", key))
		}
	}
	return ret, errs
}

func config_parse_metric_line(line string) (*metric, error) {
	vals := strings.SplitN(line, CONFIG_DELIM, 4)
	if len(vals) < 4 {
		return nil, fmt.Errorf(
			"line does not contain four %s-separated values, got %d",
			CONFIG_DELIM, len(vals))
	}
	options, errs := config_parse_metric_options(vals[2])
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

func (c *config) config_parse_metrics() ([]*metric, error) {
	metrics := []*metric{}

	pairs, ok := c.sections["metrics"]["metric"]
	if !ok {
		return nil, errors.New("no metrics defined in configuration file")
	}

	parse_in_err := false
	for _, pair := range pairs {
		metric, err := config_parse_metric_line(pair.Value)
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

func (c *config) config_parse_measure() (*config_measure, error) {
	ret := &config_measure{
		retention_time:  DEFAULT_RETENTION_TIME,
		prune_db_period: DEFAULT_PRUNE_PERIOD,
		measure_period:  DEFAULT_MEASUREMENT_PERIOD,
	}

	in_err := false
	for k, pairs := range c.sections["measure"] {
		for _, pair := range pairs {
			var err error
			switch k {
			case "retention_time":
				ret.retention_time, err = time.ParseDuration(pair.Value)
			case "prune_db_period":
				ret.prune_db_period, err = time.ParseDuration(pair.Value)
			case "measure_period":
				ret.measure_period, err = time.ParseDuration(pair.Value)
			}
			if err != nil {
				log.Printf("%s: invalid value: %v", k, err)
				in_err = true
			}
		}
	}
	if in_err {
		return nil, errors.New("parsing measure config failed")
	}

	return ret, nil
}

func (c *config) config_parse_serve() (*config_serve, error) {
	ret := &config_serve{
		width:  DEFAULT_GRAPH_WIDTH,
		height: DEFAULT_GRAPH_HEIGHT,

		pad_left:  DEFAULT_GRAPH_PAD_WIDTH_LEFT,
		pad_right: DEFAULT_GRAPH_PAD_WIDTH_RIGHT,
		pad_up:    DEFAULT_GRAPH_PAD_HEIGHT_UP,
		pad_down:  DEFAULT_GRAPH_PAD_HEIGHT_DOWN,

		label_max_y0:  DEFAULT_LABEL_MAX_Y0,
		label_shift_x: DEFAULT_LABEL_SHIFT_X,

		default_period:     DEFAULT_GRAPH_PERIOD,
		autorefresh_period: DEFAULT_REFRESH_PERIOD,
		bin_width:          DEFAULT_BIN_WIDTH,
	}

	in_err := false
	for k, pairs := range c.sections["serve"] {
		for _, pair := range pairs {
			var err error
			switch k {
			case "graph_width":
				ret.width, err = strconv.Atoi(pair.Value)
			case "graph_height":
				ret.height, err = strconv.Atoi(pair.Value)
			case "graph_pad_left":
				ret.pad_left, err = strconv.Atoi(pair.Value)
			case "graph_pad_right":
				ret.pad_right, err = strconv.Atoi(pair.Value)
			case "graph_pad_up":
				ret.pad_up, err = strconv.Atoi(pair.Value)
			case "graph_pad_down":
				ret.pad_down, err = strconv.Atoi(pair.Value)
			case "graph_label_max_y0":
				ret.label_max_y0, err = strconv.Atoi(pair.Value)
			case "graph_label_shift_x":
				ret.label_shift_x, err = strconv.Atoi(pair.Value)
			case "default_period":
				ret.default_period, err = time.ParseDuration(pair.Value)
			case "autorefresh_period":
				ret.autorefresh_period, err = time.ParseDuration(pair.Value)
			case "bin_width":
				ret.bin_width, err = time.ParseDuration(pair.Value)
			}
			if err != nil {
				log.Printf("%s: invalid value: %v", k, err)
				in_err = true
			}
		}
	}
	if in_err {
		return nil, errors.New("parsing measure config failed")
	}

	return ret, nil
}