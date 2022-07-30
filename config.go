package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
