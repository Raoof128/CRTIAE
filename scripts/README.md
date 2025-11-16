# Deployment and Operations Scripts

This directory contains automation scripts for deploying, managing, and monitoring the Red Team C2 Framework.

## Scripts Overview

### Setup & Installation

**`setup.sh`** - One-command environment setup
- Checks prerequisites (Docker, Go, Rust, Python)
- Installs missing dependencies
- Configures environment
- Initializes database

Usage:
```bash
./scripts/setup.sh
```

### Deployment

**`deploy.sh`** - Production deployment
- Builds all components
- Configures TLS certificates
- Starts services with production settings
- Runs health checks

Usage:
```bash
./scripts/deploy.sh [development|production]
```

**`deploy_docker.sh`** - Docker-only deployment
- Quick Docker Compose deployment
- Development mode by default
- Minimal configuration

Usage:
```bash
./scripts/deploy_docker.sh
```

### Health & Monitoring

**`healthcheck.sh`** - System health verification
- Checks all services
- Verifies database connectivity
- Tests API endpoints
- Validates beacon connectivity

Usage:
```bash
./scripts/healthcheck.sh
```

**`monitor.sh`** - Continuous monitoring
- Real-time beacon count
- Service status
- Resource usage
- Alerts on failures

Usage:
```bash
./scripts/monitor.sh [interval_seconds]
```

### Backup & Recovery

**`backup.sh`** - Backup database and configurations
- PostgreSQL database backup
- Configuration files backup
- Timestamped archives
- Optional encryption

Usage:
```bash
./scripts/backup.sh [output_directory]
```

**`restore.sh`** - Restore from backup
- Restores database
- Restores configurations
- Validates integrity

Usage:
```bash
./scripts/restore.sh [backup_file]
```

### Maintenance

**`cleanup.sh`** - Clean up old data
- Removes inactive beacons
- Cleans old command outputs
- Purges logs
- Vacuums database

Usage:
```bash
./scripts/cleanup.sh [days_to_keep]
```

**`update.sh`** - Update framework
- Pulls latest code
- Rebuilds components
- Runs migrations
- Restarts services

Usage:
```bash
./scripts/update.sh
```

## Quick Reference

### Initial Setup

```bash
# 1. Setup environment
./scripts/setup.sh

# 2. Deploy services
./scripts/deploy.sh development

# 3. Verify health
./scripts/healthcheck.sh
```

### Daily Operations

```bash
# Check system health
./scripts/healthcheck.sh

# Monitor beacons
./scripts/monitor.sh 60

# Backup data
./scripts/backup.sh /backups
```

### Maintenance

```bash
# Clean old data (30 days)
./scripts/cleanup.sh 30

# Update framework
./scripts/update.sh

# Restore from backup
./scripts/restore.sh /backups/backup_20240115.tar.gz
```

## Environment Variables

Scripts respect these environment variables:

- `C2_ENV` - Environment (development/production)
- `DATABASE_URL` - PostgreSQL connection string
- `SERVER_ADDR` - Teamserver listen address
- `BACKUP_DIR` - Backup directory
- `LOG_LEVEL` - Logging level (debug/info/warn/error)

Example:
```bash
export C2_ENV=production
export SERVER_ADDR=0.0.0.0:8443
./scripts/deploy.sh
```

## Automation

### Cron Jobs

**Daily backup**:
```cron
0 2 * * * /path/to/CRTIAE/scripts/backup.sh /backups
```

**Hourly health check**:
```cron
0 * * * * /path/to/CRTIAE/scripts/healthcheck.sh || /path/to/CRTIAE/scripts/deploy.sh
```

**Weekly cleanup**:
```cron
0 3 * * 0 /path/to/CRTIAE/scripts/cleanup.sh 30
```

### Systemd Service

Create `/etc/systemd/system/c2-framework.service`:

```ini
[Unit]
Description=Red Team C2 Framework
After=network.target

[Service]
Type=forking
WorkingDirectory=/path/to/CRTIAE
ExecStart=/path/to/CRTIAE/scripts/deploy.sh production
ExecStop=/usr/bin/docker-compose down
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable:
```bash
sudo systemctl enable c2-framework
sudo systemctl start c2-framework
```

## Best Practices

1. **Always backup before updates**: Run `backup.sh` before `update.sh`
2. **Test in development**: Use `development` mode before `production`
3. **Monitor regularly**: Set up cron jobs for `healthcheck.sh`
4. **Clean old data**: Run `cleanup.sh` weekly to manage database size
5. **Review logs**: Check logs after deployment with `docker-compose logs`

## Troubleshooting

### Script won't execute

```bash
# Make scripts executable
chmod +x scripts/*.sh
```

### Permission denied

```bash
# Run with sudo for Docker commands
sudo ./scripts/deploy.sh
```

### Database connection fails

```bash
# Check DATABASE_URL
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1;"
```

## Security Considerations

- **Backup encryption**: Use `backup.sh --encrypt` for sensitive data
- **Secure credentials**: Never commit database passwords
- **TLS certificates**: Use valid certificates in production
- **Access control**: Restrict script execution to authorized users

## Support

- **Documentation**: See main `README.md`
- **Issues**: https://github.com/Raoof128/CRTIAE/issues
- **Security**: See `SECURITY_POLICY.md`
