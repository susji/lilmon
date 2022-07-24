package main

import (
	"database/sql"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"strconv"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func bin_datapoints(dps []datapoint, bins int64, time_start, time_end time.Time) (
	[]float64, []time.Time, float64, float64) {

	if time_start.After(time_end) {
		panic(
			fmt.Sprintf(
				"This is a bug: time_start > time_end: %s > %s",
				time_start, time_end))
	}
	// We assume datapoints are sorted in ascending timestamp order.
	binned := make([]float64, bins, bins)
	labels := make([]time.Time, bins, bins)
	time_start_epoch := time_start.Unix()
	time_end_epoch := time_end.Unix()
	delta_t_bin_sec := (time_end_epoch - time_start_epoch) / bins
	//
	// Just histogramming, for three bins it would look like:
	//
	// time_start   bin1               bin2               bin3       time_end
	//      .------------------+------------------+------------------.
	//      | ts1   ts2   ts3  |               ts4| ts5              |
	//      '------------------+------------------+------------------.
	//
	cur_dp_i := 0
	ts_bin_left_sec := time_start_epoch
	ts_bin_right_sec := time_start_epoch + delta_t_bin_sec
	val_min := math.NaN()
	val_max := math.NaN()
	// Loop through each bin once and see how many timestamps we can fit in.
	for cur_bin := int64(0); cur_bin < bins; cur_bin++ {
		bin_value_sum := float64(0)
		datapoints_in_bin := 0
		// First sum together datapoints belonging to this bin...
		for cur_dp_i < len(dps) {
			dp_sec := dps[cur_dp_i].ts.Unix()
			if dp_sec >= ts_bin_left_sec && dp_sec <= ts_bin_right_sec {
				bin_value_sum += dps[cur_dp_i].value
				datapoints_in_bin++
				cur_dp_i++
			} else {
				break
			}
		}
		// ... and then figure out the average value, and store it.
		binned[cur_bin] = bin_value_sum / float64(datapoints_in_bin)

		// Keep track of value min and max.
		if (!math.IsNaN(val_min) && binned[cur_bin] < val_min) ||
			(math.IsNaN(val_min) && datapoints_in_bin > 0) {
			val_min = binned[cur_bin]
		}
		if (!math.IsNaN(val_max) && binned[cur_bin] > val_max) ||
			(math.IsNaN(val_max) && datapoints_in_bin > 0) {
			val_max = binned[cur_bin]
		}

		// Timestamp label is the average of bin left and right.
		labels[cur_bin] = time.Unix((ts_bin_left_sec+ts_bin_right_sec)/2, 0)

		// Slide bin timestamps over the next bin.
		ts_bin_left_sec += delta_t_bin_sec
		ts_bin_right_sec += delta_t_bin_sec
	}
	return binned, labels, val_min, val_max
}

func graph_draw(values []float64, labels []time.Time, val_min, val_max float64) image.Image {
	total_w := DEFAULT_GRAPH_WIDTH + DEFAULT_GRAPH_PAD_WIDTH_LEFT + DEFAULT_GRAPH_PAD_WIDTH_RIGHT
	total_h := DEFAULT_GRAPH_HEIGHT + DEFAULT_GRAPH_PAD_HEIGHT_UP + DEFAULT_GRAPH_PAD_HEIGHT_DOWN
	pad_w := DEFAULT_GRAPH_PAD_WIDTH_LEFT + DEFAULT_GRAPH_PAD_WIDTH_RIGHT
	pad_h := DEFAULT_GRAPH_PAD_HEIGHT_UP + DEFAULT_GRAPH_PAD_HEIGHT_DOWN
	w := total_w - pad_w
	h := total_h - pad_h
	bin_w := w / len(values)
	g := image.NewRGBA(image.Rect(0, 0, total_w, total_h))
	draw.Draw(g, g.Bounds(), &image.Uniform{COLOR_BG}, image.Point{}, draw.Src)

	marker_halfwidth := bin_w / 4
	cur_x := DEFAULT_GRAPH_PAD_WIDTH_LEFT + bin_w/2 - bin_w
	for bin := 0; bin < len(values); bin++ {
		cur_x += bin_w
		if math.IsNaN(values[bin]) {
			continue
		}
		// do Y calculations in zero reference, that is, normalize Y values as [0, 1].
		norm_y := (values[bin] - val_min) / (val_max - val_min)
		cur_y := DEFAULT_GRAPH_PAD_HEIGHT_UP + math.Floor(float64(h)-float64(h)*norm_y)
		marker := image.Rect(
			cur_x-marker_halfwidth, int(cur_y)-marker_halfwidth,
			cur_x+marker_halfwidth, int(cur_y)+marker_halfwidth)
		draw.Draw(g, marker, &image.Uniform{COLOR_FG}, image.Point{}, draw.Src)

	}

	label_max := strconv.FormatFloat(val_max, 'g', -1, 64)
	label_min := strconv.FormatFloat(val_min, 'g', -1, 64)
	graph_label(g, total_w-DEFAULT_GRAPH_PAD_WIDTH_RIGHT,
		DEFAULT_GRAPH_PAD_HEIGHT_UP+DEFAULT_LABEL_MAX_Y0, label_max)
	graph_label(g, total_w-DEFAULT_GRAPH_PAD_WIDTH_RIGHT,
		total_h-DEFAULT_GRAPH_PAD_HEIGHT_DOWN, label_min)

	label_start := labels[0].Format(TIMESTAMP_FORMAT)
	label_end := labels[len(labels)-1].Format(TIMESTAMP_FORMAT)
	graph_label(g, 0, total_h, label_start)
	graph_label(g, total_w-DEFAULT_LABEL_SHIFT_X, total_h, label_end)

	return g
}

func graph_label(img *image.RGBA, x, y int, label string) {
	// https://stackoverflow.com/a/38300583
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(COLOR_LABEL),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func graph_generate(db *sql.DB, metric string, time_start, time_end time.Time, w io.Writer) error {
	dps, err := db_datapoints_get(db, metric, time_start, time_end)
	if err != nil {
		log.Println("graph_generate: error from DB get: ", err)
		return err
	}
	binned, labels, val_min, val_max := bin_datapoints(
		dps, DEFAULT_GRAPH_BINS, time_start, time_end)
	g := graph_draw(binned, labels, val_min, val_max)
	if err := png.Encode(w, g); err != nil {
		return err
	}
	return nil
}