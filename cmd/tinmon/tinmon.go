package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

const (
	FLAG_DB_PATH    = "db-path"
	DEFAULT_DB_PATH = "/tmp/tinmon.sqlite"
	HELP_DB_PATH    = "Filepath to tinmon SQLite database"

	FLAG_SHELL    = "shell"
	DEFAULT_SHELL = "/bin/sh"
	HELP_SHELL    = "Filepath for shell to use when measuring metrics"

	FLAG_PERIOD    = "period"
	DEFAULT_PERIOD = 20 * time.Second
	HELP_PERIOD    = "How often to take new measurements"
)

type params_measure struct {
	db_path, shell string
	period         time.Duration
}

type params_serve struct {
	db_path string
}

type metric struct {
	name, description, command string
}

func db_init(db_path string) *sql.DB {
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		log.Fatal("cannot open database: %v", err)
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
CREATE TABLE IF NOT EXISTS tinmon_metric_%s(
    id INTEGER PRIMARY KEY,
    value DOUBLE PRECISION,
    timestamp DATETIME_DEFAULT CURRENT_TIMESTAMP
);
`
	in_err := false
	for n, m := range metrics {
		log.Printf(
			"Maybe creating table metric %d/%d: %s (%s)\n",
			n+1, len(metrics), m.name, m.description)
		// XXX Make sure metric name is suitable for query expansion
		_, err := db.Query(fmt.Sprintf(template_table, m.name))
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

func run_metrics(ctx context.Context, period time.Duration, shell string, metrics []*metric) {
	log.Println("Entering measurement loop with period of ", period, "...")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(period):
			for n, m := range metrics {
				n := n
				sctx, cf := context.WithTimeout(ctx, period/2+1)
				defer cf()
				go func(sctx context.Context) {
					log.Printf(
						"Running command %d/%d: %q\n",
						n+1, len(metrics), m.command)
					cmd := exec.CommandContext(sctx, shell, "-c", m.command)
					out, err := cmd.Output()
					if err != nil {
						log.Println("... run failed: ", err)
						return
					}
					cleaned := strings.TrimSpace(string(out))
					log.Printf("... run worked and returned: %q\n", cleaned)
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			cf()
			fmt.Println("got SIGINT -- bailing")
		}
	}()
	run_metrics(ctx, time.Second*15, p.shell, metrics)
}

func serve(p *params_serve) {

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
