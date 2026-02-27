package rejseplanen

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const (
	BaseURL = "https://www.rejseplanen.dk/api"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	logger     *slog.Logger
	baseURL    string
}

func NewClient(apiKey string, timeout time.Duration, logger *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		apiKey:  apiKey,
		logger:  logger,
		baseURL: BaseURL,
	}
}

type StationBoardResponse struct {
	Departure []Departure `json:"Departure"`
}

// MultiDepartureBoardResponse returns departures from multiple stations
// The API returns a flat list with stopExtId indicating which station each departure is from
type MultiDepartureBoardResponse struct {
	Departure []Departure `json:"Departure"`
}

type Departure struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	Stop          string        `json:"stop"`
	StopExtId     string        `json:"stopExtId"`
	Time          string        `json:"time"`
	Date          string        `json:"date"`
	Direction     string        `json:"direction"`
	Track         string        `json:"track"`
	JourneyID     JourneyRef    `json:"JourneyDetailRef"`
	ProductAtStop ProductAtStop `json:"ProductAtStop"`
}

type JourneyRef struct {
	Ref string `json:"ref"`
}

type ProductAtStop struct {
	Line   string `json:"line"`
	CatOut string `json:"catOut"`
}

func (c *Client) GetStationBoard(ctx context.Context, stationID string) (*StationBoardResponse, error) {
	url := fmt.Sprintf("%s/departureBoard?id=%s&format=json&accessId=%s", c.baseURL, stationID, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.logger.Debug("Fetching station board", "station_id", stationID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch station board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var board StationBoardResponse

	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &board, nil
}

// GetMultiDepartureBoard fetches departure boards for multiple stations in a single API call
// stationIDs can contain up to 10 station IDs
// duration specifies the time window in minutes (default 60 if not specified)
func (c *Client) GetMultiDepartureBoard(ctx context.Context, stationIDs []string, durationMinutes int) (*MultiDepartureBoardResponse, error) {
	if len(stationIDs) == 0 {
		return nil, fmt.Errorf("no station IDs provided")
	}
	if len(stationIDs) > 10 {
		return nil, fmt.Errorf("too many station IDs: %d (max 10)", len(stationIDs))
	}

	// Build idList parameter with pipe-separated station IDs
	idList := ""
	for i, id := range stationIDs {
		if i > 0 {
			idList += "|"
		}
		idList += id
	}

	url := fmt.Sprintf("%s/multiDepartureBoard?idList=%s&duration=%d&format=json&accessId=%s",
		c.baseURL, idList, durationMinutes, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.logger.Debug("Fetching multi departure board", "station_count", len(stationIDs), "duration_minutes", durationMinutes)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch multi departure board: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var board MultiDepartureBoardResponse

	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &board, nil
}

func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}
