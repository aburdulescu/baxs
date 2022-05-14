package conf

import (
	"bytes"
	"io/ioutil"
	"os"
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
		t.Fatal("copy failed:", err)
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

func readFile(f *testing.F, file string) []byte {
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
