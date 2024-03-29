package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"text/tabwriter"

	"bandr.me/p/baxs/internal/daemon"
	"bandr.me/p/baxs/internal/ipc"
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
  ps      List available services
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
	case "ps":
		return runPs(args)
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
  -h, --help  Print this message
  -l          Base directory for logs
  -f          Path to baxsfile

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

	daemon, err := daemon.New(*logsDir, *baxsfilePath)
	if err != nil {
		return err
	}

	return daemon.Run()
}

func runPs(args []string) error {
	fset := flag.NewFlagSet("ls", flag.ExitOnError)

	fset.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: baxs ps [options]

Print available services and their status.

Options:
  -h, --help  Print this message

`)
	}

	if err := fset.Parse(args); err != nil {
		return err
	}

	services, err := ipc.Ps()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "Name\tPID\tStatus\n")
	fmt.Fprintf(w, "----\t---\t-------\n")

	for _, s := range services {
		fmt.Fprintf(w, "%s\t%d\t%s\n", s.Name, s.Pid, s.Status)
	}

	return nil
}

func runStop(args []string) error {
	fset := flag.NewFlagSet("stop", flag.ExitOnError)

	fset.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: baxs stop [options] [serviceName...]

Stop given service(s).
If no service name is given, stop all services.

Options:
  -h, --help  Print this message
  -k          Force stop the service(s); i.e. send SIGKILL
`)
	}

	force := fset.Bool("k", false, "")

	if err := fset.Parse(args); err != nil {
		return err
	}

	return ipc.Stop(*force, fset.Args()...)
}

func runStart(args []string) error {
	fset := flag.NewFlagSet("start", flag.ExitOnError)

	fset.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: baxs start [options] [serviceName...]

Start given service(s).
If no service name is given, start all services.

Options:
  -h, --help  Print this message

`)
	}

	if err := fset.Parse(args); err != nil {
		return err
	}

	return ipc.Start(fset.Args()...)
}
