package conf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// TODO: pass struct and use reflection(e.g. like in encoding/json)?
type Conf struct {
	global Section

	names    []string
	sections []Section
}

type Section struct {
	keys   []string
	values []string
}

func (s *Section) reset() {
	s.keys = s.keys[:]
	s.values = s.values[:]
}

func (s *Section) append(k string, v string) {
	s.keys = append(s.keys, k)
	s.values = append(s.values, v)
}

func (s Section) find(k string) int {
	for i, e := range s.keys {
		if e == k {
			return i
		}
	}
	return -1
}

func (s Section) GetString(key string) (string, error) {
	i := s.find(key)
	if i == -1 {
		return "", ErrKeyNotFound
	}
	return s.values[i], nil
}

func (s Section) GetInt(key string) (int, error) {
	i := s.find(key)
	if i == -1 {
		return 0, ErrKeyNotFound
	}
	v := s.values[i]
	vi, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("cannot convert '%s' to int", v)
	}
	return int(vi), nil
}

func (s Section) GetBool(key string) (bool, error) {
	i := s.find(key)
	if i == -1 {
		return false, ErrKeyNotFound
	}
	v := s.values[i]
	switch v {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("cannot convert '%s' to bool", v)
	}
}

func (c *Conf) Parse(r io.Reader) error {
	var b bytes.Buffer
	if _, err := io.Copy(&b, r); err != nil {
		return err
	}
	if err := c.parse(b.Bytes()); err != nil {
		return err
	}
	return nil
}

func (c *Conf) Reset() {
	for _, s := range c.sections {
		s.reset()
	}

}

const whitespaces = " \t"

func parseKeyValLine(line []byte) (string, string, error) {
	k, v, found := strings.Cut(string(line), "=")
	if !found {
		return "", "", errors.New("missing sparator")
	}
	k = strings.Trim(k, whitespaces)
	v = strings.Trim(v, whitespaces)
	if len(k) == 0 {
		return "", "", errors.New("empty key")
	}
	if len(v) == 0 {
		return "", "", errors.New("empty value")
	}
	return k, v, nil
}

func parseSection(data []byte, i int) (int, string, error) {
	start := i + 1
	for ; i < len(data); i++ {
		switch data[i] {
		case ']':
			return i + 1, string(data[start:i]), nil
		case '\n':
			return i, "", fmt.Errorf("missing section close brace")
		}
	}
	// end is reached
	return i, "", fmt.Errorf("missing section close brace")
}

func (c *Conf) append(name string) {
	c.names = append(c.names, name)
	c.sections = append(c.sections, Section{})
}

func (c *Conf) reset() {
	c.names = c.names[:]
	c.sections = c.sections[:]
}

func (c *Conf) parse(data []byte) error {
	n := 1
	section_index := -1
	for i := 0; i < len(data); {
		for ; i < len(data) && (data[i] == ' ' || data[i] == '\t'); i++ {
		}
		if i == len(data) {
			break
		}
		switch data[i] {
		case '#':
			for ; i < len(data) && data[i] != '\n'; i++ {
			}
		case '\r':
			i++
		case '\n':
			i++
			n++
		case '[':
			ii, section, err := parseSection(data, i)
			if err != nil {
				return fmt.Errorf("line %d: %v", n, err)
			}
			i = ii
			c.append(section)
			section_index = len(c.sections) - 1
		default:
			end := strings.Index(string(data[i:]), "\n")
			if end == -1 {
				end = len(data)
			} else {
				end += i
			}
			line := data[i:end]
			k, v, err := parseKeyValLine(line)
			if err != nil {
				return fmt.Errorf("line %d: %v", n, err)
			}
			if section_index == -1 {
				c.global.append(k, v)
			} else {
				c.sections[section_index].append(k, v)
			}
			i = end
		}
	}
	return nil
}

var ErrKeyNotFound = errors.New("key not found")

func (c Conf) GetSection(name string) *Section {
	for i, e := range c.names {
		if e == name {
			return &c.sections[i]
		}
	}
	return nil
}

func (c Conf) GetString(key string) (string, error) {
	return c.global.GetString(key)
}

func (c Conf) GetInt(key string) (int, error) {
	return c.global.GetInt(key)
}

func (c Conf) GetBool(key string) (bool, error) {
	return c.global.GetBool(key)
}
