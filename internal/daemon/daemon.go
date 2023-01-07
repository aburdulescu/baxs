package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bandr.me/p/baxs/internal/baxsfile"
	"bandr.me/p/baxs/internal/ipc"
	"bandr.me/p/baxs/internal/process"
)

type Daemon struct {
	ptable      *process.Table
	ipcListener net.Listener
}

func New(logsDir, baxsfilePath string) (*Daemon, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}

	f, err := os.Open(baxsfilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	baxsfileEntries, err := baxsfile.Parse(f)
	if err != nil {
		return nil, err
	}

	ipcListener, err := newIpcListener()
	if err != nil {
		return nil, err
	}

	d := &Daemon{
		ptable:      process.NewTable(logsDir, baxsfileEntries),
		ipcListener: ipcListener,
	}

	return d, nil
}

func (d *Daemon) Run() error {
	defer d.ptable.Wait()

	if err := d.ptable.StartAll(); err != nil {
		return err
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("[daemon] termination signal received")
		d.ptable.StopAll()
		d.ipcListener.Close()
	}()

	d.startIpcListener()

	return nil
}

func newIpcListener() (net.Listener, error) {
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
		if errors.Is(err, io.EOF) {
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
	case ipc.OpPs:
		rsp.Ps = d.ptable.Ps()
	case ipc.OpStop:
		if len(req.Names) == 0 {
			d.ptable.StopAll()
		} else {
			for _, name := range req.Names {
				if err := d.ptable.Stop(name); err != nil {
					rsp.Err = err.Error()
				}
			}
		}
	case ipc.OpStart:
		if len(req.Names) == 0 {
			if err := d.ptable.StartAll(); err != nil {
				rsp.Err = err.Error()
			}
		} else {
			for _, name := range req.Names {
				if err := d.ptable.Start(name); err != nil {
					rsp.Err = err.Error()
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
