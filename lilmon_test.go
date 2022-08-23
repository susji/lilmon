package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

var test_config = `
path_db=/somewhere/db
measure_period=150s

[measure]
retention_time=21600h
prune_db_period=300m
shell=/bin/zsh

[serve]
listen_addr=localhost:15516
path_template=/somewhere/template
default_period=10m
graph_width=3000
graph_height=1000
bin_width=10m
downsampling_scale=3
max_bins=1500
autorefresh_period=600s
graph_format=png
graph_mimetype=image/png

[metrics]
metric=n_temp_files|Files in /tmp|y_min=0,kilo|find /tmp/ -type f|wc -l
metric=n_processes|Visible processes (all users)|y_min=0,y_max=1000|ps -A|wc -l
metric=rate_logged_in_users|Rate of user logins|deriv|who|wc -l
metric="n_subshell_constant|Plain silly||{ echo -n \"one\"; echo -n two; echo -n three; }|wc -c"
`

var test_metrics = []*metric{
	{
		name:        "test_metric_1",
		description: "a simple test metric",
		command:     `echo hello world!|wc -c`,
		options: graph_options{
			no_downsample: true,
		},
	},
}

func assert(t *testing.T, cond bool, msg ...interface{}) {
	if cond {
		return
	}
	t.Error(msg...)
}

func assertf(t *testing.T, cond bool, format string, msg ...interface{}) {
	if cond {
		return
	}
	t.Errorf(format, msg...)
}

func almost_equals(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestBinDatapoints(t *testing.T) {
	ta, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")
	tb, _ := time.Parse(time.RFC3339, "2020-01-01T13:00:00Z")

	dps := []datapoint{
		// FIRST BIN BEGINS
		{
			ts:    ta.Add(5 * time.Minute),
			value: 10,
		},
		{
			ts:    ta.Add(10 * time.Minute),
			value: 20,
		},
		{
			ts:    ta.Add(10 * time.Minute),
			value: 0,
		},
		// SECOND BIN BEGINS
		{
			ts:    ta.Add(16 * time.Minute),
			value: 100,
		},
		// THIRD BIN BEGINS
		// FOURTH BIN BEGINS
		{
			ts:    ta.Add(57 * time.Minute),
			value: -10,
		},
		{
			ts:    ta.Add(59*time.Minute + 59*time.Second),
			value: 20,
		},

		// PAST LAST BIN
		{
			ts:    ta.Add(61 * time.Minute),
			value: 100000,
		},
	}
	bins := 4

	recorded_prev_vals := make([]float64, bins)
	test_op_ident := func(i int, vals []float64, times []time.Time) float64 {
		if i == 0 {
			return vals[i]
		}
		recorded_prev_vals[i-1] = vals[i]
		return vals[i]
	}

	got_values, got_labels, got_min, got_max := bin_datapoints(
		dps, int64(bins), ta, tb, test_op_ident)
	t.Log("got_values:        ", got_values)
	t.Log("got_labels:        ", got_labels)
	t.Log("recorded_prev_vals:", recorded_prev_vals)

	want_min := 5.0
	want_max := 100.0

	if !almost_equals(got_min, want_min) {
		t.Errorf("wrong min, wanted %f but got %f", want_min, got_min)
	}
	if !almost_equals(got_max, want_max) {
		t.Errorf("wrong max, wanted %f but got %f", want_max, got_max)
	}

	if len(got_values) != bins {
		t.Fatal("wrong number of value bins, got ", len(got_values))
	}
	if len(got_labels) != bins {
		t.Fatal("wrong number of bin labels, got ", len(got_values))
	}

	want_values := []float64{(10 + 20 + 0) / 3, 100, math.NaN(), (-10 + 20) / 2}
	label_bin1, _ := time.Parse(time.RFC3339, "2020-01-01T12:07:30Z")
	label_bin2, _ := time.Parse(time.RFC3339, "2020-01-01T12:22:30Z")
	label_bin3, _ := time.Parse(time.RFC3339, "2020-01-01T12:37:30Z")
	label_bin4, _ := time.Parse(time.RFC3339, "2020-01-01T12:52:30Z")
	want_labels := []time.Time{label_bin1, label_bin2, label_bin3, label_bin4}

	for n := 0; n < len(want_values); n++ {
		if math.IsNaN(want_values[n]) {
			if !math.IsNaN(got_values[n]) {
				t.Errorf(
					"want_values[%d] should be NaN but it's %f\n",
					n, got_values[n])
			}
		} else if want_values[n] != got_values[n] {
			t.Errorf(
				"want_values[%d] does not match got_values[%d]: %f != %f",
				n, n, want_values[n], got_values[n])
		}
		if n > 0 {
			prev_want := want_values[n]
			this_recorded := recorded_prev_vals[n-1]
			if !math.IsNaN(prev_want) && prev_want != this_recorded {
				t.Errorf(
					"want_values[%d] != recorded_prev_vals[%d]: %f != %f",
					n, n-1, prev_want, this_recorded)

			}
		}
		if !want_labels[n].Equal(got_labels[n]) {
			t.Errorf(
				"want_labels[%d] does not match labels[%d]: %s != %s",
				n, n, want_labels[n], got_labels[n])
		}
	}
}

func TestMetricNames(t *testing.T) {
	valid_names := []string{
		"some_metric_1",
		"another_metric",
		"good",
		"yezzz_1010101",
		"verylooooooooooOOOOOOOOOOOOOOOOOOOOOOoooooooooong_name",
		"mittari",
		"1231903",
	}
	invalid_names := []string{
		"abc-",
		"'",
		`"`,
		" ",
		";",
		"';delete from sqlite_master where type in ('view', 'table', 'index', 'trigger');",
		"';DROP TABLE BOO;",
		`<a href="badsite.example.com">click here now to win prize</a>`,
		`<script>alert("AAAA")</script>`,
	}

	for n, valid_name := range valid_names {
		t.Run(fmt.Sprintf("%d_%s", n, valid_name), func(t *testing.T) {
			if !is_metric_name_valid(&metric{name: valid_name}) {
				t.Error("should be valid but is not: ", n, valid_name)
			}
		})
	}
	for n, invalid_name := range invalid_names {
		t.Run(fmt.Sprintf("%d_%s", n, invalid_name), func(t *testing.T) {
			if is_metric_name_valid(&metric{name: invalid_name}) {
				t.Error("should be invalid but is not: ", n, invalid_name)
			}
		})
	}
}

func TestDatabaseSmoke(t *testing.T) {
	time_start := time.Now()

	td := t.TempDir()
	db := db_init(filepath.Join(td, "test.db"))
	defer db.Close()
	err := db_migrate(db, test_metrics)
	assert(t, err == nil, "cannot migrate:", err)

	_, err = db.Exec(
		fmt.Sprintf(`INSERT INTO %s (value) VALUES (?)`,
			db_table_name_get(test_metrics[0])),
		123)
	assert(t, err == nil, "cannot insert:", err)

	dps, err := db_datapoints_get(
		db, test_metrics[0], true, 1, 300, time.Duration(1)*time.Second, time_start, time.Now())
	assert(t, err == nil, "cannot get datapoints:", err)

	assert(t, len(dps) == 1, "unexpected amount of datapoints:", len(dps))
	assert(t, almost_equals(dps[0].value, 123), "unexpected value in datapoint:", dps[0].value)

}

func TestMeasureMetric(t *testing.T) {
	// Just in case...
	ctx, cf := context.WithTimeout(context.Background(), 30*time.Second)
	defer cf()
	tc := make(chan db_task)
	go exec_metric(ctx, test_metrics[0], "/bin/sh", 1, tc)
	result := <-tc
	t.Log(result.kind, result.insert_measurement.metric, result.insert_measurement.value)
	if result.kind != DB_TASK_INSERT {
		t.Error("wanted db insertion, got ", result.kind)
	}
	if !almost_equals(result.insert_measurement.value, float64(len("hello world!\n"))) {
		t.Error("unexpected measurement value")
	}
}

func TestParseMetricLine(t *testing.T) {
	badlines := []string{
		"asd|asd",
		"",
		"1",
	}
	for n, badline := range badlines {
		t.Run(fmt.Sprintf("%d_%s", n+1, badline), func(t *testing.T) {
			_, err := config_parse_metric_line(badline)
			if err == nil {
				t.Error("should've failed but did not")
			}
		})
	}

	type goodentry struct {
		line, want_name, want_desc, want_op, want_command string
	}

	goodentries := []goodentry{
		goodentry{
			line:         "something|description here||echo this is command|wc -c",
			want_name:    "something",
			want_desc:    "description here",
			want_op:      "average",
			want_command: "echo this is command|wc -c",
		},
	}
	for n, goodentry := range goodentries {
		t.Run(fmt.Sprintf("%d_%s", n+1, goodentry.line), func(t *testing.T) {
			m, err := config_parse_metric_line(goodentry.line)
			if err != nil {
				t.Error("should've succeeded but did not:", err)
			}
			if m.name != goodentry.want_name {
				t.Error("unexpected name, got ", m.name)
			}
			if m.description != goodentry.want_desc {
				t.Error("unexpected desc, got ", m.description)
			}
			if m.command != goodentry.want_command {
				t.Error("unexpected command, got ", m.command)
			}
		})

	}
}

func TestParseOptions(t *testing.T) {
	y_min := -10.0
	y_max := 20.5
	table := []struct {
		give string
		want graph_options
	}{
		{
			give: "",
			want: graph_options{differentiate: false},
		},
		{
			give: "deriv",
			want: graph_options{differentiate: true},
		},
		{
			give: "kibi",
			want: graph_options{kibi: true},
		},
		{
			give: "kilo,no_ds",
			want: graph_options{kilo: true, no_downsample: true},
		},
		{
			give: "y_min=-10, y_max = 20.5 ",
			want: graph_options{y_min: &y_min, y_max: &y_max},
		},
	}

	for n, entry := range table {
		t.Run(fmt.Sprintf("%d_%s", n+1, entry.give), func(t *testing.T) {
			got, errs := config_parse_metric_options(entry.give)
			if len(errs) > 0 {
				t.Error("should not fail but: ", errs)
			}
			if !reflect.DeepEqual(got, entry.want) {
				t.Errorf(
					"wanted %#v, got %#v",
					entry.want, got)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	b := bytes.NewBufferString(test_config)
	c, err := config_load(b)
	if err != nil {
		t.Fatal(err)
	}

	mc, err := c.parse_measure()
	if err != nil {
		t.Error("measure:", err)
	}

	sc, err := c.parse_serve()
	if err != nil {
		t.Error("serve:", err)
	}

	metrics, err := c.parse_metrics()
	if err != nil {
		t.Error("metrics:", err)
	}
	got_metrics := 0
	for _, m := range metrics {
		switch m.name {
		case "n_temp_files":
			got_metrics |= 1
			assertf(t,
				m.command == `find /tmp/ -type f|wc -l`,
				"unexpected %s command: %s", m.name, m.command)
			assertf(t,
				m.description == "Files in /tmp",
				"unexpected description for %s: %s", m.name, m.description)
			assertf(t,
				*m.options.y_min == float64(0),
				"unexpected y_min for %s: %v", m.name, m.options.y_min)
			assertf(t,
				m.options.y_max == nil,
				"unexpected y_max for %s: %v", m.name, m.options.y_max)
			assertf(t,
				m.options.differentiate == false,
				"unexpected differentiate for %s: %T", m.name, m.options.differentiate)
			assertf(t,
				m.options.kilo == true,
				"unexpected kilo for %s: %v", m.name, m.options.kilo)
			assertf(t,
				m.options.kibi == false,
				"unexpected kibi for %s: %v", m.name, m.options.kibi)

		case "n_processes":
			got_metrics |= 2
			assertf(t,
				m.command == `ps -A|wc -l`,
				"unexpected %s command: %s", m.name, m.command)
			assertf(t,
				almost_equals(*m.options.y_max, float64(1000)),
				"unexpected y_max for %s: %v", m.name, m.options.y_max)

		case "rate_logged_in_users":
			got_metrics |= 4
			assertf(t,
				m.command == `who|wc -l`,
				"unexpected %s command: %s", m.name, m.command)
			assertf(t,
				m.options.differentiate == true,
				"unexpected %s differentiate: %T", m.name, m.options.differentiate)
		case "n_subshell_constant":
			got_metrics |= 8
			assertf(t,
				m.command == `{ echo -n "one"; echo -n two; echo -n three; }|wc -c`,
				"unexpected %s command: %s", m.name, m.command)
		}
	}
	assert(t, got_metrics == (1+2+4+8), "missing some metrics: ", got_metrics)

	assert(t,
		mc.path_db == "/somewhere/db",
		"unexpected measure path_db", mc.path_db)
	assert(t,
		mc.retention_time == time.Duration(21600)*time.Hour,
		"unexpected retention_time", mc.retention_time)
	assert(t,
		mc.prune_db_period == time.Duration(300)*time.Minute,
		"unexpected prune_db_period", mc.prune_db_period)
	assert(t,
		mc.measure_period == time.Duration(150)*time.Second,
		"unexpected measure_period", mc.measure_period)
	assert(t,
		mc.shell == "/bin/zsh",
		"unexpected shell", mc.shell)

	assert(t,
		sc.path_db == "/somewhere/db",
		"unexpected serve path_db", mc.path_db)
	assert(t,
		sc.listen_addr == "localhost:15516",
		"unexpected listen_addr", sc.listen_addr)
	assert(t,
		sc.path_template == "/somewhere/template",
		"unexpected path_template", sc.path_template)
	assert(t,
		sc.default_period == time.Duration(10)*time.Minute,
		"unexpected default_period", sc.default_period)
	assert(t,
		sc.width == 3000,
		"unexpected width", sc.width)
	assert(t,
		sc.height == 1000,
		"unexpected height", sc.height)
	assert(t,
		sc.bin_width == time.Duration(10)*time.Minute,
		"unexpected bin_width", sc.bin_width)
	assert(t,
		sc.max_bins == 1500,
		"unexpected max_bins ", sc.max_bins)
	assert(t,
		sc.autorefresh_period == time.Duration(600)*time.Second,
		"unexpected autorefresh_period", sc.autorefresh_period)
	assert(t,
		sc.graph_format == "png",
		"unexpected graph_format", sc.graph_format)
	assert(t,
		sc.graph_mimetype == "image/png",
		"unexpected graph_mimetype", sc.graph_mimetype)
	assert(t,
		sc.measure_period == time.Duration(150)*time.Second,
		"unexpected serve measure_period", sc.measure_period)
	assert(t,
		time.Duration(sc.downsampling_scale) == 3,
		"unexpected downsampling_scale", sc.downsampling_scale)
}
