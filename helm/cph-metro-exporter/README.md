# Copenhagen Metro Exporter Helm Chart

Helm chart for deploying the Copenhagen Metro Exporter to Kubernetes.

## Installation

### With API Key

```bash
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set rejseplanen.apiKey=your-api-key-here
```

### With Existing Secret

```bash
# Create secret first
kubectl create secret generic rejseplanen-api \
  --from-literal=api-key=your-api-key-here

# Install with existing secret
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set rejseplanen.existingSecret=rejseplanen-api
```

### Demo Mode (No API Key Required)

```bash
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set config.demoMode=true
```

## Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `ghcr.io/joluc/cph-metro-exporter` |
| `image.tag` | Image tag | `1.0.0` |
| `config.scrapeInterval` | Scrape interval | `30m` |
| `config.logLevel` | Log level | `info` |
| `config.demoMode` | Enable demo mode | `false` |
| `rejseplanen.apiKey` | API key (creates Secret) | `""` |
| `rejseplanen.existingSecret` | Existing Secret name | `""` |
| `serviceMonitor.enabled` | Enable Prometheus ServiceMonitor | `true` |
| `resources.requests.cpu` | CPU request | `50m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `128Mi` |

### Full Example

```bash
helm install cph-metro-exporter ./helm/cph-metro-exporter \
  --set rejseplanen.apiKey=your-api-key \
  --set config.scrapeInterval=15m \
  --set config.logLevel=debug \
  --set resources.requests.cpu=100m \
  --set resources.requests.memory=128Mi
```

### values.yaml

For more complex configurations, create a `values.yaml` file:

```yaml
image:
  tag: "1.0.0"

config:
  scrapeInterval: "30m"
  logLevel: "info"

rejseplanen:
  apiKey: "your-api-key"

resources:
  requests:
    cpu: 50m
    memory: 64Mi
  limits:
    cpu: 100m
    memory: 128Mi

serviceMonitor:
  enabled: true
  interval: 30s
```

Then install:

```bash
helm install cph-metro-exporter ./helm/cph-metro-exporter -f values.yaml
```

## Upgrading

```bash
helm upgrade cph-metro-exporter ./helm/cph-metro-exporter \
  --set rejseplanen.apiKey=your-api-key
```

## Uninstallation

```bash
helm uninstall cph-metro-exporter
```

## ServiceMonitor

If you have Prometheus Operator installed, the chart will create a ServiceMonitor to automatically scrape metrics:

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
```

## Security

The chart follows security best practices:
- Runs as non-root user (UID 65534)
- Read-only root filesystem
- Drops all capabilities
- No privilege escalation

## Requirements

- Kubernetes 1.19+
- Helm 3.0+
- Prometheus Operator (optional, for ServiceMonitor)
