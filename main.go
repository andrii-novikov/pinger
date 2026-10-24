package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/alexflint/go-arg"
)

var filename = "/tmp/.pinger"

var httpClient = &http.Client{Timeout: 30 * time.Second}

// interval accepts a Go duration ("1m", "60s") or plain seconds ("60").
type interval time.Duration

func (i *interval) UnmarshalText(b []byte) error {
	if n, err := strconv.Atoi(string(b)); err == nil {
		*i = interval(time.Duration(n) * time.Second)
		return nil
	}
	d, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*i = interval(d)
	return nil
}

var args struct {
	Addr        string        `arg:"-a,env:PING_ADDR" default:"8.8.8.8"`
	Token       string        `arg:"env:TELEGRAM_TOKEN,required"`
	ChatId      string        `arg:"env:TELEGRAM_CHAT_ID,required"`
	MinOffDelay time.Duration `arg:"-d,env:MIN_OFF_DELAY" default:"70s"`
	Interval    interval      `arg:"-i,env:PINGER_INTERVAL" help:"keep running and check every interval (60s or 60); unset: check once and exit (cron)"`
	Debug       bool          `arg:"-v,env:DEBUG"`
}

var errPingFailed = errors.New("ping failed")

func debugf(format string, v ...any) {
	if args.Debug {
		log.Printf(format, v...)
	}
}

func main() {
	log.SetFlags(log.Ltime)
	arg.MustParse(&args)

	if args.Interval <= 0 {
		if err := check(); err != nil {
			if !errors.Is(err, errPingFailed) {
				log.Print(err)
			}
			os.Exit(1)
		}
		return
	}

	log.Printf("checking %s every %s", args.Addr, time.Duration(args.Interval))
	for {
		if err := check(); err != nil && !errors.Is(err, errPingFailed) {
			log.Print(err)
		}
		time.Sleep(time.Duration(args.Interval))
	}
}

// check pings once and, when the last successful ping is older than MinOffDelay, sends "Internet is back".
// The last-ok time is only saved after the message went out, so a failed send is retried on the next check.
func check() error {
	debugf("--- run %s ---", time.Now().Format(time.RFC3339))
	debugf("pinging %s", args.Addr)

	command := fmt.Sprintf("ping -c 1 %s > /dev/null", args.Addr)
	if _, err := exec.Command("/bin/sh", "-c", command).Output(); err != nil {
		debugf("ping failed: %v", err)
		return errPingFailed
	}

	debugf("ping ok")

	okAt := loadOkAt()
	debugf("last ok at: %s (%.0fs ago)", okAt.Format(time.RFC3339), time.Since(okAt).Seconds())

	if time.Since(okAt) > args.MinOffDelay {
		downtime := time.Since(okAt)
		debugf("downtime %.0fs exceeds threshold %.0fs, sending telegram", downtime.Seconds(), args.MinOffDelay.Seconds())
		if err := sendTelegramMessage(fmt.Sprintf("Internet is back. Downtime: %s", downtime.Round(time.Second))); err != nil {
			return err
		}
		log.Printf("telegram notification sent (downtime: %s)", downtime.Round(time.Second))
	}

	if err := saveOkAt(time.Now()); err != nil {
		return err
	}
	debugf("saved ok at")
	return nil
}

func loadOkAt() time.Time {
	dat, err := os.ReadFile(filename)
	if err != nil {
		debugf("state file not found, using current time")
		return time.Now()
	}

	t, err := time.Parse(time.RFC3339, string(dat))
	if err != nil {
		// A power cut mid-write can leave the file empty or truncated. Start over instead of failing every check.
		log.Printf("ignoring unreadable %s: %v", filename, err)
		return time.Now()
	}

	return t
}

func saveOkAt(t time.Time) error {
	if err := os.WriteFile(filename, []byte(t.Format(time.RFC3339)), 0644); err != nil {
		return fmt.Errorf("save %s: %w", filename, err)
	}
	return nil
}

func sendTelegramMessage(message string) error {
	debugf("sending telegram message: %q", message)

	data := url.Values{
		"chat_id": {args.ChatId},
		"text":    {message},
	}

	resp, err := httpClient.PostForm("https://api.telegram.org/bot"+args.Token+"/sendMessage", data)
	if err != nil {
		// A *url.Error carries the request URL, which holds the bot token: keep only the cause
		var uerr *url.Error
		if errors.As(err, &uerr) {
			err = uerr.Err
		}
		return fmt.Errorf("telegram: %w", err)
	}
	defer resp.Body.Close()

	debugf("telegram response status: %s", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: %s", resp.Status)
	}
	return nil
}
