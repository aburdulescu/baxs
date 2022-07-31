package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

const usage = `usage: baxs command [flags]

Commands:
    help      print this message
    daemon    start daemon
    ctl       control services

Globals flags:
    -h, --help    print this message
`

func run() error {
	args := os.Args[1:]

	if len(args) < 1 {
		return fmt.Errorf("missing command")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	case "daemon":
		return runDaemon(args)
	case "ctl":
		return runCtl(args)
	default:
		return fmt.Errorf("unknown command '%s'", cmd)
	}
}

func runDaemon(args []string) error {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Lmicroseconds | log.LUTC)

	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `usage: baxs daemon [-h/--help] [flags]

Flags:`)
		fs.PrintDefaults()
		os.Exit(1)
	}
	configPath := fs.String("c", "", "path to config file")
	fs.Parse(args)

	if *configPath == "" {
		return fmt.Errorf("path to config file not specified")
	}

	waiter, err := newWaiter(*configPath)
	if err != nil {
		return err
	}

	if err := waiter.start(); err != nil {
		return err
	}

	if err := waiter.wait(); err != nil {
		return err
	}

	return nil
}

func runCtl(args []string) error {
	return fmt.Errorf("not implemented")
}
