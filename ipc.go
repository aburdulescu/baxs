package main

import (
	"io"
	"log"
	"net"
	"os"
)

type IPCServer struct {
	l net.Listener
}

const daemonSocketFile = "/tmp/baxs.sock"

func newIPCServer() (*IPCServer, error) {
	os.Remove(daemonSocketFile)
	l, err := net.Listen("unix", daemonSocketFile)
	if err != nil {
		return nil, err
	}
	s := &IPCServer{
		l: l,
	}
	return s, nil
}

func (s IPCServer) Close() error {
	return s.l.Close()
}

func (s IPCServer) start() {
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
	log.Println(n, string(b))
}
