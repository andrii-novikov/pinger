# pinger

Monitors internet connectivity on a Raspberry Pi and sends a Telegram notification when the connection is restored after an outage.

## How it works

Runs every minute via cron. On each run it pings an IP address — if the ping succeeds after a gap longer than `MIN_OFF_DELAY`, it sends a Telegram message with the downtime duration.

## Setup

1. Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

2. Create the log file on the Pi:

```bash
ssh pi@raspberrypi.local 'sudo touch /var/log/pinger.log && sudo chown pi /var/log/pinger.log'
```

3. Deploy and enable the cron job:

```bash
make deploy
make enable-cron
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_TOKEN` | required | Telegram bot token |
| `TELEGRAM_CHAT_ID` | required | Telegram chat ID |
| `PING_ADDR` | `8.8.8.8` | Address to ping |
| `MIN_OFF_DELAY` | `70s` | Minimum downtime before notification |
| `DEBUG` | `false` | Enable verbose logging |
| `RASPBERRY_HOST` | — | SSH target, e.g. `pi@raspberrypi.local` |
| `CRON_SCHEDULE` | `* * * * *` | Cron schedule for the job |
| `PINGER_LOG` | `/var/log/pinger.log` | Log file path on the Pi |

## Logs

```bash
tail -f /var/log/pinger.log
```
