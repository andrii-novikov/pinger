# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build        # compile for linux/arm (Raspberry Pi) and darwin/amd64
make deploy       # build + scp binary to Raspberry Pi
make enable-cron  # install/update cron job on Raspberry Pi via SSH
```

Run locally (requires env vars):
```bash
TELEGRAM_TOKEN=x TELEGRAM_CHAT_ID=y go run . -v
```

## Architecture

Single-file Go program (`main.go`) designed to run every minute via cron on a Raspberry Pi. It pings an IP address and sends a Telegram message when internet connectivity is restored after a sustained outage.

**Flow:**
1. Ping `PING_ADDR` — exit 1 on failure (cron will log it, no notification)
2. Load last-ok timestamp from `/tmp/.pinger`
3. If `time.Since(lastOk) > MIN_OFF_DELAY` → send Telegram "Internet is back" message
4. Save current time to `/tmp/.pinger`

`MIN_OFF_DELAY` defaults to `70s` — intentionally above the 60s cron interval to avoid false positives from timing jitter.

## Configuration

All config is via environment variables (set in `.env`, loaded by the Makefile):

| Var | Default | Description |
|-----|---------|-------------|
| `TELEGRAM_TOKEN` | required | Telegram bot token |
| `TELEGRAM_CHAT_ID` | required | Telegram chat ID |
| `PING_ADDR` | `8.8.8.8` | Address to ping |
| `MIN_OFF_DELAY` | `70s` | Min downtime before notification |
| `DEBUG` | `false` | Enable verbose logging |
| `RASPBERRY_HOST` | — | SSH target, e.g. `pi@raspberrypi.local` |
| `CRON_SCHEDULE` | `* * * * *` | Cron expression |
| `PINGER_LOG` | `/var/log/pinger.log` | Log file path on the Pi |

Copy `.env.example` to `.env` and fill in values before deploying.

## Deployment

The cron entry installed by `make enable-cron` passes all env vars inline (cron doesn't inherit environment) and appends stdout/stderr to `PINGER_LOG`. On the Pi, create the log file first:

```bash
sudo touch /var/log/pinger.log && sudo chown pi /var/log/pinger.log
```

View logs: `tail -f /var/log/pinger.log` or `journalctl -u cron -f`.
