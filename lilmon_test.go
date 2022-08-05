package main

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"
)

var test_metrics = []*metric{
	&metric{
		name:        "test_metric_1",
		description: "a simple test metric",
		command:     `echo hello world!|wc -c`,
	},
}

func almost_equals(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestBinDatapoints(t *testing.T) {
	ta, _ := time.Parse(time.RFC3339, "2020-01-01T12:00:00Z")
	tb, _ := time.Parse(time.RFC3339, "2020-01-01T13:00:00Z")

	dps := []datapoint{
		// FIRST BIN BEGINS
		datapoint{
			ts:    ta.Add(5 * time.Minute),
			value: 10,
		},
		datapoint{
			ts:    ta.Add(10 * time.Minute),
			value: 20,
		},
		datapoint{
			ts:    ta.Add(10 * time.Minute),
			value: 0,
		},
		// SECOND BIN BEGINS
		datapoint{
			ts:    ta.Add(16 * time.Minute),
			value: 100,
		},
		// THIRD BIN BEGINS
		// FOURTH BIN BEGINS
		datapoint{
			ts:    ta.Add(57 * time.Minute),
			value: -10,
		},
		datapoint{
			ts:    ta.Add(59*time.Minute + 59*time.Second),
			value: 20,
		},

		// PAST LAST BIN
		datapoint{
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
	db := db_init(":memory:")
	defer db.Close()
	db_migrate(db, test_metrics)
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
