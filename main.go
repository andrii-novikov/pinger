package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/alexflint/go-arg"
)

const filename = "/tmp/.pinger"

var args struct {
	Addr        string        `arg:"-a" default:"8.8.8.8"`
	Token       string        `arg:"env:TELEGRAM_TOKEN,required"`
	ChatId      string        `arg:"env:TELEGRAM_CHAT_ID,required"`
	MinOffDelay time.Duration `arg:"-d" default:"60s"`
	Debug       bool          `arg:"-v" help:"enable debug logging"`
}

var okAt time.Time

func debugf(format string, v ...any) {
	if args.Debug {
		log.Printf(format, v...)
	}
}

func main() {
	log.SetFlags(log.Ltime)
	arg.MustParse(&args)

	debugf("pinging %s", args.Addr)

	command := fmt.Sprintf("ping -c 1 %s > /dev/null", args.Addr)
	_, err := exec.Command("/bin/sh", "-c", command).Output()

	if err != nil {
		debugf("ping failed: %v", err)
		os.Exit(1)
	}

	debugf("ping ok")

	okAt = loadOkAt()
	debugf("last ok at: %s (%.0fs ago)", okAt.Format(time.RFC3339), time.Since(okAt).Seconds())

	if time.Since(okAt) > args.MinOffDelay {
		downtime := time.Since(okAt)
		debugf("downtime %.0fs exceeds threshold %.0fs, sending telegram", downtime.Seconds(), args.MinOffDelay.Seconds())
		sendTelegramMessage(fmt.Sprintf("Internet is back. Downtime: %1s", downtime))
		log.Printf("telegram notification sent (downtime: %s)", downtime.Round(time.Second))
	}

	saveOkAt(time.Now())
	debugf("saved ok at")
}

// loadOkAt reads last ok time from file
func loadOkAt() time.Time {
	dat, err := os.ReadFile(filename)
	if err != nil {
		debugf("state file not found, using current time")
		return time.Now()
	}

	t, err := time.Parse(time.RFC3339, string(dat))

	if err != nil {
		panic(err)
	}

	return t
}

// saveOkAt writes last ok time to file
func saveOkAt(t time.Time) {
	err := os.WriteFile(filename, []byte(t.Format(time.RFC3339)), 0644)

	if err != nil {
		panic(err)
	}
}

// sendTelegramMessage sends message to telegram chat
func sendTelegramMessage(message string) {
	debugf("sending telegram message: %q", message)

	data := url.Values{
		"chat_id": {args.ChatId},
		"text":    {message},
	}

	resp, err := http.PostForm("https://api.telegram.org/bot"+args.Token+"/sendMessage", data)

	if err != nil {
		panic(err)
	}

	debugf("telegram response status: %s", resp.Status)
}
