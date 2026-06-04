package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"

	. "github.com/fkmiec/quadctl/core"
	. "github.com/fkmiec/quadctl/schema"
	"github.com/fkmiec/quadctl/util"
)

var (
	quadctl *util.Quadctl
)

func main() {

	initState() //Create the initial quadctl state object

	util.InitFlags(quadctl)
	util.InitConfig(quadctl)
	util.ProcessSubcommand(quadctl)
	quadctl.QuadletSchemas = util.GetQuadletSchemas()

	quadlets := util.InitQuadlets(quadctl)

	// If no quadlets at this point, only list|ls is still a valid command.
	// Abort with a message. User probably didn't notice they were neither in a quadlet directory nor specified one as argument.
	if len(quadlets) < 1 && !(slices.Contains([]string{"list", "ls", "logs"}, quadctl.Subcommand)) {

		err := displayQuadletSelector(quadctl)
		if err != nil {
			fmt.Printf("Error: No quadlets found in directory: %s\n", quadctl.SearchDir)
			os.Exit(1)
		}
		quadlets = util.InitQuadlets(quadctl)
	}

	var commands []Command

	// Route to appropriate subcommand handler
	switch quadctl.Subcommand {
	case "ps":
		HandlePS(quadctl, quadlets)
	case "stats":
		HandleStats(quadctl, quadlets)
	case "status":
		if quadctl.IsSystemd {
			commands = HandleSystemdStatus(quadctl, quadlets)
		} else {
			HandlePS(quadctl, quadlets)
		}
	case "logs":
		if quadctl.IsSystemd {
			commands = HandleSystemdLogs(quadctl, quadlets)
		} else {
			fmt.Println("To view podman logs, use 'podman logs <container name or id>'")
			os.Exit(0)
		}
	case "images":
		HandleImages(quadlets)
	case "pull":
		commands = HandlePull(quadctl, quadlets)
	case "list", "ls":
		HandleList(quadctl)
	case "create":
		if quadctl.IsSystemd {
			commands = HandleSystemdCreate(quadctl, quadlets)
		} else {
			commands = HandleCreate(quadctl, quadlets)
		}
	case "start":
		if quadctl.IsSystemd {
			commands = HandleSystemdStart(quadctl, quadlets)
		} else {
			commands = HandleStart(quadctl, quadlets)
		}
	case "run":
		if quadctl.IsSystemd {
			fmt.Printf("Running containers with systemd (ie. 'quadctl -s run') is not supported since systemd manages the lifecycle of services independently. Use 'start' to start the services and ensure your quadlets are configured to run the desired commands on startup.\n")
		} else {
			commands = HandleRun(quadctl, quadlets)
		}
	case "stop":
		if quadctl.IsSystemd {
			commands = HandleSystemdStop(quadctl, quadlets, false)
		} else {
			commands = HandleStop(quadlets)
		}
	case "remove", "rm":
		if quadctl.IsSystemd {
			commands = HandleSystemdRemove(quadctl, quadlets)
		} else {
			commands = HandleRemove(quadlets)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", quadctl.Subcommand)
		util.PrintUsage()
		os.Exit(1)
	}

	if len(commands) > 0 {
		RunCommands(quadctl, commands)
	}
}

func initState() {

	quadctl = &util.Quadctl{
		QuadletSchemas:    map[string]map[string]SchemaOption{},
		Config:            map[string]string{},
		IsRootful:         false,
		IsSystemd:         false,
		IsPrintOnly:       false,
		IsVerbose:         false,
		IsFile:            false,
		ListDepth:         2,
		Subcommand:        "",
		SearchDir:         "",
		PodmanArgs:        "",
		RunCmd:            "",
		QuadletSrcPath:    "",    // Path to the user's source directory containing quadlet folders or files
		UseSubdirectories: true,  // Default to installing quadlets in a subdirectory to keep them organized
		UseSymbolicLinks:  false, // Default to copying files for installation to avoid potential issues with source files being moved or deleted, but can be configured to use symbolic links for a more dynamic setup
		IsReloadSystemd:   true,  // Default to reloading systemd after installation to apply changes immediately
		IsRemoveVolumes:   true,  // Default to removing volumes on uninstall since they are often not needed after uninstall and can be left behind if not removed, but can be configured to keep volumes for data persistence.
		IsRemoveNetworks:  true,  // Default to removing networks on uninstall since they are often not needed after uninstall and can be left behind if not removed, but can be configured to keep volumes for data persistence.
		QuadletRootPath:   "/etc/containers/systemd",
		QuadletUserPath:   "/etc/containers/systemd/users",
	}
	quadctl.SystemdStartTmpl = template.Must(template.New("systemdStart").Parse("systemctl {{.user}} start"))
	quadctl.SystemdStopTmpl = template.Must(template.New("systemdStop").Parse("systemctl {{.user}} stop"))
	quadctl.SystemdStatusTmpl = template.Must(template.New("systemdStatus").Parse("systemctl {{.user}} status"))
	quadctl.SystemdReloadTmpl = template.Must(template.New("systemdReload").Parse("systemctl {{.user}} daemon-reload"))
	quadctl.SystemdLogsTmpl = template.Must(template.New("systemdLogs").Parse("journalctl {{.user}} -xe"))

	// Determine if running as root
	if os.Geteuid() == 0 {
		quadctl.IsRootful = true
	}
}

func displayQuadletSelector(quadctl *util.Quadctl) error {
	quadletDirs, err := util.ListSubdirectories(quadctl.QuadletSrcPath)
	if err != nil {
		return err
	}

	if len(quadletDirs) == 0 {
		return fmt.Errorf("no quadlet directories found in %s", quadctl.QuadletSrcPath)
	}
	selected, err := util.SelectFromList(quadletDirs)
	if err != nil {
		return err
	}
	quadctl.SearchDir = filepath.Join(quadctl.QuadletSrcPath, selected)
	return nil
}
