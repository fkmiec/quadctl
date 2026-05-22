package util

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Consts and Config
const (
	ToolName = "quadctl"
)

var flagSets map[string]*flag.FlagSet

func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Orchestrator for Podman Quadlets (without systemd)\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [directory]\n\n", ToolName)
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  pull     : Pull required images\n")
	fmt.Fprintf(os.Stderr, "  create   : Create resources (force re-creation), do not start. Use -s flag to generate quadlets.\n")
	fmt.Fprintf(os.Stderr, "  start    : Create (if missing) and start services. Use -s flag to start containers with systemd.\n")
	fmt.Fprintf(os.Stderr, "  run      : Run a container (single .container file or specified with -f flag). Use --pargs to add or override podman args in the quadlet file, such as -it for interactive and --rm for ephemeral. Not applicable to systemd.\n")
	fmt.Fprintf(os.Stderr, "  stop     : Stop running services (do not remove). Use -s flag to stop containers run by systemd.\n")
	fmt.Fprintf(os.Stderr, "  remove   : Remove stopped resources. Use -s flag to remove generated quadlets.\n")
	fmt.Fprintf(os.Stderr, "  status   : Show current status. Use -s flag to see systemd status.\n")
	fmt.Fprintf(os.Stderr, "  logs     : Show logs of running containers. Use -s flag to view systemd logs.\n")
	fmt.Fprintf(os.Stderr, "  list, ls : List quadlets in the configured quadlet_path or systemd path if -s flag is used.\n")
	fmt.Fprintf(os.Stderr, "\nWrapper commands (filtered to defined resources):\n")
	fmt.Fprintf(os.Stderr, "  ps, stats, images\n")

}

func InitFlags(quadctl *Quadctl) {

	flagSets = map[string]*flag.FlagSet{}

	// Handle flags
	flag.BoolVar(&quadctl.IsSystemd, "systemd", false, "Use systemd for managing services (default: false)")
	flag.BoolVar(&quadctl.IsSystemd, "s", false, "Use systemd for managing services (default: false)")
	flag.Usage = PrintUsage
	flagSets["global"] = flag.CommandLine

	// pull
	pullFlags := flag.NewFlagSet("pull", flag.ExitOnError)
	pullFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	pullFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	pullFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	pullFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	flagSets["pull"] = pullFlags

	// create
	createFlags := flag.NewFlagSet("create", flag.ExitOnError)
	createFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	createFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	createFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	createFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	createFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	createFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	flagSets["create"] = createFlags

	// start
	startFlags := flag.NewFlagSet("start", flag.ExitOnError)
	startFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	startFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	startFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	startFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	startFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	startFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	flagSets["start"] = startFlags

	// run
	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	runFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	runFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	runFlags.StringVar(&quadctl.PodmanArgs, "pargs", "", "Additional arguments to pass to podman when using the 'run' command (e.g., --pargs='--rm -it')")
	runFlags.StringVar(&quadctl.RunCmd, "exec", "", "Command to execute in the container when using the 'run' command (e.g., --exec='/bin/bash')")
	runFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	runFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	runFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	runFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	flagSets["run"] = runFlags

	// stop
	stopFlags := flag.NewFlagSet("stop", flag.ExitOnError)
	stopFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	stopFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	stopFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	stopFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	stopFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	stopFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	flagSets["stop"] = stopFlags

	// remove
	removeFlags := flag.NewFlagSet("remove", flag.ExitOnError)
	removeFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	removeFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	removeFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	removeFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	removeFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	removeFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	flagSets["remove"] = removeFlags
	flagSets["rm"] = removeFlags

	// status
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)
	statusFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statusFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	flagSets["status"] = statusFlags

	// ps
	psFlags := flag.NewFlagSet("ps", flag.ExitOnError)
	psFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	psFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	flagSets["ps"] = psFlags

	// stats
	statsFlags := flag.NewFlagSet("stats", flag.ExitOnError)
	statsFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statsFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	flagSets["stats"] = statsFlags

	// images
	imagesFlags := flag.NewFlagSet("images", flag.ExitOnError)
	imagesFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	imagesFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	flagSets["images"] = imagesFlags

	// list, ls
	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	flagSets["list"] = listFlags
	flagSets["ls"] = listFlags

	// logs
	logsFlags := flag.NewFlagSet("logs", flag.ExitOnError)
	flagSets["logs"] = logsFlags

	flag.Parse()

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}
}

func ProcessSubcommand(quadctl *Quadctl) {
	quadctl.Subcommand = strings.ToLower(flag.Arg(0))
	flagSets[quadctl.Subcommand].Parse(flag.Args()[1:])
	quadctl.SearchDir = getSearchDir(quadctl, flagSets[quadctl.Subcommand].Arg(0))
}

func getSearchDir(quadctl *Quadctl, path string) string {

	// Determine search directory (optional path or CWD ... optional path may be relative to CWD or quadlets_path from config)
	// If no path is specified, use the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting CWD: %v\n", err)
		os.Exit(1)
	}
	// If a path is specified, determine if relative to CWD or quadlets_path
	//if flag.NArg() > 1 {
	if path != "" {
		// If os.Stat returns no error, the path is absolute or valid relative to the current working directory
		if info, err := os.Stat(path); err == nil {
			//if a file was specified, get parent directory of the file
			if !info.IsDir() {
				dir = filepath.Dir(path)
			} else {
				dir, _ = filepath.Abs(path)
			}
			// Otherwise, look for specified directory path relative to the quadlets path
		} else {
			dir = filepath.Join(quadctl.QuadletSrcPath, path)
			// If the path is not found relative to the quadlets path or is not a directory, it's an error
			if info, err := os.Stat(dir); err == nil {
				//if a file was specified, get parent directory of the file
				if !info.IsDir() {
					dir = filepath.Dir(path)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s not found\n", path)
				os.Exit(1)
			}
		}
	}

	return dir
}
