package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
)

type Config struct {
	host                      string
	port                      int
	user                      string
	password                  string
	database                  string
	schema                    string
	sslMode                   string
	table                     string
	copyTable                 string
	maxOpenConns              int
	maxIdleConns              int
	pgPrometheusNormalize     bool
	pgPrometheusLogSamples    bool
	pgPrometheusChunkInterval time.Duration
	useTimescaleDb            bool
	dbConnectRetries          int
}

type Client struct {
	db  *sql.DB
	cfg *Config
}

func metricString(m model.Metric) string {
	metricName, hasName := m[model.MetricNameLabel]
	numLabels := len(m) - 1
	if !hasName {
		numLabels = len(m)
	}
	labelStrings := make([]string, 0, numLabels)
	for label, value := range m {
		if label != model.MetricNameLabel {
			labelStrings = append(labelStrings, fmt.Sprintf("%s=%q", label, value))
		}
	}

	switch numLabels {
	case 0:
		if hasName {
			return string(metricName)
		}
		return "{}"
	default:
		sort.Strings(labelStrings)
		return fmt.Sprintf("%s{%s}", metricName, strings.Join(labelStrings, ","))
	}
}

func (c *Client) Write(samples model.Samples) error {
	begin := time.Now()
	tx, err := c.db.Begin()

	if err != nil {
		log.Error("msg", "Error on Begin when writing samples", "err", err)
		return err
	}

	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("COPY \"%s\" FROM STDIN", c.cfg.copyTable))

	if err != nil {
		log.Error("msg", "Error on Prepare when writing samples", "err", err)
		return err
	}

	for _, sample := range samples {
		milliseconds := sample.Timestamp.UnixNano() / 1000000
		line := fmt.Sprintf("%v %v %v", metricString(sample.Metric), sample.Value, milliseconds)

		if c.cfg.pgPrometheusLogSamples {
			fmt.Println(line)
		}

		_, err = stmt.Exec(line)
		if err != nil {
			log.Error("msg", "Error executing statement", "stmt", line, "err", err)
			return err
		}

	}

	_, err = stmt.Exec()
	if err != nil {
		log.Error("msg", "Error executing close of copy", "err", err)
		return err
	}

	err = stmt.Close()

	if err != nil {
		log.Error("msg", "Error on Close when writing samples", "err", err)
		return err
	}

	err = tx.Commit()

	if err != nil {
		log.Error("msg", "Error on Commit when writing samples", "err", err)
		return err
	}

	duration := time.Since(begin).Seconds()

	log.Debug("msg", "Wrote samples", "count", len(samples), "duration", duration)

	return nil
}
