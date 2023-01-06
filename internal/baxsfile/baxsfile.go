package baxsfile

import (
	"fmt"
	"io"
	"strings"
)

type Entry struct {
	Name    string
	Command string
}

func Parse(r io.Reader) ([]Entry, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.Trim(line, " \t")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		dot := strings.Index(line, ":")
		if dot == -1 {
			return nil, fmt.Errorf("failed to parse baxfile: line %d is missing ':'", i+1)
		}
		entries = append(entries, Entry{
			Name:    line[:dot],
			Command: strings.Trim(line[dot+1:], " \t"),
		})
	}
	return entries, nil
}
