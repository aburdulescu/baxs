package main

import (
	"strings"
	"testing"
)

func TestParseBaxsfile(t *testing.T) {
	r := strings.NewReader(`
# this is a comment

cmd1: foo bar baz
cmd2: fooz barz bazz
`)
	procs, err := parseBaxsfile(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(procs)
}
