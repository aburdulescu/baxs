package conf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const example = `
# this is a comment

# this is a key-value entry
key = value

# keys have one type: string, no quotes

# values can be of the following types: string(e.g. "foo"), integer(e.g. 42), bool(i.e. true/false)

# so, these are valid entries:
path = "/foo/bar"
answer = 42
is_simple = true

# that's it!
`

// tool is started like this:
// baxs daemon --conf /etc/baxs/daemon.conf --services /home/user/baxs_services

const example_daemon = `
# service directory
service_dir = "/home/me/cmds"

# base path for logs
log_dir = "/var/log/my_daemon"

# unix domain socket path
socket = "/var/run/my_daemon.sock"

# user to run as
user = "foo"
`

// name given by filename
const example_service = `
command = nc -l localhost 12345
restart = true
`

// TODO: pass struct and use reflection(e.g. like in encoding/json)?
// TODO: add support for sections
type Conf struct {
	global Section

	names    []string
	sections []Section
}

type Section struct {
	keys   []string
	types  []valueType
	values []interface{}
}

func (s *Section) reset() {
	s.keys = s.keys[:]
	s.types = s.types[:]
	s.values = s.values[:]
}

func (s *Section) append(k string, vt valueType, v interface{}) {
	s.keys = append(s.keys, k)
	s.types = append(s.types, vt)
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
	v, ok := s.values[i].(string)
	if !ok {
		return "", ErrWrongType
	}
	return v, nil
}

func (s Section) GetInt(key string) (int, error) {
	i := s.find(key)
	if i == -1 {
		return 0, ErrKeyNotFound
	}
	v, ok := s.values[i].(int)
	if !ok {
		return 0, ErrWrongType
	}
	return v, nil
}

func (s Section) GetBool(key string) (bool, error) {
	i := s.find(key)
	if i == -1 {
		return false, ErrKeyNotFound
	}
	v, ok := s.values[i].(bool)
	if !ok {
		return false, ErrWrongType
	}
	return v, nil
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

var errValueStringWithoutEndQuote = errors.New("missing closing quote for string value")

func parseValue(v string) (valueType, interface{}, error) {
	// bool
	if v == "true" {
		return valueTypeBool, true, nil
	}
	if v == "false" {
		return valueTypeBool, false, nil
	}

	// string
	if v[0] == '"' {
		if len(v) == 1 || v[len(v)-1] != '"' {
			return valueTypeUnkown, nil, errValueStringWithoutEndQuote
		}
		return valueTypeString, v[1 : len(v)-1], nil
	}

	// int
	i, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		return valueTypeUnkown, nil, fmt.Errorf("value conversion to int failed: %v", err)
	}
	return valueTypeInt, int(i), nil
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
		//log.Println(i, data[i])
		switch data[i] {
		case '#':
			for ; i < len(data) && data[i] != '\n'; i++ {
			}
			//log.Println("comment => skip line", i, n)
		case '\r':
			i++
			//log.Println("carriage return => move one char", i, n)
		case '\n':
			i++
			n++
			//			log.Println("newline => move one char and incr line count", i, n)
		case '[':
			ii, section, err := parseSection(data, i)
			if err != nil {
				return fmt.Errorf("line %d: %v", n, err)
			}
			i = ii
			c.append(section)
			section_index = len(c.sections) - 1
		default:
			//			log.Println("other => key value line", i, n)
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
			vt, vv, err := parseValue(v)
			if err != nil {
				return fmt.Errorf("line %d: %v", n, err)
			}

			if section_index == -1 {
				c.global.append(k, vt, vv)
			} else {
				c.sections[section_index].append(k, vt, vv)
			}

			//log.Printf("%d: key=%s, value=%s,%v\n", n, k, vt, vv)

			i = end
		}
	}
	return nil
}

type valueType uint8

const (
	valueTypeUnkown valueType = iota
	valueTypeString
	valueTypeInt
	valueTypeBool
)

func (t valueType) String() string {
	switch t {
	case valueTypeString:
		return "string"
	case valueTypeInt:
		return "int"
	case valueTypeBool:
		return "bool"
	default:
		return "unknown"
	}
}

var ErrKeyNotFound = errors.New("key not found")
var ErrWrongType = errors.New("value has different type than the requested one")

func (c Conf) GetSection(name string) *Section {
	for i, e := range c.names {
		if e == name {
			return &c.sections[i]
		}
	}
	return nil
}

func (c Conf) GetString(key string) (string, error) {
	i := c.global.find(key)
	if i == -1 {
		return "", ErrKeyNotFound
	}
	v, ok := c.global.values[i].(string)
	if !ok {
		return "", ErrWrongType
	}
	return v, nil
}

func (c Conf) GetInt(key string) (int, error) {
	i := c.global.find(key)
	if i == -1 {
		return 0, ErrKeyNotFound
	}
	v, ok := c.global.values[i].(int)
	if !ok {
		return 0, ErrWrongType
	}
	return v, nil
}

func (c Conf) GetBool(key string) (bool, error) {
	i := c.global.find(key)
	if i == -1 {
		return false, ErrKeyNotFound
	}
	v, ok := c.global.values[i].(bool)
	if !ok {
		return false, ErrWrongType
	}
	return v, nil
}
