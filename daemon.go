package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"baxs/ipc"
)

type Daemon struct {
	procs       ProcessTable
	ipcListener net.Listener
}

func NewDaemon(logsDir, baxsfilePath string) (*Daemon, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Open(baxsfilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	procs, err := parseBaxsfile(f)
	if err != nil {
		return nil, err
	}

	ipcListener, err := NewIpcListener()
	if err != nil {
		return nil, err
	}

	d := &Daemon{
		procs: ProcessTable{
			procs:   procs,
			logsDir: logsDir,
		},

		ipcListener: ipcListener,
	}

	return d, nil
}

func (d *Daemon) Run() error {
	defer d.procs.wg.Wait()

	if err := d.procs.StartAll(); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("[daemon] termination signal received")
		d.procs.StopAll()
		d.ipcListener.Close()
	}()

	d.startIpcListener()

	return nil
}

func NewIpcListener() (net.Listener, error) {
	os.Remove(ipc.SocketAddr)
	l, err := net.Listen("unix", ipc.SocketAddr)
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (d *Daemon) startIpcListener() {
	defer d.ipcListener.Close()
	for {
		conn, err := d.ipcListener.Accept()
		if err != nil {
			fmt.Printf("[ipcListener] accept error: %v\n", err)
			break
		}
		go d.handleIpcConn(conn)
	}
}

func (d *Daemon) handleIpcConn(conn net.Conn) {
	defer conn.Close()

	b := make([]byte, 4096)
	n, err := conn.Read(b)
	if err != nil {
		if err == io.EOF {
			return
		}
		fmt.Printf("[ipcConn] error: %v\n", err)
		return
	}

	var req ipc.Request
	if err := json.Unmarshal(b[:n], &req); err != nil {
		fmt.Printf("[ipcConn] error: %v\n", err)
		return
	}

	fmt.Printf("[ipcConn] new request: %v\n", req)

	rsp := ipc.Response{}
	switch req.Op {
	case ipc.OpLs:
		rsp.Data = d.procs.Ls()
	case ipc.OpStop:
		names, ok := req.Data.([]any)
		if !ok {
			rsp.Err = "not a []any"
		} else {
			if len(names) == 0 {
				d.procs.StopAll()
			} else {
				for _, name := range names {
					if err := d.procs.Stop(name.(string)); err != nil {
						rsp.Err = err.Error()
					}
				}
			}
		}
	case ipc.OpStart:
		names, ok := req.Data.([]any)
		if !ok {
			rsp.Err = "not a []any"
		} else {
			if len(names) == 0 {
				if err := d.procs.StartAll(); err != nil {
					rsp.Err = err.Error()
				}
			} else {
				for _, name := range names {
					if err := d.procs.Start(name.(string)); err != nil {
						rsp.Err = err.Error()
					}
				}
			}
		}
	default:
		rsp.Err = fmt.Sprintf("unknown op %s(%d)", req.Op, req.Op)
	}

	if err := json.NewEncoder(conn).Encode(rsp); err != nil {
		fmt.Printf("[ipcConn] error: %v\n", err)
		return
	}
}

func parseBaxsfile(r io.Reader) ([]Process, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var procs []Process
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
			return nil, fmt.Errorf("failed to parse baxfile: line %d is missing :", i+1)
		}
		procs = append(procs, Process{
			Name:    line[:dot],
			Command: strings.Trim(line[dot+1:], " \t"),
		})
	}
	return procs, nil
}
