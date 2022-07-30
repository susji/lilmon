package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func db_datapoints_get(db *sql.DB, metric string, time_start, time_end time.Time) (
	[]datapoint, error) {

	template_select_values := `
SELECT timestamp, value FROM lilmon_metric_%s
    WHERE
        timestamp >= DATETIME(%d, 'unixepoch')
        AND timestamp <= DATETIME(%d, 'unixepoch')
    ORDER BY timestamp ASC`
	q := fmt.Sprintf(
		template_select_values,
		metric,
		time_start.Unix(),
		time_end.Unix())
	rows, err := db.Query(q)
	if err != nil {
		log.Println("graph_generate: unable to select rows: ", err)
		return nil, err
	}
	defer rows.Close()
	dps := []datapoint{}
	for rows.Next() {
		var ts time.Time
		var value float64

		if err := rows.Scan(&ts, &value); err != nil {
			log.Println("graph_generate: row scan failed: ", err)
			break
		}
		dps = append(dps, datapoint{ts: ts, value: value})
	}
	return dps, nil
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

func db_pruner(ctx context.Context, tasks chan<- db_task, metrics []*metric,
	retention_period, prune_period time.Duration) {

	log.Println("Entering pruning loop with period of ", prune_period)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(prune_period):
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
