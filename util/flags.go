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
	pullFlags.Usage = PrintPullUsage
	flagSets["pull"] = pullFlags

	// create
	createFlags := flag.NewFlagSet("create", flag.ExitOnError)
	createFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	createFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	createFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	createFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	createFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	createFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	createFlags.Usage = PrintCreateUsage
	flagSets["create"] = createFlags

	// start
	startFlags := flag.NewFlagSet("start", flag.ExitOnError)
	startFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	startFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	startFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	startFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	startFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	startFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	startFlags.Usage = PrintStartUsage
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
	runFlags.Usage = PrintRunUsage
	flagSets["run"] = runFlags

	// stop
	stopFlags := flag.NewFlagSet("stop", flag.ExitOnError)
	stopFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	stopFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	stopFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	stopFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	stopFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	stopFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	stopFlags.Usage = PrintStopUsage
	flagSets["stop"] = stopFlags

	// remove
	removeFlags := flag.NewFlagSet("remove", flag.ExitOnError)
	removeFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	removeFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	removeFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	removeFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	removeFlags.BoolVar(&quadctl.IsVerbose, "verbose", false, "Print detailed information about command execution and warnings")
	removeFlags.BoolVar(&quadctl.IsVerbose, "v", false, "Print detailed information about command execution and warnings")
	removeFlags.Usage = PrintRemoveUsage
	flagSets["remove"] = removeFlags
	flagSets["rm"] = removeFlags

	// status
	statusFlags := flag.NewFlagSet("status", flag.ExitOnError)
	statusFlags.BoolVar(&quadctl.IsLongStatus, "long", false, "Display long format output from systemctl status (default: false)")
	statusFlags.BoolVar(&quadctl.IsLongStatus, "l", false, "Display long format output from systemctl status (default: false)")
	statusFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statusFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statusFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	statusFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	statusFlags.Usage = PrintStatusUsage
	flagSets["status"] = statusFlags

	// ps
	psFlags := flag.NewFlagSet("ps", flag.ExitOnError)
	psFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	psFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	psFlags.Usage = PrintPsUsage
	flagSets["ps"] = psFlags

	// stats
	statsFlags := flag.NewFlagSet("stats", flag.ExitOnError)
	statsFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statsFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	statsFlags.Usage = PrintStatsUsage
	flagSets["stats"] = statsFlags

	// images
	imagesFlags := flag.NewFlagSet("images", flag.ExitOnError)
	imagesFlags.BoolVar(&quadctl.IsFile, "file", false, "Specify that the provided path is a file rather than a directory (default: false)")
	imagesFlags.BoolVar(&quadctl.IsFile, "f", false, "Specify that the provided path is a file rather than a directory (default: false)")
	imagesFlags.Usage = PrintImagesUsage
	flagSets["images"] = imagesFlags

	// list, ls
	listFlags := flag.NewFlagSet("list", flag.ExitOnError)
	listFlags.IntVar(&quadctl.ListDepth, "depth", 2, "Specify the depth of the quadlet directory listing.")
	listFlags.IntVar(&quadctl.ListDepth, "d", 2, "Specify the depth of the quadlet directory listing.")
	listFlags.BoolVar(&quadctl.IsListAll, "all", false, "List quadlets in all configured paths (src, systemd user and systemd root).")
	listFlags.BoolVar(&quadctl.IsListAll, "a", false, "List quadlets in all configured paths (src, systemd user and systemd root).")
	listFlags.Usage = PrintListUsage
	flagSets["list"] = listFlags
	flagSets["ls"] = listFlags

	// logs
	logsFlags := flag.NewFlagSet("logs", flag.ExitOnError)
	logsFlags.BoolVar(&quadctl.IsPrintOnly, "print", false, "Print podman commands without executing")
	logsFlags.BoolVar(&quadctl.IsPrintOnly, "p", false, "Print podman commands without executing")
	logsFlags.Usage = PrintLogsUsage
	flagSets["logs"] = logsFlags

	flag.Parse()

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}
}

func PrintUsage() {
	fmt.Fprintf(os.Stderr, "Orchestrator for Podman Quadlets (with and without systemd)\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command> [path]\n\n", ToolName)
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  pull       : Pull required images\n")
	fmt.Fprintf(os.Stderr, "  create     : Create resources (do not start). Use -s flag to generate quadlets under systemd.\n")
	fmt.Fprintf(os.Stderr, "  start      : Create (if missing) and start resources. Use -s flag to start under systemd.\n")
	fmt.Fprintf(os.Stderr, "  run        : Run a single .container in the foreground. Not supported for systemd. See quadctl run --help.\n")
	fmt.Fprintf(os.Stderr, "  stop       : Stop running services (do not remove). Use -s flag to stop under systemd.\n")
	fmt.Fprintf(os.Stderr, "  remove, rm : Remove stopped resources. Use -s flag to remove generated quadlets under systemd.\n")
	fmt.Fprintf(os.Stderr, "  logs       : Show logs of running containers. Use -s flag to view systemd logs.\n")
	fmt.Fprintf(os.Stderr, "  list, ls   : List quadlets in the configured quadlet_path or systemd path if -s flag is used.\n")
	fmt.Fprintf(os.Stderr, "\nWrapper commands (filtered to defined resources):\n")
	fmt.Fprintf(os.Stderr, "  images : Show images defined for the set of related quadlets.\n")
	fmt.Fprintf(os.Stderr, "  ps     : Show state of containers.\n")
	fmt.Fprintf(os.Stderr, "  stats  : Show live stats for containers.\n")
	fmt.Fprintf(os.Stderr, "  status : Show current status. Use -s flag to see systemd status.\n")
	fmt.Fprintf(os.Stderr, "\nRequirements:\n")
	fmt.Fprintf(os.Stderr, "  A quadctl.ini config file is required. Default location is $HOME/.config/quadctl.\n    A default config file will be created if not found.\n")
	fmt.Fprintf(os.Stderr, "  Set QUADCTL_CONFIG_DIR=<absolute path to config directory> in /etc/environment to\n    change config location and/or ensure found when using sudo.\n")
}

func PrintPullUsage() {
	fmt.Fprintf(os.Stderr, "Pull images defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s %s [flags] [path]\n\n", ToolName, "pull")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["pull"].PrintDefaults()
}

func PrintCreateUsage() {
	fmt.Fprintf(os.Stderr, "Create resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "create")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["create"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to create under systemd.\n")
}

func PrintStartUsage() {
	fmt.Fprintf(os.Stderr, "Start resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "start")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["start"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to start under systemd.\n")
}

func PrintRunUsage() {
	fmt.Fprintf(os.Stderr, "Run resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "run")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["run"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Will run a single .container quadlet in the foreground. Other quadlets will be run in background.\n")
	fmt.Fprintf(os.Stderr, "  For example, pass '-it' in PodmanArgs for interactive terminal use.\n")
	fmt.Fprintf(os.Stderr, "  Use --pargs flag to pass podman args on the quadctl command line. Equivalent to PodmanArgs quadlet option.\n")
	fmt.Fprintf(os.Stderr, "  The run command (ie. Running in foreground) is incompatible with systemd.\n")
	fmt.Fprintf(os.Stderr, "  If multiple .container files are found, all but one must have -d (--detach) defined in PodmanArgs.\n")
	fmt.Fprintf(os.Stderr, "  Using run where all .container quadlets have -d (--detach) in PodmanArgs is same as 'quadctl start'.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to start under systemd.\n")
}

func PrintStopUsage() {
	fmt.Fprintf(os.Stderr, "Stop resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "stop")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["stop"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to stop under systemd.\n")
}

func PrintRemoveUsage() {
	fmt.Fprintf(os.Stderr, "Remove resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "remove|rm")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["remove"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Will stop running resources if needed and remove networks and volumes.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to remove under systemd.\n")
}

func PrintStatusUsage() {
	fmt.Fprintf(os.Stderr, "Display status for resources (pod, container, volume, network) defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "status")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["status"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Will display systemd status if -s flag provided. Otherwise, calls 'quadctl ps'.\n")
	fmt.Fprintf(os.Stderr, "  Calls 'podman quadlet list' for systemd status by default, 'systemctl status' with -l flag.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
	fmt.Fprintf(os.Stderr, "  Use -s to display status of quadlets under systemd.\n")
}

func PrintPsUsage() {
	fmt.Fprintf(os.Stderr, "Display state of containers defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "ps")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["ps"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Will display state of both running and stopped containers.\n")
	fmt.Fprintf(os.Stderr, "  Displays the same information whether containers are running under systemd or podman.\n")
	fmt.Fprintf(os.Stderr, "  The -s flag is not required or supported.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
}

func PrintStatsUsage() {
	fmt.Fprintf(os.Stderr, "Display live stats of running containers defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "stats")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["stats"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Displays the same information whether containers are running under systemd or podman.\n")
	fmt.Fprintf(os.Stderr, "  The -s flag is not required or supported.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
}

func PrintImagesUsage() {
	fmt.Fprintf(os.Stderr, "List images defined in quadlet files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "images")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["images"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Displays the same information whether containers are running under systemd or podman.\n")
	fmt.Fprintf(os.Stderr, "  The -s flag is not required or supported.\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
}

func PrintListUsage() {
	fmt.Fprintf(os.Stderr, "Display a tree view of quadlet directories and files.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "list|ls")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["list"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  'quadctl list' will display quadlets under your configured quadlet.src.path.\n")
	fmt.Fprintf(os.Stderr, "  'quadctl -s list' will display quadlets under your configured quadlet.user.path.\n")
	fmt.Fprintf(os.Stderr, "  'sudo quadctl -s list' will display quadlets under your configured quadlet.root.path.\n")
	fmt.Fprintf(os.Stderr, "  'quadctl list -a' will display quadlets in all three configured paths.\n")
	fmt.Fprintf(os.Stderr, "  At default depth only quadlet directories are listed. Add -d [3+] to list files.\n")
}

func PrintLogsUsage() {
	fmt.Fprintf(os.Stderr, "Display logs.\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] %s [flags] [path]\n\n", ToolName, "logs")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flagSets["stats"].PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nNotes:\n")
	fmt.Fprintf(os.Stderr, "  Helper command to open relevant logs (particularly under systemd with -s flag).\n")
	fmt.Fprintf(os.Stderr, "  Use sudo for rootless quadlets.\n")
}

func ProcessSubcommand(quadctl *Quadctl) {
	quadctl.Subcommand = strings.ToLower(flag.Arg(0))
	if flagSet, ok := flagSets[quadctl.Subcommand]; ok {
		flagSet.Parse(flag.Args()[1:])
		quadctl.SearchDir = getSearchDir(quadctl, flagSets[quadctl.Subcommand].Arg(0))
	} else {
		fmt.Fprintf(os.Stderr, "Invalid command: %s\n", quadctl.Subcommand)
		os.Exit(1)
	}
}

func getSearchDir(quadctl *Quadctl, path string) string {

	// Determine search directory (optional path or CWD ... optional path may be relative to CWD or quadlets_path from config)
	// If no path is specified, use the current working directory
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting CWD: %v\n", err)
		os.Exit(1)
	}
	// If a path is specified, determine if relative to CWD or quadlet.src.path
	if path != "" {
		// If os.Stat returns no error, the path is absolute or valid relative to the current working directory
		info, err := os.Stat(path)
		if err == nil {
			//if a file was specified, get parent directory of the file
			if !info.IsDir() {
				dir = filepath.Dir(path)
			} else {
				dir, _ = filepath.Abs(path)
			}
		} else {
			// Otherwise, look for specified directory path relative to the quadlets path
			dir = filepath.Join(quadctl.QuadletSrcPath, path)
			// If the path is not found relative to the quadlets path or is not a directory, it's an error
			info, err = os.Stat(dir)
			if err == nil {
				//if a file was specified, get parent directory of the file
				if !info.IsDir() {
					dir = filepath.Dir(dir)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error: %s not found\n", path)
				os.Exit(1)
			}
		}
	}

	return dir
}
