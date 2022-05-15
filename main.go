package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	log.Println(config)
	serviceSections := config.GetSections(func(name string) bool {
		return strings.HasPrefix(name, "service:")
	})
	for _, s := range serviceSections {
		var svc Service
		if err := s.To(&svc); err != nil {
			return err
		}
		log.Println(svc)
	}
	return nil
}
