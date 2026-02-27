package collector

import (
	"context"

	"github.com/joluc/cph-metro-exporter/internal/rejseplanen"
)

type StationBoardClient interface {
	GetStationBoard(ctx context.Context, stationID string) (*rejseplanen.StationBoardResponse, error)
	GetMultiDepartureBoard(ctx context.Context, stationIDs []string, durationMinutes int) (*rejseplanen.MultiDepartureBoardResponse, error)
}
