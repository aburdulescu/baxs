package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"bandr.me/baxs/conf"
	"golang.org/x/sys/unix"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type DaemonConf struct {
	LogsDir string `conf:"logs_dir"`
}

type ServiceConf struct {
	Command string `conf:"command"`
}

type Service struct {
	cmd  *exec.Cmd
	name string
	conf ServiceConf
}

func run() error {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)

	configPath := flag.String("c", "./baxs.conf", "path to config file")
	flag.Parse()

	waiter, err := newWaiter(*configPath)
	if err != nil {
		return err
	}

	if err := waiter.wait(); err != nil {
		return err
	}

	return nil
}

type Waiter struct {
	daemonConf DaemonConf
	services   []Service
	pidToSvc   map[int]*Service
}

func newWaiter(configPath string) (*Waiter, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var config conf.Conf
	if err := config.Parse(f); err != nil {
		return nil, err
	}

	var w Waiter

	var daemonConf DaemonConf
	if s := config.GetSection("daemon"); s == nil {
		return nil, fmt.Errorf("'daemon' section not found in config file")
	} else {
		if err := s.To(&w.daemonConf); err != nil {
			return nil, err
		}
	}

	sectionFilter := func(name string) bool {
		return strings.HasPrefix(name, "service:")
	}
	for _, s := range config.GetSections(sectionFilter) {
		var svc ServiceConf
		if err := s.To(&svc); err != nil {
			return nil, err
		}
		name := strings.Split(s.Name(), ":")[1]
		w.services = append(w.services, Service{
			name: name,
			conf: svc,
		})
	}

	os.Mkdir(daemonConf.LogsDir, 0755)

	w.pidToSvc = make(map[int]*Service)

	if err := w.startServices(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (w *Waiter) startServices() error {
	var err error
	defer func() {
		if err == nil {
			return
		}
		for _, svc := range w.services {
			if err := svc.cmd.Process.Kill(); err != nil {
				log.Printf("[%s] failed to be kill: %v\n", svc.name, err)
			}
			log.Printf("[%s] kill signal sent\n", svc.name)
			if err := svc.cmd.Wait(); err != nil {
				log.Printf("[%s] failed to be wait: %v\n", svc.name, err)
			}
		}
	}()

	for i, svc := range w.services {
		log.Printf("[%s] starting with command=%s\n", svc.name, svc.conf.Command)

		var outfile *os.File
		outfile, err = os.Create(filepath.Join(w.daemonConf.LogsDir, svc.name+".out"))
		if err != nil {
			return err
		}

		var errfile *os.File
		errfile, err = os.Create(filepath.Join(w.daemonConf.LogsDir, svc.name+".err"))
		if err != nil {
			return err
		}

		args := strings.Split(svc.conf.Command, " ")

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = outfile
		cmd.Stderr = errfile

		err = cmd.Start()
		if err != nil {
			return err
		}

		log.Printf("[%s] started with pid %v\n", svc.name, cmd.Process.Pid)

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
				log.Printf("[%s] failed to be kill: %v\n", svc.name, err)
			}
			log.Printf("[%s] kill signal sent\n", svc.name)
			if err := svc.cmd.Wait(); err != nil {
				log.Printf("[%s] failed to be wait: %v\n", svc.name, err)
			}
		}
		os.Exit(1)
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
			log.Printf("[%s] exited with exit code %d\n", svc.name, ws.ExitStatus())
		case ws.Signaled():
			log.Printf("[%s] terminated by signal %d\n", svc.name, ws.Signal())
		case ws.Stopped():
			log.Printf("[%s] stopped by signal %d\n", svc.name, ws.StopSignal())
		default:
			log.Printf("[%s] status %d\n", svc.name, ws)
		}
	}

	return nil
}

// type Server struct{}

// func newServer() *Server {
// 	return &Server{}
// }

// func (s *Server) run() {}
