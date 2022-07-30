package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
)

func measure(p *params_measure) {
	db := db_init(p.db_path)
	log.Println("Opening SQLite DB at ", p.db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	metrics, err := metrics_load(p.config_path)
	if err != nil {
		log.Fatal("config file reading failed, cannot proceed with measure")
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
	run_metrics(ctx, db, DEFAULT_MEASUREMENT_PERIOD, p.shell, metrics, ct)
}
