package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func run() error {
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
	var services []ServiceConf
	for _, s := range config.GetSections(sectionFilter) {
		var svc ServiceConf
		if err := s.To(&svc); err != nil {
			return err
		}
		services = append(services, svc)
	}

	os.Mkdir(daemonConf.LogsDir, 0755)

	for _, svc := range services {
		fmt.Println("run:", svc.Command)
		args := strings.Split(svc.Command, " ")
		outfile, err := os.Create(filepath.Join(daemonConf.LogsDir, args[0]+".out"))
		if err != nil {
			return err
		}
		errfile, err := os.Create(filepath.Join(daemonConf.LogsDir, args[0]+".err"))
		if err != nil {
			return err
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = outfile
		cmd.Stderr = errfile
		if err := cmd.Start(); err != nil {
			return err
		}
		log.Printf("%s started with pid %v\n", args[0], cmd.Process.Pid)
	}

	for range services {
		var ws unix.WaitStatus
		var ru unix.Rusage // TODO: use it
		wpid, err := unix.Wait4(-1, &ws, 0, &ru)
		if err != nil {
			return err
		}
		switch {
		case ws.Exited():
			log.Printf("pid %v exited with exit code %d\n", wpid, ws.ExitStatus())
		case ws.Signaled():
			log.Printf("pid %v terminated by signal %d\n", wpid, ws.Signal())
		case ws.Stopped():
			log.Printf("pid %v stopped by signal %d\n", wpid, ws.StopSignal())
		default:
			log.Printf("pid %v status %d\n", wpid, ws)
		}
	}

	return nil
}
