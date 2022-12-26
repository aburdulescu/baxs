package ipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type Daemon struct {
	l net.Listener
}

const daemonSocketFile = "/tmp/baxs.sock"

func NewDaemon() (*Daemon, error) {
	os.Remove(daemonSocketFile)
	l, err := net.Listen("unix", daemonSocketFile)
	if err != nil {
		return nil, err
	}
	s := &Daemon{
		l: l,
	}
	return s, nil
}

func (d Daemon) Close() error {
	return d.l.Close()
}

func (d Daemon) Start() {
	defer d.Close()
	for {
		conn, err := d.l.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	b := make([]byte, 4096)
	n, err := conn.Read(b)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Println(err)
		return
	}

	var req Request
	if err := json.Unmarshal(b[:n], &req); err != nil {
		log.Println(err)
		return
	}

	log.Println(req)

	var rsp Response
	switch req.Op {
	case OpLs:
		rsp = Response{
			Data: []string{
				"foo",
				"bar",
				"baz",
			},
		}
	default:
		rsp = Response{
			Err: fmt.Sprintf("unknown op %s(%d)", req.Op, req.Op),
		}
	}

	if err := json.NewEncoder(conn).Encode(rsp); err != nil {
		log.Println(err)
		return
	}
}

type Op uint8

const (
	OpLs Op = iota
)

func (op Op) String() string {
	switch op {
	case OpLs:
		return "ls"
	default:
		return "unknown"
	}
}

type Request struct {
	Op   Op
	Data any `json:",omitempty"`
}

type Response struct {
	Err  string
	Data any `json:",omitempty"`
}

func execRequest(req Request) (*Response, error) {
	conn, err := net.Dial("unix", daemonSocketFile)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return nil, err
	}
	if _, err := io.Copy(conn, buf); err != nil {
		return nil, err
	}
	buf.Reset()
	if _, err := io.Copy(buf, conn); err != nil {
		return nil, err
	}
	var rsp Response
	if err := json.NewDecoder(buf).Decode(&rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

func Ls() ([]string, error) {
	rsp, err := execRequest(Request{Op: OpLs})
	if err != nil {
		return nil, err
	}
	if rsp.Err != "" {
		return nil, errors.New(rsp.Err)
	}
	data, ok := rsp.Data.([]any)
	if !ok {
		return nil, errors.New("response data is not a slice")
	}
	res := make([]string, 0, len(data))
	for _, v := range data {
		res = append(res, v.(string))
	}
	return res, nil
}
