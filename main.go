package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/alexflint/go-arg"
)

const filename = ".okat"

var args struct {
	Addr        string        `arg:"-a" default:"8.8.8.8"`
	Timeout     int           `arg:"-t" default:"2"`
	Token       string        `arg:"env:TELEGRAM_TOKEN,required"`
	ChatId      string        `arg:"env:TELEGRAM_CHAT_ID,required"`
	MinOffDelay time.Duration `arg:"-d" default:"60s"`
}

var okAt time.Time

func main() {
	arg.MustParse(&args)

	command := fmt.Sprintf("ping -c 1 -t %d %s > /dev/null", args.Timeout, args.Addr)
	_, err := exec.Command("/bin/sh", "-c", command).Output()

	if err != nil {
		os.Exit(1)
	}

	okAt = loadOkAt()

	if time.Since(okAt) > args.MinOffDelay {
		sendTelegramMessage(fmt.Sprintf("Internet is back. Downtime: %1s", time.Since(okAt)))
	}

	saveOkAt(time.Now())
}

// loadOkAt reads last ok time from file
func loadOkAt() time.Time {
	dat, err := os.ReadFile(filename)
	if err != nil {
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
	data := url.Values{
		"chat_id": {args.ChatId},
		"text":    {message},
	}

	_, err := http.PostForm("https://api.telegram.org/bot"+args.Token+"/sendMessage", data)

	if err != nil {
		panic(err)
	}
}
