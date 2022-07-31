package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	RE_TIME_RANGE_LAST = regexp.MustCompile(`^last-([0-9]+(\.[0-9]+)?)h$`)
)

func determine_timestamp_format(time_start, time_end time.Time) string {
	var tf string
	dt := time_end.Sub(time_start)
	if dt > time.Hour*24*365 {
		tf = TIMESTAMP_FORMAT_YEAR
	} else if dt > time.Hour*24*30 {
		tf = TIMESTAMP_FORMAT_MONTH
	} else if dt > time.Hour*24 {
		tf = TIMESTAMP_FORMAT_DAY
	} else if dt > time.Minute*15 {
		tf = TIMESTAMP_FORMAT_HOUR
	} else {
		tf = TIMESTAMP_FORMAT_MINUTE
	}
	return tf
}

func serve_index_gen(db *sql.DB, metrics []*metric, label string,
	sconfig *config_serve, template *template.Template) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query()
		raw_time_starts, ok_start := v["time_start"]
		raw_time_ends, ok_end := v["time_end"]

		var time_start, time_end time.Time
		var err_start, err_end error

		if ok_start {
			dur_start, err := time.ParseDuration(raw_time_starts[0])
			if err == nil {
				time_start = time.Now().Add(-dur_start)
			} else {
				err_start = err
			}
		} else {
			time_start = time.Now().Add(-sconfig.default_period)
		}
		if ok_end {
			dur_end, err := time.ParseDuration(raw_time_ends[0])
			if err == nil {
				time_end = time.Now().Add(-dur_end)
			} else {
				err_end = err
			}
		} else {
			time_end = time.Now()
		}
		if err_start != nil || err_end != nil || time_start.After(time_end) {
			log.Println(
				label, ": bad time range:",
				raw_time_starts, raw_time_ends, err_start, err_end)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad time range")
			return
		}

		type MetricData struct {
			Name, Description string
		}
		md := []MetricData{}
		for _, m := range metrics {
			md = append(md, MetricData{Name: m.name, Description: m.description})
		}

		template_data := struct {
			Title                string
			Metrics              []MetricData
			TimeStart, TimeEnd   time.Time
			EpochStart, EpochEnd int64
			RefreshPeriod        time.Duration
			TimeFormat           string
			RenderTime           time.Time
		}{
			Title:         "lilmon",
			RefreshPeriod: sconfig.autorefresh_period,
			Metrics:       md,
			EpochStart:    time_start.Unix(),
			EpochEnd:      time_end.Unix(),
			TimeStart:     time_start,
			TimeEnd:       time_end,
			TimeFormat:    determine_timestamp_format(time_start, time_end),
			RenderTime:    time.Now(),
		}
		template.Execute(w, template_data)
	}
}

func serve_graph_gen(db *sql.DB, metrics []*metric, label string, sconfig *config_serve) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query()
		epoch_starts_raw, ok_start := v["epoch_start"]
		epoch_ends_raw, ok_end := v["epoch_end"]

		if !ok_start || !ok_end {
			log.Println(
				label, ": missing epoch start and/or end:",
				epoch_starts_raw, epoch_ends_raw, ok_start, ok_end)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "missing epoch start and/or end")
			return
		}

		epoch_start, err_start := strconv.ParseInt(epoch_starts_raw[0], 10, 64)
		epoch_end, err_end := strconv.ParseInt(epoch_ends_raw[0], 10, 64)

		if err_start != nil || err_end != nil || epoch_start > epoch_end {
			log.Println(label, ": bad epoch range", err_start, err_end)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad epoch range")
			return
		}

		metric_names, ok_metric := v["metric"]
		if !ok_metric {
			log.Println(label, ": metric name missing")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "missing metric name")
			return
		}
		metric := metric_find(metrics, metric_names[0])
		if !is_metric_name_valid(metric_names[0]) || metric == nil {
			log.Println(label, ": metric name invalid: ", metric_names[0])
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad metric name")
			return
		}

		if epoch_start >= epoch_end {
			log.Println(label, ": epoch_start >= epoch_end")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "epoch_start >= epoch_end")
			return
		}

		time_start := time.Unix(epoch_start, 0)
		time_end := time.Unix(epoch_end, 0)
		log.Printf(
			label+": Drawing graph for %q [%s, %s]\n",
			metric_names[0], time_start, time_end)

		b := bytes.Buffer{}
		if err := graph_generate(db, metric, time_start, time_end, &b, sconfig); err != nil {
			log.Println(label, ": PNG encoding failed: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "graph generation failed")
			return
		}
		gb := b.Bytes()
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(gb)))
		w.WriteHeader(http.StatusOK)
		w.Write(gb)
	}
}

func serve(p *params_serve) {
	template := template.Must(template.ParseFiles(p.template_path))

	config, err := config_load(p.config_path)
	if err != nil {
		log.Fatal(err)
	}
	metrics, err := config.config_parse_metrics()
	if err != nil {
		log.Fatal("config file reading failed, cannot proceed with serve: ", err)
	}
	sconfig, err := config.config_parse_serve()
	if err != nil {
		log.Fatal("parsing serve config failed: ", err)
	}

	db_path := fmt.Sprintf("%s?mode=ro", p.db_path)
	log.Println("Opening SQLite DB at ", db_path)
	db := db_init(db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	http.HandleFunc("/", serve_index_gen(db, metrics, "index", sconfig, template))
	http.HandleFunc("/graph", serve_graph_gen(db, metrics, "graph", sconfig))
	log.Println("Listening at address ", p.addr)
	http.ListenAndServe(p.addr, nil)
}
