package rejseplanen

import (
	"context"
	"log/slog"
	"math/rand"
	"time"
)

type MockClient struct {
	logger *slog.Logger
}

func NewMockClient(logger *slog.Logger) *MockClient {
	return &MockClient{
		logger: logger,
	}
}

func (m *MockClient) GetStationBoard(ctx context.Context, stationID string) (*StationBoardResponse, error) {
	m.logger.Debug("Mock: Fetching station board", "station_id", stationID)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	numDepartures := r.Intn(10) + 5
	departures := make([]Departure, numDepartures)

	lines := []string{"M1", "M2", "M3", "M4"}

	for i := 0; i < numDepartures; i++ {
		line := lines[r.Intn(len(lines))]
		departures[i] = Departure{
			Name: "Metro " + line,
			Type: line,
			ProductAtStop: ProductAtStop{
				Line:   line,
				CatOut: "MET",
			},
			JourneyID: JourneyRef{
				Ref: generateJourneyID(stationID, i, r),
			},
			Stop:      stationID,
			Time:      time.Now().Add(time.Duration(r.Intn(30)) * time.Minute).Format("15:04"),
			Date:      time.Now().Format("02.01.06"),
			Direction: "Test Direction",
		}
	}

	return &StationBoardResponse{
		Departure: departures,
	}, nil
}

func (m *MockClient) GetMultiDepartureBoard(ctx context.Context, stationIDs []string, durationMinutes int) (*MultiDepartureBoardResponse, error) {
	m.logger.Debug("Mock: Fetching multi departure board", "station_count", len(stationIDs), "duration_minutes", durationMinutes)

	allDepartures := []Departure{}

	for _, stationID := range stationIDs {
		board, err := m.GetStationBoard(ctx, stationID)
		if err != nil {
			return nil, err
		}

		// Add station ID to each departure
		for _, dep := range board.Departure {
			dep.StopExtId = stationID
			allDepartures = append(allDepartures, dep)
		}
	}

	return &MultiDepartureBoardResponse{
		Departure: allDepartures,
	}, nil
}

func generateJourneyID(stationID string, index int, r *rand.Rand) string {
	sharedJourneyChance := 0.3
	if r.Float64() < sharedJourneyChance {
		return "shared_journey_" + stationID[:4]
	}
	return stationID + "_journey_" + time.Now().Format("20060102150405") + "_" + string(rune(index))
}
