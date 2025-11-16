# Monitoring and Observability

This directory contains monitoring and observability configurations for the Red Team C2 Framework.

## Overview

The framework includes comprehensive monitoring capabilities:

- **Metrics**: Prometheus for time-series metrics
- **Dashboards**: Grafana for visualization
- **Logging**: Centralized log aggregation
- **Alerting**: Alert rules for critical events

## Components

### Prometheus

**Configuration**: `prometheus.yml`

Metrics collected:
- Beacon count (active, inactive, total)
- Command queue depth
- API request latency
- Database query performance
- System resource usage

**Access**: http://localhost:9090

### Grafana

**Configuration**: `grafana/`

Pre-built dashboards:
- C2 Operations Overview
- Beacon Activity
- System Performance
- Security Alerts

**Access**: http://localhost:3000
**Default credentials**: admin / admin

### Log Aggregation

**Configuration**: `logging/`

Centralized logging for:
- Teamserver logs
- Database logs
- Beacon activity logs
- Audit trail

### Alerts

**Configuration**: `alerts/`

Alert rules for:
- Beacon disconnections
- Failed authentication attempts
- High resource usage
- Database connection failures

## Quick Start

### Deploy with Monitoring

```bash
# Start monitoring stack
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# Access Grafana
open http://localhost:3000

# Access Prometheus
open http://localhost:9090
```

### View Metrics

```bash
# List all metrics
curl http://localhost:9090/api/v1/label/__name__/values

# Query beacon count
curl 'http://localhost:9090/api/v1/query?query=c2_beacons_active'

# Query command queue
curl 'http://localhost:9090/api/v1/query?query=c2_commands_queued'
```

### Configure Alerts

Edit `alerts/rules.yml`:

```yaml
groups:
  - name: c2_alerts
    rules:
      - alert: HighBeaconDropRate
        expr: rate(c2_beacons_disconnected[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High beacon drop rate detected"
```

## Metrics Reference

### Beacon Metrics

- `c2_beacons_total` - Total beacons registered
- `c2_beacons_active` - Currently active beacons
- `c2_beacons_inactive` - Inactive beacons
- `c2_beacons_connected` - Connection rate
- `c2_beacons_disconnected` - Disconnection rate

### Command Metrics

- `c2_commands_total` - Total commands issued
- `c2_commands_queued` - Commands in queue
- `c2_commands_executed` - Commands executed
- `c2_commands_failed` - Failed commands
- `c2_command_duration_seconds` - Command execution time

### API Metrics

- `c2_api_requests_total` - Total API requests
- `c2_api_request_duration_seconds` - Request latency
- `c2_api_errors_total` - API errors

### Database Metrics

- `c2_db_connections_active` - Active connections
- `c2_db_query_duration_seconds` - Query latency
- `c2_db_size_bytes` - Database size

## Dashboards

### C2 Operations Overview

**File**: `grafana/dashboards/c2_overview.json`

Panels:
- Active beacons (last 5 minutes)
- Command execution rate
- API request rate
- Top 10 active beacons
- Recent beacon activity timeline

### Beacon Activity

**File**: `grafana/dashboards/beacon_activity.json`

Panels:
- Beacon check-in frequency
- Command queue per beacon
- Output collection rate
- Beacon geographic distribution
- Connection duration histogram

### System Performance

**File**: `grafana/dashboards/system_performance.json`

Panels:
- CPU usage
- Memory usage
- Disk I/O
- Network traffic
- Database performance

## Logging

### Log Levels

- `ERROR` - Critical errors requiring immediate attention
- `WARN` - Warning conditions
- `INFO` - Informational messages
- `DEBUG` - Detailed debugging information

### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "component": "teamserver",
  "message": "Beacon checked in",
  "beacon_id": "beacon-123",
  "ip": "192.168.1.100"
}
```

### Viewing Logs

```bash
# All logs
docker-compose logs -f

# Specific service
docker-compose logs -f teamserver

# Filter by level
docker-compose logs teamserver | grep ERROR

# Export logs
docker-compose logs --no-color > logs_export.txt
```

## Alerting

### Slack Integration

```yaml
# alerts/alertmanager.yml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'YOUR_WEBHOOK_URL'
        channel: '#c2-alerts'
        title: 'C2 Alert: {{ .GroupLabels.alertname }}'
        text: '{{ .Annotations.summary }}'
```

### Email Alerts

```yaml
receivers:
  - name: 'email'
    email_configs:
      - to: 'operator@example.com'
        from: 'c2-alerts@example.com'
        smarthost: 'smtp.example.com:587'
```

### PagerDuty Integration

```yaml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_SERVICE_KEY'
```

## Best Practices

1. **Regular Review**: Check dashboards daily
2. **Alert Tuning**: Adjust thresholds to reduce false positives
3. **Log Retention**: Rotate logs to manage disk space
4. **Backup Metrics**: Export important metrics regularly
5. **Security**: Restrict access to monitoring interfaces

## Troubleshooting

### Prometheus Not Scraping

```bash
# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Verify network connectivity
docker-compose exec prometheus wget -O- http://teamserver:8443/metrics
```

### Grafana Can't Connect to Prometheus

```bash
# Check Grafana data sources
curl -u admin:admin http://localhost:3000/api/datasources

# Test connection from Grafana container
docker-compose exec grafana wget -O- http://prometheus:9090/api/v1/query?query=up
```

### Missing Metrics

```bash
# Enable metrics endpoint in teamserver
# Add to teamserver configuration:
export ENABLE_METRICS=true
export METRICS_PORT=9090
```

## Security Considerations

- **Authentication**: Secure Grafana with strong passwords
- **Network Isolation**: Use internal Docker network
- **TLS**: Enable HTTPS for Grafana and Prometheus
- **Access Control**: Limit who can view metrics
- **Data Retention**: Configure appropriate retention policies

## Support

- **Prometheus Docs**: https://prometheus.io/docs/
- **Grafana Docs**: https://grafana.com/docs/
- **Issues**: https://github.com/Raoof128/CRTIAE/issues
