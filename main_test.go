package main

import (
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
