package main

import (
	"database/sql"
	"flag"
	"log"
	"os/exec"
	"strings"

	_ "github.com/glebarez/go-sqlite"
)

type metric struct {
	command     string
	description string
}

func db_init(db_path string) *sql.DB {
	db, err := sql.Open("sqlite", db_path)
	if err != nil {
		log.Fatal("cannot open database: %v", err)
	}

	var db_version string
	if err := db.QueryRow("select sqlite_version()").Scan(&db_version); err != nil {
		log.Println("warning: unable to get sqlite version: ", err)
	} else {
		log.Println("database version: ", db_version)
	}
	return db
}

func run_metrics(shell string, metrics []*metric) error {
	for n, m := range metrics {
		log.Printf("Running command %d/%d: %q\n", n+1, len(metrics), m.command)
		cmd := exec.Command(shell, "-c", m.command)
		out, err := cmd.Output()
		if err != nil {
			log.Println("... run failed: ", err)
			continue
		}
		cleaned := strings.TrimSpace(string(out))
		log.Printf("... run worked and returned: %q\n", cleaned)
	}
	return nil
}

func main() {
	var db_path string
	var shell string

	flag.StringVar(&db_path, "db-path", "/tmp/tinmon.sqlite", "Path where to store tinmon database file")
	flag.StringVar(&shell, "shell", "/bin/sh", "Shell to invoke when obtaining metrics")
	flag.Parse()

	db := db_init(db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	metrics := []*metric{
		&metric{
			command:     "find /tmp/ -type f | wc -l",
			description: "count of files under /tmp",
		},
		&metric{
			command:     "vm_stat |fgrep 'Pages active:'|cut -d ':' -f 2|cut -d '.' -f1",
			description: "pages of memory in use",
		},
	}

	run_metrics(shell, metrics)

}
