package main

import (
	"math"
	"testing"
	"time"
)

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

	got_values, got_labels := bin_datapoints(dps, int64(bins), ta, tb)
	t.Log(got_values)
	t.Log(got_labels)

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

		if !want_labels[n].Equal(got_labels[n]) {
			t.Errorf(
				"want_labels[%d] does not match labels[%d]: %s != %s",
				n, n, want_labels[n], got_labels[n])
		}
	}
}
