package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
)

type IPCDaemon struct {
	l net.Listener
}

const daemonSocketFile = "/tmp/baxs.sock"

func newIPCDaemon() (*IPCDaemon, error) {
	os.Remove(daemonSocketFile)
	l, err := net.Listen("unix", daemonSocketFile)
	if err != nil {
		return nil, err
	}
	s := &IPCDaemon{
		l: l,
	}
	return s, nil
}

func (s IPCDaemon) Close() error {
	return s.l.Close()
}

func (s IPCDaemon) start() {
	defer s.Close()
	for {
		conn, err := s.l.Accept()
		if err != nil {
			log.Println(err)
			break
		}
		go handleIPCConn(conn)
	}
}

func handleIPCConn(conn net.Conn) {
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
	var req IPCRequest
	if err := json.Unmarshal(b[:n], &req); err != nil {
		log.Println(err)
		return
	}
	log.Println(req)
	rsp := IPCResponse{
		Err: "here be dragons",
	}
	if err := json.NewEncoder(conn).Encode(rsp); err != nil {
		log.Println(err)
		return
	}
}

type IPCRequest struct {
	Cmd  string
	Data any
}

type IPCResponse struct {
	Err  string
	Data any
}

func execIPCRequest(req IPCRequest) (*IPCResponse, error) {
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
	var rsp IPCResponse
	if err := json.NewDecoder(buf).Decode(&rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}
