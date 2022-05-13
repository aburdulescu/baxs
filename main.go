package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"bandr.me/baxs/conf"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
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
	return nil
}
