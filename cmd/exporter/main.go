package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joluc/cph-metro-exporter/internal/collector"
	"github.com/joluc/cph-metro-exporter/internal/config"
	"github.com/joluc/cph-metro-exporter/internal/rejseplanen"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.Load()

	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	// Create Rejseplanen client
	var client collector.StationBoardClient
	if cfg.DemoMode {
		logger.Info("Running in demo mode with mock data")
		client = rejseplanen.NewMockClient(logger)
	} else {
		client = rejseplanen.NewClient(cfg.APIKey, cfg.RequestTimeout, logger)
	}

	// Create and register collector
	metroCollector := collector.NewMetroCollector(client, logger, cfg.ScrapeInterval)
	prometheus.MustRegister(metroCollector)

	// Set up HTTP handlers
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html>
<head><title>Copenhagen Metro Exporter</title></head>
<body>
<h1>Copenhagen Metro Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/health">Health</a></p>
</body>
</html>`))
	})

	logger.Info("Starting Copenhagen Metro Exporter", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		logger.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
