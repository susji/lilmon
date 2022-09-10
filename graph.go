package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func op_identity(i int, vals []float64, _ []time.Time) float64 {
	return vals[i]
}

func op_derivative(i int, vals []float64, times []time.Time) float64 {
	if i == 0 {
		return math.NaN()
	}
	// We implement the simple two-point forward difference approximation
	// for the derivative.

	// We go back as far as necessary to find a previous data point for the
	// derivative. Error increases relative to the distance, but it's better
	// than spitting out NaNs...
	previ := i - 1
	for previ >= 0 && math.IsNaN(vals[previ]) {
		previ--
	}
	if previ == -1 {
		return math.NaN()
	}

	seconds_now := times[i].Unix()
	seconds_prev := times[previ].Unix()
	delta_t := seconds_now - seconds_prev
	val_now := vals[i]
	val_prev := vals[previ]
	delta_v := val_now - val_prev
	dv := delta_v / float64(delta_t)
	return dv
}

func bin_datapoints(dps []datapoint, bins int64, time_start, time_end time.Time, op bin_op) (
	[]float64, []time.Time, float64, float64) {

	if time_start.After(time_end) {
		panic(
			fmt.Sprintf(
				"This is a bug: time_start > time_end: %s > %s",
				time_start, time_end))
	}
	// We assume datapoints are sorted in ascending timestamp order.
	binned := make([]float64, bins, bins)
	result := make([]float64, bins, bins)
	labels := make([]time.Time, bins, bins)
	time_start_epoch := time_start.Unix()
	time_end_epoch := time_end.Unix()
	delta_t_bin_sec := (time_end_epoch - time_start_epoch) / bins
	//
	// First we place values into their bins:
	//
	// time_start   bin1               bin2               bin3       time_end
	//      .------------------+------------------+------------------.
	//      | ts1   ts2   ts3  |               ts4| ts5              |
	//      '------------------+------------------+------------------.
	//
	// ... and afterwards the take an average of the in-bin-value.
	//
	cur_dp_i := 0
	ts_bin_left_sec := time_start_epoch
	ts_bin_right_sec := time_start_epoch + delta_t_bin_sec
	val_min := math.NaN()
	val_max := math.NaN()
	// NaN through each bin once and see how many timestamps we can fit in.
	for cur_bin := int64(0); cur_bin < bins; cur_bin++ {
		bin_value_sum_cur := float64(0)
		n_datapoints_cur := 0
		// First sum together datapoints belonging to this bin...
		for cur_dp_i < len(dps) {
			dp_sec := dps[cur_dp_i].ts.Unix()
			if dp_sec >= ts_bin_left_sec && dp_sec <= ts_bin_right_sec {
				bin_value_sum_cur += dps[cur_dp_i].value
				n_datapoints_cur++
				cur_dp_i++
			} else {
				break
			}
		}
		// ... and then do the requested operation - usually average of
		// the current bin values.
		binned[cur_bin] = bin_value_sum_cur / float64(n_datapoints_cur)

		// Timestamp label is the average of bin left and right.
		labels[cur_bin] = time.Unix((ts_bin_left_sec+ts_bin_right_sec)/2, 0)

		// Do bin value transform before we interpret the value.
		result[cur_bin] = op(
			int(cur_bin),
			binned,
			labels)

		// Keep track of value min and max.
		if (!math.IsNaN(val_min) && result[cur_bin] < val_min) ||
			(math.IsNaN(val_min) && n_datapoints_cur > 0) {
			val_min = result[cur_bin]
		}
		if (!math.IsNaN(val_max) && result[cur_bin] > val_max) ||
			(math.IsNaN(val_max) && n_datapoints_cur > 0) {
			val_max = result[cur_bin]
		}

		// Slide bin timestamps over the next bin.
		ts_bin_left_sec += delta_t_bin_sec
		ts_bin_right_sec += delta_t_bin_sec
	}
	return result, labels, val_min, val_max
}

func val_to_unit_prefix_base10(v float64) (bool, float64, string) {
	var d int64
	var s string

	switch {
	case v >= 1000*1000*1000*1000*1000*1000:
		d = 1000 * 1000 * 1000 * 1000 * 1000 * 1000
		s = "E"
	case v >= 1000*1000*1000*1000*1000:
		d = 1000 * 1000 * 1000 * 1000 * 1000
		s = "P"
	case v >= 1000*1000*1000*1000:
		d = 1000 * 1000 * 1000 * 1000
		s = "T"
	case v >= 1000*1000*1000:
		d = 1000 * 1000 * 1000
		s = "G"
	case v >= 1000*1000:
		d = 1000 * 1000
		s = "M"
	case v >= 1000:
		d = 1000
		s = "k"
	default:
		return false, v, ""
	}
	return true, v / float64(d), s
}

func val_to_unit_prefix_base2(v float64) (bool, float64, string) {
	var d int64
	var s string

	switch {
	case v >= 1024*1024*1024*1024*1024*1024:
		d = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
		s = "Ei"
	case v >= 1024*1024*1024*1024*1024:
		d = 1024 * 1024 * 1024 * 1024 * 1024
		s = "Pi"
	case v >= 1024*1024*1024*1024:
		d = 1024 * 1024 * 1024 * 1024
		s = "Ti"
	case v >= 1024*1024*1024:
		d = 1024 * 1024 * 1024
		s = "Gi"
	case v >= 1024*1024:
		d = 1024 * 1024
		s = "Mi"
	case v >= 1024:
		d = 1024
		s = "Ki"
	default:
		return false, v, ""
	}
	return true, v / float64(d), s
}

func val_format_for_printing(v float64) string {
	return strconv.FormatFloat(v, 'g', 3, 64)
}

type TransformerTicker struct {
	ValueTransformer func(float64) (bool, float64, string)
}

func (t TransformerTicker) Ticks(min, max float64) []plot.Tick {
	got := plot.DefaultTicks{}.Ticks(min, max)
	for i := range got {
		if got[i].Label == "" {
			continue
		}
		v := got[i].Value
		changed, vt, s := t.ValueTransformer(v)
		if !changed {
			continue
		}
		got[i].Label = val_format_for_printing(vt)
		if len(s) > 0 {
			got[i].Label += " " + s
		}
	}
	return got
}

type NeatFloatTicker struct{}

func (NeatFloatTicker) Ticks(min, max float64) []plot.Tick {
	got := plot.DefaultTicks{}.Ticks(min, max)
	for i := range got {
		if got[i].Label == "" {
			continue
		}
		got[i].Label = val_format_for_printing(got[i].Value)
	}
	return got
}

func graph_generate(db *sql.DB, metric *metric, force_no_ds bool,
	time_start, time_end time.Time, w io.Writer, sconfig *config_serve) error {
	// To have sensible graphs, the bin width (delta-t) should be
	//   - equal or greater than our measurement period and
	//   - smaller than the amount of horizontal pixels divided by some
	//     small coefficient..
	bins := int(time_end.Sub(time_start) / sconfig.bin_width)
	if bins > sconfig.max_bins {
		bins = sconfig.max_bins
	}
	if bins == 0 {
		return errors.New("cannot graph zero bins")
	}

	t0 := time.Now()

	dps, err := db_datapoints_get(
		db, metric, force_no_ds, sconfig.downsampling_scale, bins,
		sconfig.measure_period, time_start, time_end)
	if err != nil {
		log.Println("graph_generate: error from DB get: ", err)
		return err
	}

	t1 := time.Now()

	op := op_identity
	if metric.options.differentiate {
		op = op_derivative
	}
	// Heavy lifting: obtain the binned data.
	binned, labels, _, _ := bin_datapoints(
		dps, int64(bins), time_start, time_end, op)

	xys := plotter.XYs{}
	for i := 0; i < len(binned); i++ {
		if math.IsNaN(binned[i]) {
			continue
		}
		xys = append(xys, plotter.XY{X: float64(labels[i].Unix()), Y: binned[i]})
	}

	s, err := NewScatterBars(xys)
	if err != nil {
		return err
	}
	s.GlyphStyle.Color = COLOR_FILL
	s.GlyphStyle.Radius = vg.Length(sconfig.glyph_size)
	s.LineStyle.Color = COLOR_FILL
	s.LineStyle.Width = vg.Length(sconfig.line_thickness)

	p := plot.New()
	p.Add(s)
	p.Add(plotter.NewGrid())
	p.BackgroundColor = COLOR_BG
	p.X.LineStyle.Color = COLOR_LABEL
	p.Y.LineStyle.Color = COLOR_LABEL
	p.Title.Text = ""
	p.X.Label.Text = ""
	p.Y.Label.Text = ""
	p.X.Tick.Marker = plot.TimeTicks{
		Format: determine_timestamp_format(time_start, time_end),
		Time: func(t float64) time.Time {
			return time.Unix(int64(t), 0)
		},
	}

	var y_ticker plot.Ticker
	switch {
	case metric.options.kilo:
		y_ticker = TransformerTicker{ValueTransformer: val_to_unit_prefix_base10}
	case metric.options.kibi:
		y_ticker = TransformerTicker{ValueTransformer: val_to_unit_prefix_base2}
	default:
		y_ticker = NeatFloatTicker{}
	}
	p.Y.Tick.Marker = y_ticker

	p.X.Min = float64(time_start.Unix())
	p.X.Max = float64(time_end.Unix())

	if metric.options.y_min != nil {
		p.Y.Min = *metric.options.y_min
	}
	if metric.options.y_max != nil {
		p.Y.Max = *metric.options.y_max
	}

	wt, err := p.WriterTo(
		vg.Length(sconfig.width), vg.Length(sconfig.height), sconfig.graph_format)
	if err != nil {
		return err
	}
	_, err = wt.WriteTo(w)
	if err != nil {
		return err
	}

	t2 := time.Now()

	log.Println("t1-t0=", t1.Sub(t0), ", t2-t1=", t2.Sub(t1), ", t2-t0=", t2.Sub(t0))

	return nil
}
