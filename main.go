package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	versionCommit = "none"
	versionDate   = "none"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

const usage = `usage: baxs command [flags]

Commands:
    help      print this message
    version   print version information
    daemon    start daemon
    ctl       control services

Globals flags:
    -h, --help    print this message
    --version     print version information
`

func mainErr() error {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("missing command")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	case "version", "--version":
		fmt.Println("Commit:", versionCommit)
		fmt.Println("Date:", versionDate)
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

	ipcServer, err := newIPCServer()
	if err != nil {
		return err
	}

	go ipcServer.start()

	if err := waiter.start(); err != nil {
		return err
	}

	if err := waiter.wait(); err != nil {
		return err
	}

	return nil
}

func runCtl(args []string) error {
	conn, err := net.Dial("unix", daemonSocketFile)
	if err != nil {
		return err
	}
	defer conn.Close()
	req := IPCRequest{
		Cmd: "ls",
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(req); err != nil {
		return err
	}
	if _, err := io.Copy(conn, buf); err != nil {
		return err
	}
	buf.Reset()
	if _, err := io.Copy(buf, conn); err != nil {
		return err
	}
	var rsp IPCResponse
	if err := json.NewDecoder(buf).Decode(&rsp); err != nil {
		return err
	}
	fmt.Println(rsp)
	return err
}
