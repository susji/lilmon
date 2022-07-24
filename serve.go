package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	RE_TIMERANGE_LAST = regexp.MustCompile(`^last-([0-9]+(\.[0-9]+)?)h$`)
)

func parse_timerange_expr(e string) (start time.Time, end time.Time, err error) {
	e = strings.TrimSpace(e)
	if m := RE_TIMERANGE_LAST.FindStringSubmatch(e); m != nil {
		frac_hours, _ := strconv.ParseFloat(m[1], 64)
		if frac_hours == 0 {
			return start, end, errors.New("hours is zero")
		}
		hours, minutes := math.Modf(frac_hours)
		minutes = 100 * minutes * 60 / 100
		return time.Now().Add(
			-time.Hour*time.Duration(hours) -
				time.Minute*time.Duration(minutes)), time.Now(), nil
	}
	return start, end, errors.New("invalid timerange expression")
}

func serve_index_gen(db *sql.DB, metrics []*metric) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query()
		timerange, ok := v["time"]

		time_start := time.Now().Add(-DEFAULT_GRAPH_PERIOD)
		time_end := time.Now()

		if ok {
			new_time_start, new_time_end, err := parse_timerange_expr(timerange[0])
			if err != nil {
				log.Println("serve_index: bad time: ", err)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad time")
				return
			}
			time_start = new_time_start
			time_end = new_time_end
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
        <a href="/?time=last-1h">hour</a>
        <a href="/?time=last-3h">3 hours</a>
        <a href="/?time=last-12h">12 hours</a>
        <a href="/?time=last-24h">day</a>
        <a href="/?time=last-168h">week</a>
        <a href="/?time=last-720h">month</a>
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
				`%s<img src="/graph?metric=%s&time_start=%d&time_end=%d">`,
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

func serve_graph_gen(db *sql.DB, metrics []*metric) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		v := req.URL.Query()
		metric, ok1 := v["metric"]
		time_start_raw, ok2 := v["time_start"]
		time_end_raw, ok3 := v["time_end"]

		if !ok1 {
			log.Println("serve_graph: metric name missing")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "missing metric name")
			return
		}
		if !is_metric_name_valid(metric[0]) {
			log.Println("serve_graph: metric name invalid: ", metric[0])
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad metric name")
			return
		}

		var time_start, time_end time.Time

		if ok2 {
			time_start_seconds, err := strconv.ParseInt(time_start_raw[0], 10, 64)
			if err != nil {
				log.Println("serve_graph: bad time_start: ", err)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad time_start")
				return
			}
			time_start = time.Unix(time_start_seconds, 0)
		} else {
			time_start = time.Now().Add(-DEFAULT_GRAPH_PERIOD)
		}
		if ok3 {
			time_end_seconds, err := strconv.ParseInt(time_end_raw[0], 10, 64)
			if err != nil {
				log.Println("serve_graph: bad time_end: ", err)
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad time_end")
				return
			}
			time_end = time.Unix(time_end_seconds, 0)
		} else {
			time_end = time.Now()
		}
		if time_start.After(time_end) {
			log.Println("serve_graph: bad time range: ", time_start, time_end)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "bad time range")
			return
		}

		log.Printf(
			"serve_graph: Drawing graph for %q [%s, %s]\n",
			metric[0], time_start, time_end)

		b := bytes.Buffer{}
		if err := graph_generate(db, metric[0], time_start, time_end, &b); err != nil {
			log.Println("serve_graph: PNG encoding failed: ", err)
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

	http.HandleFunc("/", serve_index_gen(db, metrics))
	http.HandleFunc("/graph", serve_graph_gen(db, metrics))
	log.Println("Listening at address ", p.addr)
	http.ListenAndServe(p.addr, nil)
}
