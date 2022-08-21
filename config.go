package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/susji/tinyini"
)

func config_load(r io.Reader) (*config, error) {
	sections, errs := tinyini.Parse(r)
	if len(errs) != 0 {
		log.Println("errors when reading configuration file")
		for n, err := range errs {
			log.Printf("[%d] %v\n", n+1, err)
		}
		return nil, errors.New("invalid configuration file")
	}
	return &config{
		sections: sections,
	}, nil

}
func config_load_file(filepath string) (*config, error) {
	log.Println("attempting to read settings from ", filepath)
	f, err := os.Open(filepath)
	if err != nil {
		log.Println("cannot open configuration file for reading: ", err)
		return nil, err
	}
	c, err := config_load(f)
	if err != nil {
		return nil, fmt.Errorf("unable to handle configuration file %q: %w", filepath, err)
	}
	return c, nil
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
		case "kibi":
			ret.kibi = true
		case "kilo":
			ret.kilo = true
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
		case "no_ds":
			ret.no_downsample = true
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

func (c *config) parse_metrics() ([]*metric, error) {
	metrics := []*metric{}
	in_err := false
	for k, pairs := range c.sections["metrics"] {
		for _, pair := range pairs {
			switch k {
			case "metric":
				metric, err := config_parse_metric_line(pair.Value)
				if err != nil {
					log.Printf(
						"%d: parsing metric line failed: %v\n",
						pair.Lineno, err)
					in_err = true
					continue
				}
				metrics = append(metrics, metric)
			default:
				log.Printf(
					"metrics section supports only 'metric' definitions "+
						"but line %d has something else.", pair.Lineno)
				in_err = true
			}
		}

	}

	if err := validate_metrics(metrics); err != nil {
		log.Println("metrics validation failed: ", err)
		return nil, err
	}
	if in_err {
		return nil, errors.New("metrics section contained errors")
	}
	return metrics, nil
}

func (c *config) parse_common() (string, time.Duration, error) {
	var path_db string
	measure_period := DEFAULT_MEASUREMENT_PERIOD

	in_err := false

	for k, pairs := range c.sections[""] {
		for _, pair := range pairs {
			var err error
			switch k {
			case "path_db":
				path_db = pair.Value
			case "measure_period":
				measure_period, err = time.ParseDuration(pair.Value)
				if err == nil && measure_period.Seconds() < 1 {
					err = errors.New("must be at last 1 second")
				}
			default:
				err = fmt.Errorf("%d: unrecognized config item: %s",
					pair.Lineno, k)
			}
			if err != nil {
				log.Printf("%s: invalid value: %v", k, err)
				in_err = true
			}
		}
	}
	if in_err {
		return "", time.Duration(0), errors.New("errors in common section")
	}
	if path_db == "" {
		return "", time.Duration(0), errors.New("no database path in common section")
	}
	return path_db, measure_period, nil
}

func (c *config) parse_measure() (*config_measure, error) {
	ret := &config_measure{
		retention_time:  DEFAULT_RETENTION_TIME,
		prune_db_period: DEFAULT_PRUNE_PERIOD,
		path_db:         DEFAULT_DB_PATH,
		shell:           DEFAULT_SHELL,
	}

	in_err := false

	if path_db, measure_period, cerr := c.parse_common(); cerr == nil {
		ret.path_db = path_db
		ret.measure_period = measure_period

	} else {
		in_err = true
		log.Println(cerr)
	}

	for k, pairs := range c.sections["measure"] {
		for _, pair := range pairs {
			var err error
			switch k {
			case "retention_time":
				ret.retention_time, err = time.ParseDuration(pair.Value)
			case "prune_db_period":
				ret.prune_db_period, err = time.ParseDuration(pair.Value)
			case "shell":
				ret.shell = pair.Value
			default:
				err = fmt.Errorf(
					"%d: unrecognized config item: %s",
					pair.Lineno, k)
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

func (c *config) parse_serve() (*config_serve, error) {
	ret := &config_serve{
		width:  DEFAULT_GRAPH_WIDTH,
		height: DEFAULT_GRAPH_HEIGHT,

		default_period:     DEFAULT_GRAPH_PERIOD,
		autorefresh_period: DEFAULT_REFRESH_PERIOD,
		bin_width:          DEFAULT_BIN_WIDTH,
		max_bins:           DEFAULT_MAX_BINS,
		downsampling_scale: DEFAULT_DOWNSAMPLING_SCALE,

		graph_format:   DEFAULT_GRAPH_FORMAT,
		graph_mimetype: DEFAULT_GRAPH_MIMETYPE,

		path_db:       DEFAULT_DB_PATH,
		path_template: DEFAULT_TEMPLATE_PATH,

		listen_addr: DEFAULT_ADDR,
	}

	in_err := false

	if path_db, measure_period, cerr := c.parse_common(); cerr == nil {
		ret.path_db = path_db
		ret.measure_period = measure_period
	} else {
		in_err = true
		log.Println(cerr)
	}

	for k, pairs := range c.sections["serve"] {
		for _, pair := range pairs {
			var err error
			switch k {
			case "graph_width":
				ret.width, err = strconv.Atoi(pair.Value)
			case "graph_height":
				ret.height, err = strconv.Atoi(pair.Value)
			case "default_period":
				ret.default_period, err = time.ParseDuration(pair.Value)
			case "autorefresh_period":
				ret.autorefresh_period, err = time.ParseDuration(pair.Value)
			case "bin_width":
				ret.bin_width, err = time.ParseDuration(pair.Value)
			case "max_bins":
				ret.max_bins, err = strconv.Atoi(pair.Value)
			case "downsampling_scale":
				ret.downsampling_scale, err = strconv.Atoi(pair.Value)
				if err == nil && ret.downsampling_scale < 1 {
					err = errors.New("must be greater than zero")
				}
			case "graph_format":
				ret.graph_format = pair.Value
			case "graph_mimetype":
				ret.graph_mimetype = pair.Value
			case "listen_addr":
				ret.listen_addr = pair.Value
			case "path_template":
				ret.path_template = pair.Value
			default:
				err = fmt.Errorf(
					"%d: unrecognized config item: %s",
					pair.Lineno, k)
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
