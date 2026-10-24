package main

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntervalUnmarshalText(t *testing.T) {
	cases := map[string]time.Duration{
		"60":  60 * time.Second,
		"60s": 60 * time.Second,
		"2m":  2 * time.Minute,
		"0":   0,
	}
	for in, want := range cases {
		var i interval
		if err := i.UnmarshalText([]byte(in)); err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if time.Duration(i) != want {
			t.Errorf("%q: got %s, want %s", in, time.Duration(i), want)
		}
	}

	var i interval
	if err := i.UnmarshalText([]byte("soon")); err == nil {
		t.Error(`"soon": expected an error`)
	}
}

func TestLoadOkAtSurvivesBadStateFile(t *testing.T) {
	filename = filepath.Join(t.TempDir(), "state")
	for _, content := range []string{"", "2026-10-2"} {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		if got := loadOkAt(); time.Since(got) > time.Minute {
			t.Errorf("%q: got %s, want about now", content, got)
		}
	}

	ok := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := saveOkAt(ok); err != nil {
		t.Fatal(err)
	}
	if got := loadOkAt(); !got.Equal(ok) {
		t.Errorf("got %s, want %s", got, ok)
	}
}

func TestTelegramErrorHidesToken(t *testing.T) {
	args.Token = "123:secret"
	httpClient.Transport = failingTransport{}
	defer func() { httpClient.Transport = nil }()

	err := sendTelegramMessage("hi")
	if err == nil || strings.Contains(err.Error(), args.Token) {
		t.Fatalf("got %v, want an error without the token", err)
	}
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("dial tcp: network is unreachable")
}
