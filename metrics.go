package pq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	methodExec     = "exec"
	methodQuery    = "query"
	methodQueryRow = "query_row"
)

var clientDurationSummary *prometheus.SummaryVec

type metricsAdapter struct {
	Client
	name string
}

func init() {
	clientDurationSummary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "pq_operations_durations_seconds",
			Help: "pq_operations_durations_seconds",
		},
		[]string{"client_name", "method"},
	)

	prometheus.MustRegister(clientDurationSummary)
}

func (ta *metricsAdapter) Exec(ctx context.Context, sql string, args ...interface{}) (result RowsAffected, err error) {
	start := time.Now()

	rowsAffected, err := ta.Client.Exec(ctx, sql, args...)

	ta.observe(methodExec, start)
	return rowsAffected, err
}

func (ta *metricsAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	start := time.Now()

	rows, err := ta.Client.Query(ctx, sql, args...)

	ta.observe(methodQuery, start)
	return rows, err
}

func (ta *metricsAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	start := time.Now()

	row := ta.Client.QueryRow(ctx, sql, args...)

	ta.observe(methodQueryRow, start)
	return row
}

func (ta *metricsAdapter) observe(method string, startedAt time.Time) {
	duration := time.Since(startedAt)
	clientDurationSummary.WithLabelValues(ta.name, method).Observe(duration.Seconds())
}

// prometheusCollector exports metrics from db.DBStats as prometheus` gauges.
type prometheusCollector struct {
	mu                sync.RWMutex
	dbs               map[string]*pgxpool.Pool
	openedConnections *prometheus.Desc
	maxConnections    *prometheus.Desc
	unusedConnections *prometheus.Desc
}

var errAlreadyRegistered = errors.New("already registered")

// register adds connection to pool. Returns an error on duplicate pool name.
func (pc *prometheusCollector) register(name string, conn *pgxpool.Pool) error {
	if name == "" {
		name = "default"
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if _, exists := pc.dbs[name]; exists {
		return errAlreadyRegistered
	}

	pc.dbs[name] = conn

	return nil
}

var collector *prometheusCollector

func init() {
	collector = &prometheusCollector{
		dbs:               make(map[string]*pgxpool.Pool),
		openedConnections: prometheus.NewDesc("db_open_connections", "db open connections", []string{"name"}, nil),
		maxConnections:    prometheus.NewDesc("db_max_connections", "db max connections", []string{"name"}, nil),
		unusedConnections: prometheus.NewDesc("db_unused_connections", "db unused connections", []string{"name"}, nil),
	}

	prometheus.MustRegister(collector)
}

// Describe prometheus.Collector interface implementation
func (pc *prometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- pc.openedConnections
}

// Collect prometheus.Collector interface implementation
func (pc *prometheusCollector) Collect(ch chan<- prometheus.Metric) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	for poolName, pool := range pc.dbs {
		stats := pool.Stat()

		ch <- prometheus.MustNewConstMetric(pc.openedConnections, prometheus.GaugeValue, float64(stats.AcquiredConns()), poolName)
		ch <- prometheus.MustNewConstMetric(pc.unusedConnections, prometheus.GaugeValue, float64(stats.IdleConns()), poolName)
		ch <- prometheus.MustNewConstMetric(pc.maxConnections, prometheus.GaugeValue, float64(stats.MaxConns()), poolName)
	}
}
