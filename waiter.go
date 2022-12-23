package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type Config struct {
	Daemon   DaemonConfig
	Services []ServiceConfig
}

type DaemonConfig struct {
	LogsDir string
}

type ServiceConfig struct {
	Name    string
	Command string
}

type Service struct {
	cmd    *exec.Cmd
	config ServiceConfig
}

type Waiter struct {
	daemonConfig DaemonConfig
	services     []Service
	pidToSvc     map[int]*Service
}

func newWaiter(configPath string) (*Waiter, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, err
	}

	var w Waiter

	w.daemonConfig = config.Daemon

	for _, s := range config.Services {
		w.services = append(w.services, Service{
			config: s,
		})
	}

	if err := os.MkdirAll(w.daemonConfig.LogsDir, 0755); err != nil {
		return nil, err
	}

	w.pidToSvc = make(map[int]*Service)

	return &w, nil
}

func (w *Waiter) start() error {
	var err error
	defer func() {
		if err == nil {
			return
		}
		for _, svc := range w.services {
			if svc.cmd == nil {
				continue
			}
			if err := svc.cmd.Process.Kill(); err != nil {
				log.Printf("[%s] failed to be kill: %v\n", svc.config.Name, err)
			}
			log.Printf("[%s] kill signal sent\n", svc.config.Name)
			if err := svc.cmd.Wait(); err != nil {
				log.Printf("[%s] failed to be wait: %v\n", svc.config.Name, err)
			}
		}
	}()

	for i, svc := range w.services {
		log.Printf("[%s] starting with command=%s\n", svc.config.Name, svc.config.Command)

		var outfile *os.File
		outfile, err = os.Create(filepath.Join(w.daemonConfig.LogsDir, svc.config.Name+".out"))
		if err != nil {
			return err
		}

		var errfile *os.File
		errfile, err = os.Create(filepath.Join(w.daemonConfig.LogsDir, svc.config.Name+".err"))
		if err != nil {
			return err
		}

		args := strings.Split(svc.config.Command, " ")

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = outfile
		cmd.Stderr = errfile

		err = cmd.Start()
		if err != nil {
			return err
		}

		log.Printf("[%s] started with pid %v\n", svc.config.Name, cmd.Process.Pid)

		w.services[i].cmd = cmd
		w.pidToSvc[cmd.Process.Pid] = &w.services[i]
	}

	return nil
}

func (w *Waiter) wait() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Println("received signal", sig)
		for _, svc := range w.services {
			if err := svc.cmd.Process.Kill(); err != nil {
				log.Printf("[%s] failed to kill: %v\n", svc.config.Name, err)
			}
			log.Printf("[%s] kill signal sent\n", svc.config.Name)
			if err := svc.cmd.Wait(); err != nil {
				log.Printf("[%s] failed to wait: %v\n", svc.config.Name, err)
			}
		}
		os.Exit(127 + int(sig.(syscall.Signal)))
	}()

	for range w.services {
		var ws unix.WaitStatus
		var ru unix.Rusage // TODO: use it
		wpid, err := unix.Wait4(-1, &ws, 0, &ru)
		if err != nil {
			return err
		}
		svc, found := w.pidToSvc[wpid]
		if !found {
			continue
		}
		switch {
		case ws.Exited():
			log.Printf("[%s] exited with exit code %d\n", svc.config.Name, ws.ExitStatus())
		case ws.Signaled():
			log.Printf("[%s] terminated by signal %d\n", svc.config.Name, ws.Signal())
		case ws.Stopped():
			log.Printf("[%s] stopped by signal %d\n", svc.config.Name, ws.StopSignal())
		default:
			log.Printf("[%s] status %d\n", svc.config.Name, ws)
		}
	}

	return nil
}
