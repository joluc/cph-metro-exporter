package collector

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "cph_metro"
)

var (
	// All Copenhagen Metro stations (8603xxx IDs are metro-specific)
	// Queried in batches using Multi Departure Board API for fleet tracking
	allMetroStations = []string{
		// M1/M2 shared trunk
		"8603301", // Vanløse St. (M1, M2)
		"8603302", // Flintholm St. (M1, M2)
		"8603303", // Lindevang St. (M1, M2)
		"8603304", // Fasanvej St. (M1, M2)
		"8603305", // Frederiksberg St. (M1, M2, M3)
		"8603306", // Forum St. (M1, M2)
		"8603307", // Nørreport St. (M1, M2)
		"8603308", // Kongens Nytorv St. (M1, M2, M3, M4)
		"8603309", // Christianshavn St. (M1, M2)

		// M1 branch (to Vestamager)
		"8603310", // Islands Brygge St. (M1)
		"8603311", // DR Byen St. (M1)
		"8603312", // Sundby St. (M1)
		"8603313", // Bella Center St. (M1)
		"8603315", // Ørestad St. (M1)
		"8603317", // Vestamager St. (M1)

		// M2 branch (to Lufthavnen)
		"8603321", // Amagerbro St. (M2)
		"8603322", // Lergravsparken St. (M2)
		"8603323", // Øresund St. (M2)
		"8603324", // Amager Strand St. (M2)
		"8603326", // Femøren St. (M2)
		"8603327", // Kastrup St. (M2)
		"8603328", // Københavns Lufthavn St. (M2)

		// M3 Cityringen & M4 shared
		"8603330", // København H (M3, M4)
		"8603331", // Rådhuspladsen St. (M3, M4)
		"8603332", // Gammel Strand St. (M3, M4)
		"8603333", // Marmorkirken St. (M3, M4)
		"8603334", // Østerport St. (M3, M4)

		// M3 Cityringen only
		"8603335", // Trianglen St. (M3)
		"8603336", // Poul Henningsens Plads St. (M3)
		"8603337", // Vibenshus Runddel St. (M3)
		"8603338", // Skjolds Plads St. (M3)
		"8603339", // Nørrebro St. (M3)
		"8603340", // Nørrebros Runddel St. (M3)
		"8603341", // Nuuks Plads St. (M3)
		"8603342", // Aksel Møllers Have St. (M3)
		"8603343", // Frederiksberg Allé St. (M3)
		"8603344", // Enghave Plads St. (M3)

		// M4 branch (to Orientkaj & København Syd)
		"8603345", // Orientkaj St. (M4)
		"8603346", // Nordhavn St. (M4)
		"8603347", // Havneholmen St. (M4)
		"8603348", // Enghave Brygge St. (M4)
		"8603349", // Sluseholmen St. (M4)
		"8603350", // Mozarts Plads St. (M4)
		"8603351", // København Syd St. (M4)
	}
)

type MetroCollector struct {
	client             StationBoardClient
	activeServices     *prometheus.Desc
	scrapeSuccess      *prometheus.Desc
	scrapeDuration     *prometheus.Desc
	scrapeInterval     time.Duration
	cacheMu            sync.Mutex
	cachedServices     map[string]int
	lastSuccessfulPull time.Time
	logger             *slog.Logger
}

func NewMetroCollector(client StationBoardClient, logger *slog.Logger, scrapeInterval time.Duration) *MetroCollector {
	if scrapeInterval <= 0 {
		scrapeInterval = 30 * time.Minute
	}

	return &MetroCollector{
		client: client,
		activeServices: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "active_services"),
			"Number of metro departures in the next 10 minutes per line",
			[]string{"line"},
			nil,
		),
		scrapeSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_success"),
			"1 if the scrape succeeded, 0 otherwise",
			nil,
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "scrape_duration_seconds"),
			"Duration of the scrape in seconds",
			nil,
			nil,
		),
		scrapeInterval: scrapeInterval,
		logger:         logger,
	}
}

func (c *MetroCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeServices
	ch <- c.scrapeSuccess
	ch <- c.scrapeDuration
}

func (c *MetroCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	start := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, v)
	}))
	defer start.ObserveDuration()

	servicesByLine, success := c.getServiceData(ctx)

	ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, success)

	if servicesByLine != nil {
		for line, count := range servicesByLine {
			ch <- prometheus.MustNewConstMetric(
				c.activeServices,
				prometheus.GaugeValue,
				float64(count),
				line,
			)
		}
	}
}

func (c *MetroCollector) getServiceData(ctx context.Context) (map[string]int, float64) {
	if cached := c.getCachedServiceData(); cached != nil {
		return cached, 1.0
	}

	servicesByLine := c.collectServiceData(ctx)
	if servicesByLine != nil {
		c.cacheMu.Lock()
		c.cachedServices = copyServiceMap(servicesByLine)
		c.lastSuccessfulPull = time.Now()
		c.cacheMu.Unlock()
		return servicesByLine, 1.0
	}

	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	if c.cachedServices != nil {
		// Keep serving stale data but signal refresh failures via scrape_success.
		return copyServiceMap(c.cachedServices), 0.0
	}

	return nil, 0.0
}

func (c *MetroCollector) getCachedServiceData() map[string]int {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if c.cachedServices == nil {
		return nil
	}
	if time.Since(c.lastSuccessfulPull) >= c.scrapeInterval {
		return nil
	}

	return copyServiceMap(c.cachedServices)
}

func copyServiceMap(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (c *MetroCollector) collectServiceData(ctx context.Context) map[string]int {
	servicesByLine := make(map[string]int)
	var mu sync.Mutex
	var wg sync.WaitGroup

	errCount := 0
	totalBatches := 0

	// Query departures in next 10 minutes to get active services
	const durationMinutes = 10

	// Split stations into batches of 10 (API limit for multiDepartureBoard)
	batchSize := 10
	for i := 0; i < len(allMetroStations); i += batchSize {
		end := i + batchSize
		if end > len(allMetroStations) {
			end = len(allMetroStations)
		}
		batch := allMetroStations[i:end]
		totalBatches++

		wg.Add(1)
		go func(stationIDs []string) {
			defer wg.Done()

			board, err := c.client.GetMultiDepartureBoard(ctx, stationIDs, durationMinutes)
			if err != nil {
				c.logger.Debug("Failed to fetch multi departure board", "error", err)
				mu.Lock()
				errCount++
				mu.Unlock()
				return
			}

			// Count metro departures by line
			for _, dep := range board.Departure {
				// Filter for metro only
				if dep.ProductAtStop.CatOut != "MET" {
					continue
				}

				line := strings.ToUpper(strings.TrimSpace(dep.ProductAtStop.Line))

				mu.Lock()
				servicesByLine[line]++
				mu.Unlock()
			}
		}(batch)
	}

	wg.Wait()

	// If all batches failed, return nil to indicate error
	if errCount == totalBatches {
		c.logger.Error("All batch fetches failed")
		return nil
	}

	// Ensure all metro lines are present even if count is 0
	for _, line := range []string{"M1", "M2", "M3", "M4"} {
		if _, exists := servicesByLine[line]; !exists {
			servicesByLine[line] = 0
		}
	}

	c.logger.Info("Collected service data",
		"total_services", sum(servicesByLine),
		"services_by_line", servicesByLine,
		"batches_queried", totalBatches,
		"batches_failed", errCount,
		"time_window_minutes", durationMinutes)

	return servicesByLine
}

func sum(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}