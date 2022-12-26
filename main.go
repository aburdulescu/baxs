package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"bandr.me/p/baxs/internal/ipc"
	"bandr.me/p/baxs/internal/waiter"
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
    ls        list services

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
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			return fmt.Errorf("failed to read build info")
		}
		fmt.Println(
			bi.Main.Version,
			bi.GoVersion,
			findBuildSetting(bi.Settings, "GOOS"),
			findBuildSetting(bi.Settings, "GOARCH"),
			findBuildSetting(bi.Settings, "vcs.revision"),
			findBuildSetting(bi.Settings, "vcs.time"),
		)
		return nil
	case "daemon":
		return runDaemon(args)
	case "ls":
		return runLs(args)
	default:
		return fmt.Errorf("unknown command '%s'", cmd)
	}
}

func findBuildSetting(settings []debug.BuildSetting, key string) string {
	for _, v := range settings {
		if v.Key == key {
			return v.Value
		}
	}
	return ""
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

	waiter, err := waiter.New(*configPath)
	if err != nil {
		return err
	}

	ipcDaemon, err := ipc.NewDaemon()
	if err != nil {
		return err
	}

	go ipcDaemon.Start()

	if err := waiter.Start(); err != nil {
		return err
	}

	if err := waiter.Wait(); err != nil {
		return err
	}

	return nil
}

func runLs(args []string) error {
	services, err := ipc.Ls()
	if err != nil {
		return err
	}
	fmt.Println(services)
	return nil
}
