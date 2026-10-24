# pinger

Monitors internet connectivity on a Raspberry Pi and sends a Telegram notification when the connection is restored after an outage.

## How it works

Each check pings an IP address. If the ping succeeds after a gap longer than `MIN_OFF_DELAY`, it sends a Telegram message with the downtime duration. The last-ok time is saved only after the message is sent, so a failed send is retried on the next check.

Two ways to run it:

- **Container (how the Pi runs it):** the `Dockerfile` builds an image that runs in interval mode (`PINGER_INTERVAL=60s`). It is the `pinger` service of the homelab repo's Pi stack. On push to `main`, CI tests, builds the image and dispatches the homelab **Deploy to Pi** workflow (`service=pinger`), which builds it on the Pi; that needs the `DEPLOY_TOKEN` secret.
- **Cron (one check per run):** leave `PINGER_INTERVAL` unset and use the steps below.

## Cron setup

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
| `PINGER_INTERVAL` | unset | Keep running and check every interval (`60s` or `60`); unset: one check, exit 1 if the ping fails |
| `DEBUG` | `false` | Enable verbose logging |
| `RASPBERRY_HOST` | — | SSH target, e.g. `pi@raspberrypi.local` |
| `CRON_SCHEDULE` | `* * * * *` | Cron schedule for the job |
| `PINGER_LOG` | `/var/log/pinger.log` | Log file path on the Pi |

## Logs

```bash
tail -f /var/log/pinger.log
```
