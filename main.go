package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"

	"bandr.me/p/baxs/ipc"
)

func main() {
	if err := mainErr(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: baxs <command> [options]

Options:
  -h, --help  print this message
  --version   print version information

Commands:
  daemon  Start the daemon
  ls      List available services
  stop    Stop service(s)
  start   Start service(s)

Run 'baxs <command> -h' for more information about a command.

`)
	}

	printVersion := flag.Bool("version", false, "")

	flag.Parse()

	if *printVersion {
		bi, _ := debug.ReadBuildInfo()
		g := func(key string) string {
			for _, v := range bi.Settings {
				if v.Key == key {
					return v.Value
				}
			}
			return ""
		}
		fmt.Println(bi.Main.Version, bi.GoVersion, g("GOOS"), g("GOARCH"), g("vcs.revision"), g("vcs.time"))
		return nil
	}

	args := flag.Args()

	if len(args) < 1 {
		flag.Usage()
		return fmt.Errorf("command was not specified")
	}

	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "daemon":
		return runDaemon(args)
	case "ls":
		return runLs(args)
	case "stop":
		return runStop(args)
	case "start":
		return runStart(args)
	default:
		return fmt.Errorf("unknown command '%s'", cmd)
	}
}

func runDaemon(args []string) error {
	fset := flag.NewFlagSet("daemon", flag.ExitOnError)

	fset.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: baxs daemon [options]

Options:
  -l  Base directory for logs
  -f  Path to baxsfile

`)
	}

	logsDir := fset.String("l", "", "")
	baxsfilePath := fset.String("f", "", "")

	if err := fset.Parse(args); err != nil {
		return err
	}

	if *logsDir == "" {
		fset.Usage()
		return fmt.Errorf("logs dir(-l) must be specified")
	}

	if *baxsfilePath == "" {
		fset.Usage()
		return fmt.Errorf("path to baxfile(-f) must be specified")
	}

	daemon, err := NewDaemon(*logsDir, *baxsfilePath)
	if err != nil {
		return err
	}

	return daemon.Run()
}

func runLs(args []string) error {
	services, err := ipc.Ls()
	if err != nil {
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer w.Flush()
	fmt.Fprintf(w, "Name\tStatus\n")
	fmt.Fprintf(w, "----\t-------\n")
	for _, s := range services {
		fmt.Fprintf(w, "%s\t%s\n", s.Name, s.Status)
	}
	return nil
}

func runStop(args []string) error {
	return ipc.Stop(args...)
}

func runStart(args []string) error {
	return ipc.Start(args...)
}
