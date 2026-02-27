# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - Initial Release

### Features
- Real-time Copenhagen Metro service monitoring
- Tracks active services (departures in next 10 minutes) per line (M1, M2, M3, M4)
- Comprehensive coverage of all 44 metro stations across the network
- Multi Departure Board API integration with batching (10 stations per request)
- Prometheus metrics exposition on port 9100
- Docker support with multi-architecture images (amd64, arm64)
- Helm chart for Kubernetes deployments
- Demo mode for testing without API key

### Technical Details
- Queries 5 batches of stations every 30 minutes
- API budget: ~7,200 requests/month (14% of quota)
- Time window: 10-minute lookahead for service counting
- Built with Go 1.23, using Rejseplanen API 2.0

### Metrics
- `cph_metro_active_services{line}` - Number of departures in next 10 minutes per line
- `cph_metro_scrape_success` - Scrape health indicator
- `cph_metro_scrape_duration_seconds` - Scrape performance metric

### Note on Approach
This exporter tracks **active services** (scheduled departures) rather than unique physical trains, since the Rejseplanen API 2.0 doesn't provide journey IDs for metro trains. This provides a real-time view of service frequency and capacity, which is more useful for monitoring purposes than attempting to count physical trains.

**Copenhagen Metro Fleet Reference:**
- Total: 81 trainsets
- M1/M2: 42 trains (shared fleet)
- M3/M4: 39 trains (shared fleet)

