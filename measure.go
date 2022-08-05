package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
)

func measure(p *params_measure) {
	db_path := fmt.Sprintf("%s?_pragma=journal_mode(WAL)", p.db_path)
	log.Println("Opening SQLite DB at ", db_path)
	db := db_init(db_path)
	defer func() {
		if err := db.Close(); err != nil {
			log.Println("warning: error when closing database: ", err)
		}
	}()

	config, err := config_load(p.config_path)
	if err != nil {
		log.Fatal(err)
	}
	metrics, err := config.config_parse_metrics()
	if err != nil {
		log.Fatal("config file reading failed, cannot proceed with measure: ", err)
	}
	mconfig, err := config.config_parse_measure()
	if err != nil {
		log.Fatal(err)
	}

	if err := db_migrate(db, metrics); err != nil {
		log.Fatal("cannot proceed with measure: ", err)
	}

	log.Println("database retention period is ", mconfig.retention_time)

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
	go db_pruner(ctx, ct, metrics, mconfig.retention_time, mconfig.prune_db_period)
	run_metrics(ctx, db, mconfig.measure_period, p.shell, metrics, ct)
}
