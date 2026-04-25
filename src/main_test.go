package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestUrlValue(t *testing.T) {
	link := "http://345h34j5h.ololo.com/playlists/uplist/439rth06843jknekwbrgjermg/playlist.m3u8"
	values, err := url.Parse(link)

	if err != nil {
		t.Fatal("failed to parse link")
	}

	args := strings.Split(values.Host, ".")
	host := strings.Join(args[1:], ".")

	if host != "ololo.com" {
		t.Fatal("failed to get host")
	}

}
