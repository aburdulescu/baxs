package conf

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestCommentsAndWhitespaces(t *testing.T) {
	f, err := os.Open("testdata/comments_and_whitespaces.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var c Conf
	if err := c.Parse(f); err != nil {
		t.Fatal(err)
	}
}

func TestGetSections(t *testing.T) {
	r := bytes.NewBuffer([]byte(`
[x:a]
foo = bar
[x:b]
foo = bar
`))
	var c Conf
	if err := c.Parse(r); err != nil {
		t.Fatal(err)
	}
	s := c.GetSections(func(name string) bool {
		return strings.HasPrefix(name, "x:")
	})
	if len(s) != 2 {
		t.Fatal("")
	}
}

func TestKeysAndValues(t *testing.T) {
	f, err := os.Open("testdata/keys_and_values.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var c Conf
	if err := c.Parse(f); err != nil {
		t.Fatal(err)
	}
	{
		v, err := c.GetString("key1")
		if err != nil {
			t.Fatal(err)
		}
		if v != "23" {
			t.Fatal("wrong value:", v)
		}
	}
	{
		v, err := c.GetInt("key2")
		if err != nil {
			t.Fatal(err)
		}
		if v != 42 {
			t.Fatal("wrong value:", v)
		}
	}
	{
		v, err := c.GetBool("key3")
		if err != nil {
			t.Fatal(err)
		}
		if v != true {
			t.Fatal("wrong value:", v)
		}
	}
}

func TestSections(t *testing.T) {
	f, err := os.Open("testdata/sections.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var c Conf
	if err := c.Parse(f); err != nil {
		t.Fatal(err)
	}
	{
		s := c.GetSection("foo")
		if s == nil {
			t.Fatal("section not found")
		}
		v, err := s.GetString("k")
		if err != nil {
			t.Fatal(err)
		}
		if v != "v" {
			t.Fatal("wrong value:", v)
		}
	}
	{
		s := c.GetSection("bar")
		if s == nil {
			t.Fatal("section not found")
		}
		v, err := s.GetString("k")
		if err != nil {
			t.Fatal(err)
		}
		if v != "v" {
			t.Fatal("wrong value:", v)
		}
	}
}

func TestSectionTo(t *testing.T) {
	data := `
[foo]
s = dummy
i = 42
b = true
extra = xx
`
	buf := bytes.NewBufferString(data)

	var c Conf
	if err := c.Parse(buf); err != nil {
		t.Fatal(err)
	}

	s := c.GetSection("foo")

	dst := struct {
		S string `conf:"s"`
		I int    `conf:"i"`
		B bool   `conf:"b"`
		F float64
		f float64
	}{}

	if err := s.To(&dst); err != nil {
		t.Fatal(err)
	}

	if dst.S != "dummy" {
		t.Fatalf("wrong value for 'S': %v", dst.S)
	}
	if dst.I != 42 {
		t.Fatalf("wrong value for 'I': %v", dst.I)
	}
	if dst.B != true {
		t.Fatalf("wrong value for 'B': %v", dst.B)
	}
}

func readFile(f testing.TB, file string) []byte {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		f.Fatal(err)
	}
	return b
}

func FuzzConf(f *testing.F) {
	f.Add(readFile(f, "testdata/comments_and_whitespaces.conf"))
	f.Add(readFile(f, "testdata/keys_and_values.conf"))
	f.Add(readFile(f, "testdata/sections.conf"))
	var c Conf
	f.Fuzz(func(t *testing.T, data []byte) {
		buf := bytes.NewBuffer(data)
		c.Parse(buf)
		c.Reset()
	})
}

// TODO: benchmarks

func BenchmarkConf(b *testing.B) {
	data := []byte(`# comment

foo = bar

[sec1]
key = value
key2 = value
key = value
key2 = value
key = value
key2 = value
key = value
key2 = value

[sec2]
key = value
key2 = value
key = value
key2 = value
key = value
key2 = value


`)

	buf := bytes.NewBuffer(data)

	var c Conf

	for i := 0; i < b.N; i++ {
		c.Parse(buf)
		c.Reset()
	}
}
