package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"bandr.me/p/baxs/ipc"
)

type ProcessState uint8

const (
	ProcessIdle ProcessState = iota
	ProcessFailed
	ProcessRunning
	ProcessStopped
	ProcessFinished
)

func (s ProcessState) String() string {
	switch s {
	case ProcessIdle:
		return "idle"
	case ProcessFailed:
		return "failed"
	case ProcessRunning:
		return "running"
	case ProcessStopped:
		return "stopped"
	case ProcessFinished:
		return "finished"
	default:
		return fmt.Sprintf("invalid state: %d", s)
	}
}

type Process struct {
	Name    string
	Command string
	Cmd     *exec.Cmd
	State   ProcessState
}

func (p *Process) UpdateState() {
	if p.Cmd == nil {
		return
	}
	if p.Cmd.ProcessState == nil {
		return
	}
	ps := p.Cmd.ProcessState
	p.State = ProcessFailed
	if ps.Success() {
		p.State = ProcessFinished
		return
	}
	ws := ps.Sys().(syscall.WaitStatus)
	if ws.Signaled() {
		sig := ws.Signal()
		if sig == syscall.SIGTERM || sig == syscall.SIGINT {
			p.State = ProcessStopped
		}
	}
}

func (p *Process) Stop() {
	if p.Cmd == nil {
		return
	}
	if p.State != ProcessRunning {
		return
	}
	if err := p.Cmd.Process.Kill(); err != nil {
		fmt.Printf("[daemon] failed to kill [%s]: %v\n", p.Name, err)
	}
}

type ProcessTable struct {
	mu    sync.Mutex
	procs []Process

	wg sync.WaitGroup

	logsDir string
}

// try starting all procs, if one of them fails to start => exit
func (pt *ProcessTable) StartAll() error {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	return pt.startAll()
}

func (pt *ProcessTable) startAll() error {
	for i, p := range pt.procs {
		if err := pt.start(&pt.procs[i]); err != nil {
			pt.stopAll()
			return fmt.Errorf("[daemon] failed to start [%s]: %v", p.Name, err)
		}
	}
	return nil
}

func (pt *ProcessTable) Start(name string) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	for i, p := range pt.procs {
		if p.Name == name {
			return pt.start(&pt.procs[i])
		}
	}
	return fmt.Errorf("cannot find service with name: %v", name)
}

func (pt *ProcessTable) start(p *Process) error {
	if p.State == ProcessRunning {
		fmt.Printf("[%s] already running\n", p.Name)
		return nil
	}

	fmt.Printf("[%s] starting with command '%s'\n", p.Name, p.Command)

	logfilePath := filepath.Join(pt.logsDir, p.Name+".log")
	logfile, err := os.OpenFile(logfilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	args := strings.Split(p.Command, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Printf("[%s] started with pid %v\n", p.Name, cmd.Process.Pid)

	p.State = ProcessRunning

	p.Cmd = cmd

	pt.wg.Add(1)

	go func() {
		defer logfile.Close()
		defer pt.wg.Done()
		if err := p.Cmd.Wait(); err != nil {
			fmt.Printf("[%s] error: %v\n", p.Name, err)
		} else {
			fmt.Printf("[%s] done\n", p.Name)
		}
		p.UpdateState()
	}()

	return nil
}

func (pt *ProcessTable) StopAll() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.stopAll()
}

func (pt *ProcessTable) stopAll() {
	for _, p := range pt.procs {
		p.Stop()
	}
}

func (pt *ProcessTable) Stop(name string) error {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	for _, p := range pt.procs {
		if p.Name == name {
			p.Stop()
			return nil
		}
	}
	return fmt.Errorf("cannot find service with name: %v", name)
}

func (pt *ProcessTable) Ls() []ipc.LsResult {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	var result []ipc.LsResult
	for _, p := range pt.procs {
		result = append(result, ipc.LsResult{
			Name:   p.Name,
			Status: p.State.String(),
		})
	}
	return result
}
