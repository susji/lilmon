package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func measure(p *params_measure) {
	db := db_init(p.db_path)
	log.Println("Opening SQLite DB at ", p.db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	metrics := metrics_get()
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

func metrics_get() []*metric {
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
	return metrics
}
