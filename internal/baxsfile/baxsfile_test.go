package baxsfile

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	r := strings.NewReader(`
# this is a comment

cmd1: foo bar baz
cmd2: fooz barz bazz
`)
	procs, err := Parse(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(procs)
}
