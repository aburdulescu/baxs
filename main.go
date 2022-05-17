package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bandr.me/baxs/conf"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

type Service struct {
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
	sectionFilter := func(name string) bool {
		return strings.HasPrefix(name, "service:")
	}
	var services []Service
	for _, s := range config.GetSections(sectionFilter) {
		var svc Service
		if err := s.To(&svc); err != nil {
			return err
		}
		services = append(services, svc)
	}
	for _, svc := range services {
		fmt.Println("run:", svc.Command)
		args := strings.Split(svc.Command, " ")
		b, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	}
	return nil
}
