package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joluc/cph-metro-exporter/internal/collector"
	"github.com/joluc/cph-metro-exporter/internal/rejseplanen"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if string(body) != "OK" {
		t.Errorf("expected body 'OK', got %s", string(body))
	}
}

func TestMetricsEndpoint(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a mock server for Rejseplanen API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"DepartureBoard": {
				"Departure": [
					{
						"name": "Metro M1",
						"type": "M1",
						"line": "M1",
						"JourneyDetailRef": "journey1"
					}
				]
			}
		}`))
	}))
	defer mockServer.Close()

	client := rejseplanen.NewClient("", 5*time.Second, logger)
	client.SetBaseURL(mockServer.URL)

	metroCollector := collector.NewMetroCollector(client, logger, 30*time.Second)

	reg := prometheus.NewRegistry()
	reg.MustRegister(metroCollector)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if !strings.Contains(bodyStr, "cph_metro_active_services") {
		t.Error("expected metrics to contain cph_metro_active_services")
	}

	if !strings.Contains(bodyStr, "cph_metro_scrape_success") {
		t.Error("expected metrics to contain cph_metro_scrape_success")
	}
}
