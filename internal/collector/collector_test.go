package collector

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/joluc/cph-metro-exporter/internal/rejseplanen"
	"github.com/prometheus/client_golang/prometheus"
)

type mockClient struct {
	boards map[string]*rejseplanen.StationBoardResponse
	err    error
	mu     sync.Mutex
	calls  int
}

func (m *mockClient) GetStationBoard(ctx context.Context, stationID string) (*rejseplanen.StationBoardResponse, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}
	if board, ok := m.boards[stationID]; ok {
		return board, nil
	}
	return &rejseplanen.StationBoardResponse{Departure: []rejseplanen.Departure{}}, nil
}

func (m *mockClient) GetMultiDepartureBoard(ctx context.Context, stationIDs []string, durationMinutes int) (*rejseplanen.MultiDepartureBoardResponse, error) {
	m.mu.Lock()
	m.calls++
	m.mu.Unlock()

	if m.err != nil {
		return nil, m.err
	}

	// Build flat response from individual station boards
	allDepartures := []rejseplanen.Departure{}

	for _, stationID := range stationIDs {
		if board, ok := m.boards[stationID]; ok {
			// Add station ID to each departure
			for _, dep := range board.Departure {
				dep.StopExtId = stationID
				allDepartures = append(allDepartures, dep)
			}
		}
	}

	return &rejseplanen.MultiDepartureBoardResponse{
		Departure: allDepartures,
	}, nil
}

func newMockClient(boards map[string]*rejseplanen.StationBoardResponse, err error) *mockClient {
	return &mockClient{
		boards: boards,
		err:    err,
	}
}

func (m *mockClient) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.calls
}

func TestMetroCollector(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name          string
		mockBoards    map[string]*rejseplanen.StationBoardResponse
		mockErr       error
		expectMetrics map[string]float64
		expectSuccess float64
	}{
		{
			name: "unique trains with synthetic IDs across stations",
			mockBoards: map[string]*rejseplanen.StationBoardResponse{
				"8600626": { // Nørreport
					Departure: []rejseplanen.Departure{
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
							Direction:     "Vestamager",
							Time:          "12:30:00",
							Track:         "1",
						},
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M2", CatOut: "MET"},
							Direction:     "Lufthavnen",
							Time:          "12:31:00",
							Track:         "2",
						},
					},
				},
				"8600650": { // Kongens Nytorv
					Departure: []rejseplanen.Departure{
						// Same M1 train seen at next station (should deduplicate)
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
							Direction:     "Vestamager",
							Time:          "12:30:00",
							Track:         "1",
						},
						// Different M3 train
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M3", CatOut: "MET"},
							Direction:     "Cityringen",
							Time:          "12:32:00",
							Track:         "3",
						},
					},
				},
			},
			expectMetrics: map[string]float64{
				"M1": 1, // Deduplicated across stations
				"M2": 1,
				"M3": 1,
				"M4": 0,
			},
			expectSuccess: 1.0,
		},
		{
			name: "different trains same line distinguished by time/track",
			mockBoards: map[string]*rejseplanen.StationBoardResponse{
				"8600626": {
					Departure: []rejseplanen.Departure{
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
							Direction:     "Vestamager",
							Time:          "12:30:00",
							Track:         "1",
						},
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
							Direction:     "Vestamager",
							Time:          "12:35:00", // Different time = different train
							Track:         "1",
						},
					},
				},
			},
			expectMetrics: map[string]float64{
				"M1": 2, // Two different M1 trains
				"M2": 0,
				"M3": 0,
				"M4": 0,
			},
			expectSuccess: 1.0,
		},
		{
			name: "no trains",
			mockBoards: map[string]*rejseplanen.StationBoardResponse{
				"8600626": {Departure: []rejseplanen.Departure{}},
			},
			expectMetrics: map[string]float64{
				"M1": 0,
				"M2": 0,
				"M3": 0,
				"M4": 0,
			},
			expectSuccess: 1.0,
		},
		{
			name: "filters non-metro departures",
			mockBoards: map[string]*rejseplanen.StationBoardResponse{
				"8600626": {
					Departure: []rejseplanen.Departure{
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
							Direction:     "Vestamager",
							Time:          "12:30:00",
							Track:         "1",
						},
						{
							ProductAtStop: rejseplanen.ProductAtStop{Line: "1A", CatOut: "BUS"},
							Direction:     "Somewhere",
							Time:          "12:31:00",
							Track:         "2",
						},
					},
				},
			},
			expectMetrics: map[string]float64{
				"M1": 1, // Only metro counted
				"M2": 0,
				"M3": 0,
				"M4": 0,
			},
			expectSuccess: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockClient(tt.mockBoards, tt.mockErr)
			collector := NewMetroCollector(mock, logger, 0)

			// Create a custom registry for testing
			reg := prometheus.NewRegistry()
			reg.MustRegister(collector)

			// Collect metrics
			metricFamilies, err := reg.Gather()
			if err != nil {
				t.Fatalf("failed to gather metrics: %v", err)
			}

			// Verify metrics
			for _, mf := range metricFamilies {
				switch mf.GetName() {
				case "cph_metro_running_trains":
					for _, m := range mf.GetMetric() {
						line := ""
						for _, l := range m.GetLabel() {
							if l.GetName() == "line" {
								line = l.GetValue()
								break
							}
						}
						if expected, ok := tt.expectMetrics[line]; ok {
							if m.GetGauge().GetValue() != expected {
								t.Errorf("line %s: expected %v trains, got %v", line, expected, m.GetGauge().GetValue())
							}
						}
					}
				case "cph_metro_scrape_success":
					for _, m := range mf.GetMetric() {
						if m.GetGauge().GetValue() != tt.expectSuccess {
							t.Errorf("expected scrape_success %v, got %v", tt.expectSuccess, m.GetGauge().GetValue())
						}
					}
				}
			}
		})
	}
}

func TestMetroCollectorUsesScrapeIntervalCache(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mock := newMockClient(map[string]*rejseplanen.StationBoardResponse{
		"8600626": {
			Departure: []rejseplanen.Departure{
				{
					ProductAtStop: rejseplanen.ProductAtStop{Line: "M1", CatOut: "MET"},
					Direction:     "Vestamager",
					Time:          "12:30:00",
					Track:         "1",
				},
			},
		},
	}, nil)
	collector := NewMetroCollector(mock, logger, time.Hour)

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	// First gather - should query all stations
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("first gather failed: %v", err)
	}
	firstCallCount := mock.CallCount()

	// Second gather within cache interval - should use cache, no new calls
	if _, err := reg.Gather(); err != nil {
		t.Fatalf("second gather failed: %v", err)
	}
	secondCallCount := mock.CallCount()

	// Call count should be the same (no additional API calls due to caching)
	if secondCallCount != firstCallCount {
		t.Fatalf("expected cache to prevent additional calls: first=%d, second=%d", firstCallCount, secondCallCount)
	}
}
