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

func FuzzParse(f *testing.F) {
	f.Add([]byte(`
    #   akakakka

# valid entry
foo: bar baz

# valid entry with leading whitespace
  \t  foo: bar baz

           \t\n






`))

	f.Fuzz(func(t *testing.T, data []byte) {
		t.Log(Parse(strings.NewReader(string(data))))
	})
}
