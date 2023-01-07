package ipc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
)

const SocketAddr = "/tmp/baxs.sock"

type Op uint8

const (
	OpPs Op = iota
	OpStop
	OpStart
)

func (op Op) String() string {
	switch op {
	case OpPs:
		return "ps"
	case OpStop:
		return "stop"
	case OpStart:
		return "start"
	default:
		return "unknown"
	}
}

type Request struct {
	Names []string `json:",omitempty"`
	Op    Op
}

type Response struct {
	Err string

	// for ps
	Ps []PsResult `json:",omitempty"`
}

type PsResult struct {
	Name   string
	Status string
	Pid    int
}

func Ps() ([]PsResult, error) {
	rsp, err := execRequest(Request{Op: OpPs})
	if err != nil {
		return nil, err
	}
	if rsp.Err != "" {
		return nil, errors.New(rsp.Err)
	}
	return rsp.Ps, nil
}

func Stop(names ...string) error {
	rsp, err := execRequest(Request{Op: OpStop, Names: names})
	if err != nil {
		return err
	}
	if rsp.Err != "" {
		return errors.New(rsp.Err)
	}
	return nil
}

func Start(names ...string) error {
	rsp, err := execRequest(Request{Op: OpStart, Names: names})
	if err != nil {
		return err
	}
	if rsp.Err != "" {
		return errors.New(rsp.Err)
	}
	return nil
}

func execRequest(req Request) (*Response, error) {
	conn, err := net.Dial("unix", SocketAddr)
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
