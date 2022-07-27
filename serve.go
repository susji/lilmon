package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	RE_TIME_RANGE_LAST = regexp.MustCompile(`^last-([0-9]+(\.[0-9]+)?)h$`)
)

func serve_index_gen(db *sql.DB, metrics []*metric, label string) http.HandlerFunc {
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
			time_start = time.Now().Add(-DEFAULT_GRAPH_PERIOD)
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

		fmt.Fprintf(w, `
<html>
  <head>
    <meta http-equiv="refresh" content="%d">
  </head>
  <body>
    <div>
      <code>
        Show last
        <a href="/?time_start=30m">30 minutes</a>
        <a href="/?time_start=1h">hour</a>
        <a href="/?time_start=3h">3 hours</a>
        <a href="/?time_start=6h">6 hours</a>
        <a href="/?time_start=12h">12 hours</a>
        <a href="/?time_start=24h">day</a>
        <a href="/?time_start=72h">3 days</a>
        <a href="/?time_start=168h">week</a>
        <a href="/?time_start=720h">month</a>
      </code>
    </div>
`, DEFAULT_REFRESH_PERIOD)
		// XXXX Do proper html templating here
		indent := `    `
		for n, m := range metrics {
			fmt.Fprintln(w, indent, "<div>")
			fmt.Fprintln(
				w,
				indent, "<pre>", n, ": ",
				m.name, ": ",
				m.description, "</pre>")
			fmt.Fprintf(
				w,
				`%s<img src="/graph?metric=%s&epoch_start=%d&epoch_end=%d">`,
				indent,
				m.name, time_start.Unix(), time_end.Unix())
			fmt.Fprintln(w)
			fmt.Fprintln(w, indent, "</div>")
		}
		fmt.Fprintf(w, `
    <hr>
    <pre>lilmon</pre>
    <pre>%s (autorefresh @ %d sec)</pre>
  </body>
</html>
`, time.Now().Format(TIMESTAMP_FORMAT), DEFAULT_REFRESH_PERIOD)
	}
}

func serve_graph_gen(db *sql.DB, metrics []*metric, label string) http.HandlerFunc {
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

		time_start := time.Unix(epoch_start, 0)
		time_end := time.Unix(epoch_end, 0)
		log.Printf(
			label+": Drawing graph for %q [%s, %s]\n",
			metric_names[0], time_start, time_end)

		b := bytes.Buffer{}
		if err := graph_generate(db, metric, time_start, time_end, &b); err != nil {
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
	db_path := fmt.Sprintf("%s?mode=ro", p.db_path)
	log.Println("Opening SQLite DB at ", db_path)
	db := db_init(db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()
	metrics, err := metrics_load(p.config_path)
	if err != nil {
		log.Fatal("config file reading failed, cannot proceed with serve")
	}
	if err := db_migrate(db, metrics); err != nil {
		log.Fatal("cannot proceed with serve: ", err)
	}

	http.HandleFunc("/", serve_index_gen(db, metrics, "index"))
	http.HandleFunc("/graph", serve_graph_gen(db, metrics, "graph"))
	log.Println("Listening at address ", p.addr)
	http.ListenAndServe(p.addr, nil)
}
