# Copenhagen Metro Exporter

Prometheus exporter for Copenhagen Metro real-time service monitoring.

## Overview

Tracks active metro services across the Copenhagen Metro network by counting scheduled departures in the next 10 minutes per line (M1, M2, M3, M4).

**Why departures, not trains?** The Rejseplanen API 2.0 doesn't provide journey IDs for metro trains, so we count scheduled departures instead. This gives a real-time view of service frequency and capacity.

## Quick Start

### Docker

```bash
# Run with demo data
docker run -p 9100:9100 -e DEMO_MODE=true ghcr.io/joluc/cph-metro-exporter:latest

# Run with your API key
docker run -p 9100:9100 \
  -e REJSEPLANEN_API_KEY=your-api-key \
  ghcr.io/joluc/cph-metro-exporter:latest

# Check metrics
curl http://localhost:9100/metrics
```

### Binary

```bash
# Build
go build -o exporter ./cmd/exporter

# Run with demo mode
./exporter -demo-mode

# Run with API key
REJSEPLANEN_API_KEY=your-key ./exporter
```

## Metrics

```
cph_metro_active_services{line="M1|M2|M3|M4"}  # Departures in next 10 minutes
cph_metro_scrape_success                        # 1 if scrape succeeded, 0 otherwise
cph_metro_scrape_duration_seconds               # Scrape duration
```

## Configuration

Environment variables:
- `REJSEPLANEN_API_KEY` - API key from Rejseplanen Labs
- `PORT` - HTTP server port (default: 9100)
- `SCRAPE_INTERVAL` - Cache refresh interval (default: 30m)
- `LOG_LEVEL` - Log level: debug, info, warn, error (default: info)
- `DEMO_MODE` - Enable demo mode with mock data (default: false)

Command-line flags: `-port`, `-scrape-interval`, `-log-level`, `-demo-mode`, `-api-key`

## API Usage

- **Stations**: 44 metro stations across the network
- **Scrape interval**: 30 minutes (default)
- **API requests**: ~240/day (~7,200/month)
- **Budget**: 50,000 requests/month (14% utilization)

Get your free API key at [Rejseplanen Labs](https://www.rejseplanen.dk/api).

## Kubernetes

```bash
# Install with API key
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set rejseplanen.apiKey=your-api-key

# Install with demo mode
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set config.demoMode=true
```

See [helm/cph-metro-exporter/README.md](helm/cph-metro-exporter/README.md) for detailed configuration options.

## Development

```bash
# Run tests
go test ./...

# Build locally
go build ./cmd/exporter

# Build Docker image
docker build -t cph-metro-exporter .
```

## License

[Your license here]

## Fleet Reference

Copenhagen Metro operates 81 trainsets:
- M1/M2 lines: 42 trains (shared fleet)
- M3/M4 lines: 39 trains (shared fleet)
