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
service_dir = /home/me/cmds

# base path for logs
log_dir = /var/log/my_daemon

# unix domain socket path
socket = /var/run/my_daemon.sock

# user to run as
user = foo
`

// name given by filename
const example_service = `
command = nc -l localhost 12345
restart = true
`

// TODO: pass struct and use reflection(e.g. like in encoding/json)?
type Conf struct {
	keys   []string
	types  []valueType
	values []interface{}
}

func Parse(r io.Reader) (*Conf, error) {
	var b bytes.Buffer
	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	var c Conf
	if err := c.parse(b.Bytes()); err != nil {
		return nil, err
	}
	return &c, nil
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

func (c *Conf) parse(data []byte) error {
	n := 1
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

			c.append(k, vt, vv)

			//log.Printf("%d: key=%s, value=%s,%v\n", n, k, vt, vv)

			i = end
		}
	}
	return nil
}

func (c *Conf) append(k string, vt valueType, v interface{}) {
	c.keys = append(c.keys, k)
	c.types = append(c.types, vt)
	c.values = append(c.values, v)
}

func (c *Conf) find(k string) int {
	for i, e := range c.keys {
		if e == k {
			return i
		}
	}
	return -1
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

func (c *Conf) GetString(key string) (string, error) {
	i := c.find(key)
	if i == -1 {
		return "", ErrKeyNotFound
	}
	v, ok := c.values[i].(string)
	if !ok {
		return "", ErrWrongType
	}
	return v, nil
}

func (c *Conf) GetInt(key string) (int, error) {
	i := c.find(key)
	if i == -1 {
		return 0, ErrKeyNotFound
	}
	v, ok := c.values[i].(int)
	if !ok {
		return 0, ErrWrongType
	}
	return v, nil
}

func (c *Conf) GetBool(key string) (bool, error) {
	i := c.find(key)
	if i == -1 {
		return false, ErrKeyNotFound
	}
	v, ok := c.values[i].(bool)
	if !ok {
		return false, ErrWrongType
	}
	return v, nil
}
