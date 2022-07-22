package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	FLAG_DB_PATH    = "db-path"
	DEFAULT_DB_PATH = "/tmp/lilmon.sqlite"
	HELP_DB_PATH    = "Filepath to lilmon SQLite database"

	FLAG_SHELL    = "shell"
	DEFAULT_SHELL = "/bin/sh"
	HELP_SHELL    = "Filepath for shell to use when measuring metrics"

	FLAG_PERIOD    = "period"
	DEFAULT_PERIOD = 20 * time.Second
	HELP_PERIOD    = "How often to take new measurements"

	FLAG_ADDR    = "addr"
	DEFAULT_ADDR = "localhost:15515"
	HELP_ADDR    = "HTTP listening address"
)

const (
	DEFAULT_RETENTION_TIME = 7 * 24 * time.Hour
)

type params_measure struct {
	db_path, shell string
	period         time.Duration
}

type params_serve struct {
	db_path string
	addr    string
}

type metric struct {
	name, description, command string
}

type measurement struct {
	metric *metric
	value  float64
}

const (
	DB_TASK_PRUNE_TABLE = iota
	DB_TASK_INSERT
)

type db_task struct {
	kind int

	prune_metric_name      string
	prune_retention_period time.Duration

	insert_measurement *measurement
}

func db_init(db_path string) *sql.DB {
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		log.Fatalf("cannot open database: %v\n", err)
	}

	var db_version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&db_version); err != nil {
		log.Println("warning: unable to get sqlite version: ", err)
	} else {
		log.Println("database version: ", db_version)

	}
	return db
}

func db_migrate(db *sql.DB, metrics []*metric) error {
	template_table := `
CREATE TABLE IF NOT EXISTS lilmon_metric_%s (
    id INTEGER PRIMARY KEY,
    value DOUBLE PRECISION,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP);
CREATE INDEX IF NOT EXISTS index_lilmon_metric_%s
    ON lilmon_metric_%s (value, timestamp);
`
	in_err := false
	for n, m := range metrics {
		log.Printf(
			"Maybe creating tables and indices for metric %d/%d: %s (%s)\n",
			n+1, len(metrics), m.name, m.description)

		_, err := db.Query(fmt.Sprintf(template_table, m.name, m.name, m.name))
		if err != nil {
			log.Printf("failed to create table for metric %s: %v ", m.name, err)
			in_err = true
		}
	}
	if in_err {
		return errors.New("database migration encountered errors")
	}
	return nil
}

func db_writer(ctx context.Context, db *sql.DB, tasks <-chan db_task) {
	template_insert := `INSERT INTO lilmon_metric_%s (value) VALUES (?)`
	template_prune := `DELETE FROM lilmon_metric_%s WHERE timestamp < DATETIME('now', '-%d seconds')`
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-tasks:
			switch task.kind {
			case DB_TASK_INSERT:
				metric := task.insert_measurement.metric.name
				value := task.insert_measurement.value
				_, err := db.ExecContext(
					ctx,
					fmt.Sprintf(template_insert, metric),
					value)
				if err != nil {
					log.Printf(
						"metric insert failed for %s with value %f: %v\n",
						metric, value, err)
				}
			case DB_TASK_PRUNE_TABLE:
				metric := task.prune_metric_name
				retention_period := task.prune_retention_period

				log.Printf(
					"Pruning metric %s for older than %s entries.\n",
					metric, retention_period)
				q := fmt.Sprintf(
					template_prune,
					metric,
					int64(retention_period/time.Second))
				_, err := db.ExecContext(ctx, q)
				if err != nil {
					log.Println("Pruning failed: ", err)
				}
			default:
				panic(fmt.Sprintf("This is a bug: db_task.kind == %d", task.kind))
			}
		}
	}
}

func run_metrics(ctx context.Context, db *sql.DB, period time.Duration, shell string,
	metrics []*metric, tasks chan<- db_task) {

	log.Println("Entering measurement loop with period of ", period, "...")
	_cur := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(period):
			for _n, _m := range metrics {
				_cur++
				cur := _cur
				n := _n
				m := _m
				sctx, cf := context.WithTimeout(ctx, period/2+1)
				defer cf()
				go func(sctx context.Context) {
					log.Printf(
						"{%d} Running command %d/%d: %q\n",
						cur, n+1, len(metrics), m.command)
					cmd := exec.CommandContext(sctx, shell, "-c", m.command)
					out, err := cmd.Output()
					if err != nil {
						log.Printf(
							"{%d}... run failed: %v\n",
							cur, err)
						return
					}
					cleaned := strings.TrimSpace(string(out))
					log.Printf(
						"{%d}... run worked and returned: %q\n",
						cur, cleaned)

					val, err := strconv.ParseFloat(cleaned, 64)
					if err != nil {
						log.Printf(
							"{%d}... but it's not floaty: %v\n",
							cur, err)
						return
					}

					tasks <- db_task{
						kind: DB_TASK_INSERT,
						insert_measurement: &measurement{
							metric: m,
							value:  val,
						}}

				}(sctx)
			}
		}
	}
}

func validate_metrics(metrics []*metric) error {
	in_err := false
	re_name := regexp.MustCompile("^[-_a-zA-Z0-9]{1,512}$")
	for n, m := range metrics {
		log.Printf("Validating metric name %d/%d: %s\n", n+1, len(metrics), m.name)
		if !re_name.MatchString(m.name) {
			log.Println("... and the name is not valid.")
			in_err = true
		}
	}
	if in_err {
		return errors.New("one or more metrics did not validate")
	}
	return nil
}

func db_pruner(ctx context.Context, tasks chan<- db_task, metrics []*metric, retention_period time.Duration) {
	PRUNE_PERIOD := 15 * time.Second
	log.Println("Entering pruning loop with period of ", PRUNE_PERIOD)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(PRUNE_PERIOD):
			for _, m := range metrics {
				tasks <- db_task{
					kind:                   DB_TASK_PRUNE_TABLE,
					prune_metric_name:      m.name,
					prune_retention_period: retention_period,
				}
			}
		}
	}
}

func measure(p *params_measure) {
	db := db_init(p.db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	metrics := []*metric{
		&metric{
			name:        "n_temp_files",
			description: "count of files under /tmp",
			command:     "find /tmp/ -type f | wc -l",
		},
		&metric{
			name:        "n_memory_pages_used",
			description: "pages of memory in use",
			command:     "vm_stat |fgrep 'Pages active:'|cut -d ':' -f 2|cut -d '.' -f1",
		},
	}
	if err := validate_metrics(metrics); err != nil {
		log.Fatal("cannot proceed with measure: ", err)
	}
	if err := db_migrate(db, metrics); err != nil {
		log.Fatal("cannot proceed with measure: ", err)
	}

	ctx, cf := context.WithCancel(context.Background())

	ci := make(chan os.Signal, 1)
	signal.Notify(ci, os.Interrupt)
	go func() {
		for range ci {
			cf()
			fmt.Println("got SIGINT -- bailing")
		}
	}()

	ct := make(chan db_task)
	go db_writer(ctx, db, ct)
	go db_pruner(ctx, ct, metrics, DEFAULT_RETENTION_TIME)
	run_metrics(ctx, db, time.Second*15, p.shell, metrics, ct)
}

func serve_index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "<pre>lilmon</pre>")
}

func serve_graph(w http.ResponseWriter, req *http.Request) {
	gb := []byte{}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(gb)
}

func serve(p *params_serve) {
	http.HandleFunc("/", serve_index)
	http.HandleFunc("/graph", serve_graph)
	log.Println("Listening at address ", p.addr)
	http.ListenAndServe(p.addr, nil)
}

func main() {
	var p_measure params_measure
	var p_serve params_serve

	if len(os.Args) <= 1 {
		fmt.Printf("usage: %s [subcommand]\n", filepath.Base(os.Args[0]))
		fmt.Println("subcommand is either `measure', `serve', or `help'`.")
		os.Exit(1)
	}

	cmd_measure := flag.NewFlagSet("measure", flag.ExitOnError)
	cmd_measure.StringVar(&p_measure.db_path, FLAG_DB_PATH, DEFAULT_DB_PATH, HELP_DB_PATH)
	cmd_measure.StringVar(&p_measure.shell, FLAG_SHELL, DEFAULT_SHELL, HELP_SHELL)
	cmd_measure.DurationVar(&p_measure.period, FLAG_PERIOD, DEFAULT_PERIOD, HELP_PERIOD)

	cmd_serve := flag.NewFlagSet("serve", flag.ExitOnError)
	cmd_serve.StringVar(&p_serve.db_path, FLAG_DB_PATH, DEFAULT_DB_PATH, HELP_DB_PATH)
	cmd_serve.StringVar(&p_serve.addr, FLAG_ADDR, DEFAULT_ADDR, HELP_ADDR)

	switch os.Args[1] {
	case "measure":
		cmd_measure.Parse(os.Args[2:])
		measure(&p_measure)
	case "serve":
		cmd_serve.Parse(os.Args[2:])
		serve(&p_serve)
	case "help":
		fmt.Println("The subcommands are:")
		fmt.Println()
		fmt.Println("    measure          measure metrics until interrupted")
		fmt.Println("    serve            display measurements via HTTP")
		fmt.Println("    help             show this help")
		fmt.Println()
		os.Exit(0)
	default:
		fmt.Println("unknown subcommand: ", os.Args[1])
		os.Exit(2)
	}
}
