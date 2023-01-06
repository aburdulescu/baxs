package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"bandr.me/p/baxs/internal/baxsfile"
	"bandr.me/p/baxs/internal/ipc"
)

type State uint8

const (
	Idle State = iota
	Failed
	Running
	Stopped
	Finished
)

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Failed:
		return "failed"
	case Running:
		return "running"
	case Stopped:
		return "stopped"
	case Finished:
		return "finished"
	default:
		return fmt.Sprintf("invalid state: %d", s)
	}
}

type Process struct {
	Name    string
	Command string
	Cmd     *exec.Cmd
	State   State
}

func (p *Process) UpdateState() {
	if p.Cmd == nil {
		return
	}
	if p.Cmd.ProcessState == nil {
		return
	}
	ps := p.Cmd.ProcessState
	p.State = Failed
	if ps.Success() {
		p.State = Finished
		return
	}
	ws, _ := ps.Sys().(syscall.WaitStatus)
	if ws.Signaled() {
		sig := ws.Signal()
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			p.State = Stopped
		}
	}
}

func (p *Process) Stop() {
	if p.Cmd == nil {
		return
	}
	if p.State != Running {
		return
	}
	if err := p.Cmd.Process.Kill(); err != nil {
		fmt.Printf("[daemon] failed to kill [%s]: %v\n", p.Name, err)
	}
}

type Table struct {
	mu    sync.Mutex
	procs []Process

	wg sync.WaitGroup

	logsDir string
}

func NewTable(logsDir string, entries []baxsfile.Entry) *Table {
	res := &Table{
		procs:   make([]Process, 0, len(entries)),
		logsDir: logsDir,
	}
	for _, e := range entries {
		res.procs = append(res.procs, Process{
			Name:    e.Name,
			Command: e.Command,
		})
	}
	return res
}

func (t *Table) Wait() {
	t.wg.Wait()
}

// try starting all procs, if one of them fails to start => exit
func (t *Table) StartAll() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.startAll()
}

func (t *Table) startAll() error {
	for i, p := range t.procs {
		if err := t.start(&t.procs[i]); err != nil {
			t.stopAll()
			return fmt.Errorf("[daemon] failed to start [%s]: %w", p.Name, err)
		}
	}
	return nil
}

func (t *Table) Start(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, p := range t.procs {
		if p.Name == name {
			return t.start(&t.procs[i])
		}
	}
	return fmt.Errorf("cannot find service with name: %v", name)
}

func (t *Table) start(p *Process) error {
	if p.State == Running {
		fmt.Printf("[%s] already running\n", p.Name)
		return nil
	}

	fmt.Printf("[%s] starting with command '%s'\n", p.Name, p.Command)

	logfilePath := filepath.Join(t.logsDir, p.Name+".log")
	logfile, err := os.OpenFile(logfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	args := strings.Split(p.Command, " ")

	// TODO: maybe fail if args[0] is shell?

	// #nosec G204 -- don't care right now, command comes from user configured file
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Printf("[%s] started with pid %v\n", p.Name, cmd.Process.Pid)

	p.State = Running

	p.Cmd = cmd

	t.wg.Add(1)

	go func() {
		defer logfile.Close()
		defer t.wg.Done()
		if err := p.Cmd.Wait(); err != nil {
			fmt.Printf("[%s] error: %v\n", p.Name, err)
		} else {
			fmt.Printf("[%s] done\n", p.Name)
		}
		p.UpdateState()
	}()

	return nil
}

func (t *Table) StopAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.stopAll()
}

func (t *Table) stopAll() {
	for _, p := range t.procs {
		p.Stop()
	}
}

func (t *Table) Stop(name string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, p := range t.procs {
		if p.Name == name {
			p.Stop()
			return nil
		}
	}
	return fmt.Errorf("cannot find service with name: %v", name)
}

func (t *Table) Ps() []ipc.PsResult {
	t.mu.Lock()
	defer t.mu.Unlock()
	var result []ipc.PsResult
	for _, p := range t.procs {
		result = append(result, ipc.PsResult{
			Name:   p.Name,
			Status: p.State.String(),
		})
	}
	return result
}
