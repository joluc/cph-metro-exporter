package rejseplanen

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestGetStationBoard(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful response",
			mockResponse: map[string]interface{}{
				"Departure": []map[string]interface{}{
					{
						"name":      "Metro M1",
						"type":      "M1",
						"stop":      "Nørreport St.",
						"time":      "14:30",
						"date":      "23.02.26",
						"direction": "Vestamager",
						"JourneyDetailRef": map[string]interface{}{
							"ref": "journey123",
						},
						"ProductAtStop": map[string]interface{}{
							"line":   "M1",
							"catOut": "MET",
						},
					},
					{
						"name":      "Metro M2",
						"type":      "M2",
						"stop":      "Nørreport St.",
						"time":      "14:32",
						"date":      "23.02.26",
						"direction": "Lufthavnen",
						"JourneyDetailRef": map[string]interface{}{
							"ref": "journey456",
						},
						"ProductAtStop": map[string]interface{}{
							"line":   "M2",
							"catOut": "MET",
						},
					},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			mockResponse:   nil,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "invalid json",
			mockResponse:   "invalid json",
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					if str, ok := tt.mockResponse.(string); ok {
						w.Write([]byte(str))
					} else {
						json.NewEncoder(w).Encode(tt.mockResponse)
					}
				}
			}))
			defer server.Close()

			client := NewClient("test-key", 5*time.Second, logger)
			client.baseURL = server.URL

			ctx := context.Background()
			board, err := client.GetStationBoard(ctx, "8600626")

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if board == nil {
					t.Error("expected board but got nil")
				} else if len(board.Departure) != 2 {
					t.Errorf("expected 2 departures, got %d", len(board.Departure))
				}
			}
		})
	}
}

func TestGetStationBoardContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"Departure": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", 5*time.Second, logger)
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := client.GetStationBoard(ctx, "8600626")
	if err == nil {
		t.Error("expected context timeout error but got none")
	}
}
