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

	f, err := os.Open(*configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var config conf.Conf
	if err := config.Parse(f); err != nil {
		return err
	}

	var daemonConf DaemonConf
	if s := config.GetSection("daemon"); s == nil {
		return fmt.Errorf("'daemon' section not found in config file")
	} else {
		if err := s.To(&daemonConf); err != nil {
			return err
		}
	}

	sectionFilter := func(name string) bool {
		return strings.HasPrefix(name, "service:")
	}
	var services []Service
	for _, s := range config.GetSections(sectionFilter) {
		var svc ServiceConf
		if err := s.To(&svc); err != nil {
			return err
		}
		name := strings.Split(s.Name(), ":")[1]
		services = append(services, Service{
			name: name,
			conf: svc,
		})
	}

	os.Mkdir(daemonConf.LogsDir, 0755)

	pidToSvc := make(map[int]*Service)

	for i, svc := range services {
		log.Printf("[%s] starting with command=%s\n", svc.name, svc.conf.Command)
		outfile, err := os.Create(filepath.Join(daemonConf.LogsDir, svc.name+".out"))
		if err != nil {
			return err
		}
		errfile, err := os.Create(filepath.Join(daemonConf.LogsDir, svc.name+".err"))
		if err != nil {
			return err
		}
		args := strings.Split(svc.conf.Command, " ")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = outfile
		cmd.Stderr = errfile
		if err := cmd.Start(); err != nil {
			return err
		}
		log.Printf("[%s] started with pid %v\n", svc.name, cmd.Process.Pid)
		services[i].cmd = cmd
		pidToSvc[cmd.Process.Pid] = &services[i]
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-c
		log.Println("received signal", sig)
		for _, svc := range services {
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

	for range services {
		var ws unix.WaitStatus
		var ru unix.Rusage // TODO: use it
		wpid, err := unix.Wait4(-1, &ws, 0, &ru)
		if err != nil {
			return err
		}
		svc, found := pidToSvc[wpid]
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

// type Waiter struct{}

// func newWaiter() (*Waiter, error) {
// 	return &Waiter{}
// }

// func (w *Waiter) wait() error {}

// type Server struct{}

// func newServer() *Server {
// 	return &Server{}
// }

// func (s *Server) run() {}
