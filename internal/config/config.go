package config

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Port            string
	APIKey          string
	LogLevel        string
	ScrapeInterval  time.Duration
	RequestTimeout  time.Duration
	DemoMode        bool
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.Port, "port", getEnv("PORT", "9100"), "HTTP server port")
	flag.StringVar(&cfg.APIKey, "api-key", getEnv("REJSEPLANEN_API_KEY", ""), "Rejseplanen API key")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.DurationVar(&cfg.ScrapeInterval, "scrape-interval", getDurationEnv("SCRAPE_INTERVAL", 30*time.Minute), "Scrape interval")
	flag.DurationVar(&cfg.RequestTimeout, "request-timeout", getDurationEnv("REQUEST_TIMEOUT", 10*time.Second), "HTTP request timeout")
	flag.BoolVar(&cfg.DemoMode, "demo-mode", os.Getenv("DEMO_MODE") == "true", "Enable demo mode with mock data")

	flag.Parse()

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
