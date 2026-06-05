package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/fkmiec/quadctl/util"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/jedib0t/go-pretty/v6/table"
)

// --- CORE LOGIC HANDLERS ---

func HandlePull(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	images := []string{}
	for _, q := range quadlets {
		if q.Type == ".container" {
			if imgSec, ok := q.Sections["Container"]; ok {
				if imgList, ok := imgSec["Image"]; ok && len(imgList) > 0 {
					images = append(images, imgList[0])
				}
			}
		} else if q.Type == ".kube" {
			for _, res := range q.KubeResources {
				if res["type"] == "container" {
					images = append(images, res["image"].(string))
				}
			}
		}
	}

	for _, img := range images {
		//fmt.Printf("=> Pulling image: %s\n", img)
		c := NewCommand(fmt.Sprintf("Pulling image %s", img))
		c.Cmd = []string{"podman", "pull", img}
		commands = append(commands, c)
	}

	return commands
	//RunCommands(quadctl, commands)
}

// handleCreate generates and executes 'podman create' commands for all resources, but first checks if they exist and prints warnings if they do,
// suggesting to run 'remove' first if intent is to re-create. It also handles special cases like auto-restart configuration warnings.
func HandleCreate(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	for _, q := range quadlets {

		// For .kube, podman kube play will be called in start step. Return a no-op command with a warning here if verbose output is enabled.
		if q.Type == ".kube" && quadctl.IsVerbose {
			cmd := NewCommand(fmt.Sprintf("Creating %s %s", q.Type, q.ID))
			cmd.Warnings = append(cmd.Warnings, fmt.Sprintf("Podman kube play handles creation. Nothing to do for %s %s", q.Type, q.ID))
			cmd.Cmd = []string{"echo"}
			commands = append(commands, cmd)
			continue
		}

		//Only create if resource doesn't exist.
		if !resourceExists(q.Type, q.ID) {
			// For 'run' command, skip creating containers since 'podman run' will create them if they don't exist.
			if quadctl.Subcommand == "run" && q.Type == ".container" {
				continue
			}
			args, warns := generateCreateCommand(quadctl, q)
			cmd := NewCommand(fmt.Sprintf("Creating %s %s", q.Type, q.ID))
			cmd.Cmd = args

			for _, w := range warns {
				cmd.Warnings = append(cmd.Warnings, fmt.Sprintf("%s: %s", filepath.Base(q.Filepath), w))
			}

			// Warn about restart policy configuration, if applicable
			if q.RestartPolicy != "" && q.RestartPolicy != "no" {
				cmd.Warnings = append(cmd.Warnings, fmt.Sprintf("[INFO] %s: Restart policy configured (%s). Ensure podman-restart.service is enabled.\n", q.Filepath, q.RestartPolicy))
			}
			// Warn about AutoUpdate configuration, if applicable
			if q.GeneratedNames["auto_update"] != "" {
				cmd.Warnings = append(cmd.Warnings, fmt.Sprintf("[INFO] %s: Image AutoUpdate enabled (%s)\n", q.Filepath, q.GeneratedNames["auto_update"]))
			}

			commands = append(commands, cmd)

		} else {
			if quadctl.IsVerbose {
				cmd := NewCommand(fmt.Sprintf("Creating %s %s", q.Type, q.ID))
				cmd.Cmd = []string{"echo"}
				cmd.Warnings = append(cmd.Warnings, fmt.Sprintf(" [INFO] %s %s already exists. To force re-creation of ALL resources, run 'quadctl remove' first.\n", q.Type, q.ID))
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}

// Call handleCreate. Then start.
func HandleStart(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	//Create, if necessary
	cmds := HandleCreate(quadctl, quadlets)
	commands = append(commands, cmds...)

	// Stop if already running (podman ps -a only returns a list if systemd services are running. Once stopped, it returns empty.)
	if info, err := getContainerPS(quadlets); err == nil && len(info) > 0 {
		if strings.Contains(info[0][3], "Up") {
			cmd := HandleStop(quadctl, quadlets)
			commands = append(commands, cmd...)
		}
	}

	//Start
	for _, q := range quadlets {
		// Use generateStartupCommands
		cmd, warns := generateStartupCommand(quadctl, q)

		//fmt.Printf("Command for quadlet %s is %v\n", q.ID, cmd)

		if len(cmd) > 0 {
			c := NewCommand(fmt.Sprintf("Starting %s %s", q.Type, q.ID))
			c.Cmd = cmd
			c.Warnings = warns
			commands = append(commands, c)
		}
	}
	return commands
}

// Call handleCreate. Then start.
func HandleRun(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	//Check how many .container quadlets there are and how many with --detach or -d podman args.
	//If more than one .container and more than one of them don't have --detach or -d,
	//print a warning and exit.
	nonDetachedContainers := 0
	var foregroundQuadlet *util.Quadlet
	var foregroundQuadletCommand Command
	for _, q := range quadlets {
		if q.Type == ".container" {
			pArgs := q.Sections["Container"]["PodmanArgs"]
			if !slices.Contains(pArgs, "--detach") && !slices.Contains(pArgs, "-d") {
				foregroundQuadlet = q
				nonDetachedContainers++
			}
		}
	}
	if nonDetachedContainers > 1 {
		fmt.Fprintf(os.Stderr, "Error: 'quadctl run' can only run one container in the foreground. Add --detach or -d to PodmanArgs for all other .container quadlets. Execute quadctl run --help for details.\n")
		os.Exit(1)
	}

	commands := []Command{}

	//Create non-container resources, if necessary (HandleCreate will skip .container quadlets for the 'run' command, but create volumes, networks, pods if needed)
	c := HandleCreate(quadctl, quadlets)
	commands = append(commands, c...)

	//Start
	for _, q := range quadlets {
		// Only run containers and kubes. Pods, networks and volumes will be started/created as needed by the containers.
		if q.Type != ".container" && q.Type != ".kube" {
			continue
		}
		// For 'run' command, we need to generate 'podman run' commands instead of 'podman start' for containers.
		cmd, warns := generateRunCommand(quadctl, q)
		warnings := []string{}
		for _, w := range warns {
			warnings = append(warnings, fmt.Sprintf("%s: %s\n", filepath.Base(q.Filepath), w))
		}
		if len(cmd) > 0 {
			command := NewCommand(fmt.Sprintf("Running %s %s", q.Type, q.ID))
			command.Cmd = cmd
			command.Warnings = warnings

			if foregroundQuadlet != nil && q.ID == foregroundQuadlet.ID {
				foregroundQuadletCommand = command
				continue
			}
			commands = append(commands, command)
		}
	}
	if foregroundQuadlet != nil {
		// Run the foreground container command last since it will block and we want all other containers to be up before it runs.
		commands = append(commands, foregroundQuadletCommand)
	}
	return commands
}

func HandleStop(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	// Reverse order for safe stopping
	for i := len(quadlets) - 1; i >= 0; i-- {
		q := quadlets[i]
		cmd := generateStopCommand(quadctl, q)
		if len(cmd) > 0 {
			c := NewCommand(fmt.Sprintf("Stopping %s %s", q.Type, q.ID))
			c.Cmd = cmd
			commands = append(commands, c)
		}
	}
	return commands
}

func HandleRemove(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	// Reverse order for safe removal
	for i := len(quadlets) - 1; i >= 0; i-- {
		q := quadlets[i]
		resType := q.Type
		resName := q.ID
		if q.Type == ".container" {
			resName = q.GeneratedNames["container"]
		}

		rmCmd := []string{"podman"}
		switch resType {
		case ".kube":
			if quadctl.IsRemoveVolumes || quadctl.IsRemoveNetworks || q.Sections["Kube"]["KubeDownForce"][0] == "true" {
				rmCmd = append(rmCmd, "play", "kube", "--down", "--force", q.KubernetesYaml)
			} else {
				rmCmd = append(rmCmd, "play", "kube", "--down", q.KubernetesYaml)
			}
		case ".container":
			rmCmd = append(rmCmd, "container", "rm", "-f", resName)
		case ".pod":
			rmCmd = append(rmCmd, "pod", "rm", "-f", resName)
		case ".network":
			rmCmd = append(rmCmd, "network", "rm", resName)
		case ".volume":
			rmCmd = append(rmCmd, "volume", "rm", resName)
		}

		c := NewCommand(fmt.Sprintf("Removing %s %s", resType, resName))
		c.Cmd = rmCmd
		commands = append(commands, c)
	}
	return commands
}

func HandleSystemdCreate(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	var targetDir string

	if quadctl.IsRootful {
		targetDir = quadctl.QuadletRootPath
	} else {
		targetDir = quadctl.QuadletUserPath
	}

	// Ensure permissions to write to the target directory
	fileInfo, err := os.Stat(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing quadlet path %s: %v.\n", targetDir, err)
		if targetDir == quadctl.QuadletUserPath {
			fmt.Fprintf(os.Stderr, "If installing rootless quadlets to /etc/... or /run/... you may need to grant your user write permissions to the target directory.\n")
		}
		os.Exit(1)
	} else {
		if !fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "Quadlet path %s is not a directory. Ensure the path points to a directory and try again.\n", targetDir)
			os.Exit(1)
		}
		perm := fileInfo.Mode().Perm()
		if perm&0200 != 0200 && perm&0020 != 0020 && perm&0002 != 0002 {
			fmt.Fprintf(os.Stderr, "Quadlet path %s is not writable. Ensure the directory is writable and try again.\n", targetDir)
			if targetDir == quadctl.QuadletUserPath {
				fmt.Fprintf(os.Stderr, "If installing rootless quadlets to /etc/containers/systemd... or /usr/share/containers/systemd... you may need to grant your user write permissions to the target directory.\n")
			}
			os.Exit(1)
		}
	}

	c := NewCommand(fmt.Sprintf("Systemd installing quadlets to %s", targetDir))
	if quadctl.IsVerbose {
		c.PreFn = func(c *Command) {}
		c.PostFn = func(c *Command) {}
	}

	// Systemd create is mostly file operations.
	// For file operations, we use golang functions rather than podman, systemd or bash commands ...
	// Encapsulate code to run in a slice of functions that will be executed in a custom command when the command is run.
	funcs := []func(){}
	c.Output = append(c.Output, fmt.Sprintf("Creating target directory %s", targetDir))
	f := func() {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating target directory: %v\n", err)
			os.Exit(1)
		}
	}
	funcs = append(funcs, f)

	// If there was a .quadlets file, all the quadlets were extracted to and/or copied to a temp directory.
	// Replace the original search directory with the temp directory for copy operations involve in systemd create op
	searchDir := quadctl.SearchDir
	if quadctl.DotQuadletsPath != "" {
		searchDir = quadctl.DotQuadletsPath
		if quadctl.UseSymbolicLinks {
			fmt.Fprintf(os.Stderr, "Error: Cannot use symbolic links with .quadlets files.\n  The individual quadlets in a .quadlets file must be extracted to a temp directory before install to systemd.\n  Cannot link to temp directory.\n")
			os.Exit(1)
		}
	}

	// Use links if configured to do so
	if quadctl.UseSymbolicLinks {
		c.Output = append(c.Output, "Using symbolic links for installation.")
		if quadctl.UseSubdirectories {
			// Link the entire source directory as a subdirectory in the target location to keep related quadlets together
			dest := filepath.Join(targetDir, filepath.Base(searchDir))
			c.Output = append(c.Output, fmt.Sprintf("Linking directory %s -> %s", dest, searchDir))
			f := func() {
				if err := os.Symlink(searchDir, dest); err != nil {
					//if err := runCommand([]string{prefix, "ln", "-s", sourceDir, filepath.Join(targetDir, filepath.Base(sourceDir))}); err != nil {
					fmt.Fprintf(os.Stderr, "Error linking target directory: %v\n", err)
					os.Exit(1)
				}
			}
			funcs = append(funcs, f)
		} else {
			// Link the individual quadlet files directly into the target location
			for _, q := range quadlets {
				dest := filepath.Join(targetDir, filepath.Base(q.Filepath))
				c.Output = append(c.Output, fmt.Sprintf("Linking %s -> %s", dest, q.Filepath))
				f := func() {
					if err := os.Symlink(q.Filepath, dest); err != nil {
						//if err := runCommand([]string{prefix, "ln", "-s", q.Filepath, dest}); err != nil {
						fmt.Fprintf(os.Stderr, " Failed to link: %v\n", err)
						os.Exit(1)
					}
				}
				funcs = append(funcs, f)
				// Also link drop-in directory if exists
				dropInDir := q.Filepath + ".d"
				if info, err := os.Stat(dropInDir); err == nil && info.IsDir() {
					destDropIn := dest + ".d"
					c.Output = append(c.Output, fmt.Sprintf("Linking directory %s -> %s", destDropIn, dropInDir))
					f := func() {
						if err := os.Symlink(dropInDir, destDropIn); err != nil {
							//if err := runCommand([]string{prefix, "ln", "-s", dropInDir, destDropIn}); err != nil {
							fmt.Fprintf(os.Stderr, "  Failed to link dir: %v\n", err)
							os.Exit(1)
						}
					}
					funcs = append(funcs, f)
				}
			}
		}
		// Otherwise copy files to the target directory using podman quadlet install
	} else {
		var destDropIn string
		// If the user configured to use a subdirectory to organize quadlets, we create the directory and move files after podman quadlet install step.
		if quadctl.UseSubdirectories {
			//Create the subdirectory at target location
			dest := filepath.Join(targetDir, filepath.Base(searchDir))
			c.Output = append(c.Output, fmt.Sprintf("Copying directory %s to %s", filepath.Base(searchDir), dest))
			f := func() {
				if err := util.CopyDir(searchDir, dest); err != nil {
					fmt.Fprintf(os.Stderr, "  Failed to copy dir: %v\n", err)
					os.Exit(1)
				}
			}
			funcs = append(funcs, f)
		} else {
			for _, q := range quadlets {
				c.Output = append(c.Output, fmt.Sprintf("Copying file %s to %s", filepath.Base(q.Filepath), filepath.Join(targetDir, filepath.Base(q.Filepath))))
				f := func() {
					if err := util.CopyFile(q.Filepath, filepath.Join(targetDir, filepath.Base(q.Filepath))); err != nil {
						fmt.Fprintf(os.Stderr, "  Failed to copy file: %v\n", err)
						os.Exit(1)
					}
				}
				funcs = append(funcs, f)
			}
		}
		// Copy drop-in directories if exist
		for _, q := range quadlets {
			dropInDir := q.Filepath + ".d"
			if info, err := os.Stat(dropInDir); err == nil && info.IsDir() {

				// Set dropInDir
				if quadctl.UseSubdirectories {
					destDropIn = filepath.Join(targetDir, filepath.Base(searchDir), filepath.Base(q.Filepath)+".d")
				} else {
					destDropIn = filepath.Join(targetDir, filepath.Base(q.Filepath)+".d")
				}
				c.Output = append(c.Output, fmt.Sprintf("Copying directory %s to %s", filepath.Base(dropInDir), destDropIn))
				f := func() {
					if err := util.CopyDir(dropInDir, destDropIn); err != nil {
						fmt.Fprintf(os.Stderr, "  Failed to copy dir: %v\n", err)
						os.Exit(1)
					}
				}
				funcs = append(funcs, f)
			}
		}
	}

	// Custom run function that will, when executed by runCommands(), execute the anonymous functions created above.
	c.RunFn = func(c *Command) {
		for _, f := range funcs {
			f()
		}
		if quadctl.IsVerbose {
			fmt.Println(c.Label + "... Done")
			for _, line := range c.Output {
				fmt.Println(" => " + line)
			}
		}
	}

	commands = append(commands, c)

	// Reload systemd to recognize the new quadlet services
	commands = append(commands, HandleSystemdReload(quadctl)...)

	return commands
}

func HandleSystemdStart(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {
	//Ideally, call handleInstall if needed. How to check if the required systemd services are installed?
	/*
		❯ sudo podman quadlet list
		NAME                   UNIT NAME                    PATH ON DISK                                           STATUS      APPLICATION
		homebox-app.container  homebox-app.service          /etc/containers/systemd/homebox/homebox-app.container  Not loaded
		homebox-data.volume    homebox-data-volume.service  /etc/containers/systemd/homebox/homebox-data.volume    Not loaded
		homebox.pod            homebox-pod.service          /etc/containers/systemd/homebox/homebox.pod            Not loaded
	*/

	commands := []Command{}

	// Create if not existing
	info, _ := listSystemdInstalledQuadlets(quadlets)
	if len(info) < len(quadlets) {
		//fmt.Printf("installed count: %d, quadlet count: %d\n", len(info), len(quadlets))
		cmd := HandleSystemdCreate(quadctl, quadlets)
		commands = append(commands, cmd...)
	} else {
		// Reload quadlet definitions if not done as part of create step
		cmd := HandleSystemdReload(quadctl)
		commands = append(commands, cmd...)
	}

	// Stop if already running (podman ps -a only returns a list if systemd services are running. Once stopped, it returns empty.)
	if info, err := getContainerPS(quadlets); err == nil && len(info) > 0 {
		cmd := HandleSystemdStop(quadctl, quadlets, false)
		commands = append(commands, cmd...)
	}

	// Start the systemd services
	var buf bytes.Buffer
	data := map[string]string{}
	if !quadctl.IsRootful {
		data["user"] = "--user"
	}

	err := quadctl.SystemdStartTmpl.Execute(&buf, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing systemd start template: %v\n", err)
		os.Exit(1)
	}

	// Only start the pod and any loose containers
	for _, q := range quadlets {
		if (q.Type == ".container" && q.ParentPod == "") || q.Type == ".pod" || q.Type == ".kube" {
			args := util.ParseFields(buf.String())
			args = append(args, q.ServiceName)
			cmd := NewCommand(fmt.Sprintf("Starting %s %s", q.Type, q.ID))
			cmd.Cmd = args
			commands = append(commands, cmd)
		}

		// For networks and volumes, we rely on the fact that systemd will start them automatically when the containers that depend on them are started.
	}
	return commands
}

func HandleSystemdStop(quadctl *util.Quadctl, quadlets []*util.Quadlet, stopNetAndVol bool) []Command {

	commands := []Command{}

	// Stop the systemd services
	var buf bytes.Buffer
	data := map[string]string{}
	if !quadctl.IsRootful {
		data["user"] = "--user"
	}
	err := quadctl.SystemdStopTmpl.Execute(&buf, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing systemd stop template: %v\n", err)
		os.Exit(1)
	}

	for _, q := range quadlets {
		var args []string
		// Stop a container directly only if it is not part of a pod.
		if (q.Type == ".container" && q.ParentPod == "") || q.Type == ".pod" || q.Type == ".kube" {
			// Stop the pod and any related containers.
			args = util.ParseFields(buf.String())
			args = append(args, q.ServiceName)
		} else {
			// Stop network and volume services (Only used when called by handleUninstall. Ensures cleanup of volumes and networks).
			if stopNetAndVol && (q.Type == ".network" || q.Type == ".volume") {
				args = util.ParseFields(buf.String())
				args = append(args, q.ServiceName)
			}
		}
		if len(args) == 0 {
			continue
		}
		cmd := NewCommand(fmt.Sprintf("Systemd stopping %s %s", q.Type, q.ID))
		cmd.Cmd = args
		commands = append(commands, cmd)
	}
	return commands
}

func HandleSystemdRemove(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {
	var targetDir string
	if quadctl.IsRootful {
		targetDir = quadctl.QuadletRootPath
	} else {
		targetDir = quadctl.QuadletUserPath
	}

	commands := []Command{}

	// Ensure any running services are stopped before uninstalling
	cmds := HandleSystemdStop(quadctl, quadlets, true)
	commands = append(commands, cmds...)

	// Systemd removal is mostly file operations.
	// For file operations, we use golang functions rather than podman, systemd or bash commands ...
	// Encapsulate code to run in a slice of functions that will be executed in a custom command when the command is run.
	funcs := []func(){}
	c := NewCommand(fmt.Sprintf("Removing quadlets from %s", targetDir))
	if quadctl.IsVerbose {
		c.PreFn = func(c *Command) {}
		c.PostFn = func(c *Command) {}
	}

	//If targetDir exists, remove files.
	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		if quadctl.UseSymbolicLinks {
			if quadctl.UseSubdirectories {
				//remove link to directory
				link := filepath.Join(targetDir, filepath.Base(quadctl.SearchDir))
				c.Output = append(c.Output, fmt.Sprintf("Removing symbolic link: %s", link))
				f := func() {
					_ = os.Remove(link)
				}
				funcs = append(funcs, f)
			} else {
				//remove individual file links
				for _, q := range quadlets {
					dest := filepath.Join(targetDir, filepath.Base(q.Filepath))
					c.Output = append(c.Output, fmt.Sprintf("Removing symbolic link: %s", dest))
					f := func() {
						if err := os.Remove(dest); err != nil {
							fmt.Fprintf(os.Stderr, "Failed to remove %s: %v\n", dest, err)
						}
					}
					funcs = append(funcs, f)
					// Also remove link to drop-in directory if exists
					dropInDir := dest + ".d"
					if info, err := os.Stat(dropInDir); err == nil && info.IsDir() {
						c.Output = append(c.Output, fmt.Sprintf("Removing symbolic link: %s", dropInDir))
						f := func() {
							if err := os.Remove(dropInDir); err != nil {
								fmt.Fprintf(os.Stderr, "Failed to remove symlink to drop-in dir %s: %v\n", dropInDir, err)
							}
						}
						funcs = append(funcs, f)
					}
				}
			}
		} else {
			if quadctl.UseSubdirectories {
				//remove directory and all files within
				dest := filepath.Join(targetDir, filepath.Base(quadctl.SearchDir))
				c.Output = append(c.Output, fmt.Sprintf("Removing directory and files at: %s", dest))
				f := func() {
					_ = os.RemoveAll(dest)
				}
				funcs = append(funcs, f)
			} else {
				for _, q := range quadlets {
					file := filepath.Join(targetDir, filepath.Base(q.Filepath))
					if info, err := os.Stat(file); err == nil && info.IsDir() {
						c.Output = append(c.Output, fmt.Sprintf("Removing file: %s", file))
						f := func() {
							if err := os.Remove(file); err != nil {
								fmt.Fprintf(os.Stderr, "Failed to remove file %s: %v\n", file, err)
							}
						}
						funcs = append(funcs, f)
					}
				}
			}
		}

		//Expressly remove volume and network resources that might be left behind
		for _, q := range quadlets {
			if q.Type == ".volume" && quadctl.IsRemoveVolumes {
				c.Output = append(c.Output, fmt.Sprintf("Removing volume %s", q.ID))
				var fn func()
				//Default name has systemd- prefix. If non-default name was specified, use it, otherwise use default prefix.
				if volName := q.Sections["Volume"]["VolumeName"]; volName != nil {
					fn = func() {
						_ = runCommandSilently([]string{"podman", "volume", "rm", "-f", volName[0]})
					}
				} else {
					fn = func() {
						_ = runCommandSilently([]string{"podman", "volume", "rm", "-f", "systemd-" + q.ID})
					}
				}
				funcs = append(funcs, fn)
			}
			if q.Type == ".network" && quadctl.IsRemoveNetworks {
				c.Output = append(c.Output, fmt.Sprintf("Removing network %s", q.ID))
				var fn func()
				//Default name has systemd- prefix. If non-default name was specified, use it, otherwise use default prefix.
				if networkName := q.Sections["Network"]["NetworkName"]; networkName != nil {
					fn = func() {
						_ = runCommandSilently([]string{"podman", "network", "rm", "-f", networkName[0]})
					}
				} else {
					fn = func() {
						_ = runCommandSilently([]string{"podman", "network", "rm", "-f", "systemd-" + q.ID})
					}
				}
				funcs = append(funcs, fn)
			}
		}
	}

	// Custom run function that will, when executed by runCommands(), execute the anonymous functions created above.
	c.RunFn = func(c *Command) {
		for _, f := range funcs {
			f()
		}
		if quadctl.IsVerbose {
			fmt.Println(c.Label + "... Done")
			for _, line := range c.Output {
				fmt.Println(" => " + line)
			}
		}
	}

	commands = append(commands, c)

	// Reload systemd to ensure it picks up the changes after removal.
	cmds = HandleSystemdReload(quadctl)
	commands = append(commands, cmds...)

	return commands
}

func HandleSystemdStatus(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	if quadctl.IsLongStatus {
		commands := []Command{}

		var buf bytes.Buffer
		data := map[string]string{}
		if !quadctl.IsRootful {
			data["user"] = "--user"
		}
		err := quadctl.SystemdStatusTmpl.Execute(&buf, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing systemd status template: %v\n", err)
			os.Exit(1)
		}
		args := util.ParseFields(buf.String())
		for _, q := range quadlets {
			args = append(args, q.ServiceName)
		}
		if quadctl.IsPrintOnly {
			c := NewCommand("Getting systemd status")
			c.Cmd = args
			commands = append(commands, c)
		} else {
			runCommand(args)
		}
		return commands
	} else {
		displayListOfSystemdInstalledQuadlets(quadlets)
		return []Command{}
	}
}

func HandleSystemdLogs(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	commands := []Command{}

	var buf bytes.Buffer
	data := map[string]string{}
	if !quadctl.IsRootful {
		data["user"] = "--user"
	}
	err := quadctl.SystemdLogsTmpl.Execute(&buf, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing systemd logs: %v\n", err)
		os.Exit(1)
	}

	cmd := util.ParseFields(buf.String())
	if quadctl.IsPrintOnly {
		c := NewCommand("Opening systemd logs")
		c.Cmd = cmd
		commands = append(commands, c)
	} else {
		runCommand(cmd)
	}
	return commands
}

func HandleSystemdReload(quadctl *util.Quadctl) []Command {
	var buf bytes.Buffer
	data := map[string]string{}
	if !quadctl.IsRootful {
		data["user"] = "--user"
	}
	err := quadctl.SystemdReloadTmpl.Execute(&buf, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing systemd reload template: %v\n", err)
		os.Exit(1)
	}
	command := util.ParseFields(buf.String())
	cmd := NewCommand("Reloading systemd")
	cmd.Cmd = command
	return []Command{cmd}
}

func HandlePS(quadctl *util.Quadctl, quadlets []*util.Quadlet) {

	psInfo, err := getContainerPS(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"CONTAINER ID", "NAME", "POD", "STATE", "PORTS", "IMAGE", "CREATED"})
	format := "2006-01-02 15:04:05.999999999 -0700 MST"
	for _, info := range psInfo {
		if len(info) >= 7 {

			createdDatetime, err := time.Parse(format, strings.TrimSpace(info[6]))
			createdDuration := "unknown"
			if err == nil {
				createdDuration = time.Since(createdDatetime).Round(time.Second).String() + " ago"
			}
			t.AppendRow(table.Row{
				strings.TrimSpace(info[0]),
				strings.TrimSpace(info[1]),
				strings.TrimSpace(info[2]),
				strings.TrimSpace(info[3]),
				strings.TrimSpace(info[4]),
				strings.TrimSpace(info[5]),
				strings.TrimSpace(createdDuration),
			})
		}
	}
	t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
	t.Render()
}

func HandleStats(quadctl *util.Quadctl, quadlets []*util.Quadlet) {

	psInfo, err := getContainerPS(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if len(psInfo) < 1 {
		fmt.Printf("Error: Found no containers running or created that are related to quadlets in directory: %s\n", quadctl.SearchDir)
		os.Exit(1)
	}

	//cmd := []string{"podman", "stats", "--no-stream"}
	cmd := []string{"podman", "stats"}

	for _, info := range psInfo {
		id := strings.TrimSpace(info[0])
		cmd = append(cmd, id)
	}

	_ = runCommand(cmd)
}

func HandleImages(quadlets []*util.Quadlet) {

	//REPOSITORY                                 TAG         IMAGE ID      CREATED       SIZE
	cmd := []string{"podman", "images", "--noheading", "--filter", "reference=ADD_ID_HERE", "--format", "{{.Repository}}|{{.Tag}}|{{.ID}}|{{.Created}}|{{.Size}}"}
	imageInfo := [][]string{}

	// Fetch image info for each container
	psInfo, err := getContainerPS(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if len(psInfo) > 0 {
		for _, info := range psInfo {
			name := strings.TrimSpace(info[5]) // IMAGE ID from ps output
			if len(name) < 12 {
				continue
			}
			cmd[4] = "reference=" + name
			output, err := runCommandCapture(cmd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error fetching image info for container %s: %v\n", info[0], err)
				continue
			}
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				parts := strings.Split(line, "|")
				if len(parts) >= 5 {
					imageInfo = append(imageInfo, parts)
				} else {
					// Typically an empty newline
					//fmt.Printf("Warning: Unexpected output from podman ps. Expected 5 or more values. Got: %s\n", line)
					continue
				}
			}
		}
	} else {
		// If no containers are found, we can still fetch image info for the quadlet files
		fmt.Fprintf(os.Stderr, "No containers found, fetching image info from quadlet files...\n")
		for _, q := range quadlets {
			// Images only pertain to containers
			if q.Type == ".container" {
				if imgSec, ok := q.Sections["Container"]; ok {
					if imgList, ok := imgSec["Image"]; ok && len(imgList) > 0 {
						name := strings.TrimSpace(imgList[0]) // IMAGE ID from quadlet file
						if len(name) < 12 {
							continue
						}
						cmd[4] = "reference=" + name
						output, err := runCommandCapture(cmd)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error fetching image info for quadlet %s: %v\n", q.ID, err)
							continue
						}
						lines := strings.Split(output, "\n")
						for _, line := range lines {
							parts := strings.Split(line, "|")
							if len(parts) >= 5 {
								imageInfo = append(imageInfo, parts)
							}
						}
					}
				}
			} else if q.Type == ".kube" {
				for _, res := range q.KubeResources {
					if res["type"] == "container" {
						name := strings.TrimSpace(res["image"].(string)) // IMAGE ID from quadlet file
						if len(name) < 12 {
							continue
						}
						cmd[4] = "reference=" + name
						output, err := runCommandCapture(cmd)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error fetching image info for .kube %s container %s: %v\n", q.ID, res["name"].(string), err)
							continue
						}
						lines := strings.Split(output, "\n")
						for _, line := range lines {
							parts := strings.Split(line, "|")
							if len(parts) >= 5 {
								imageInfo = append(imageInfo, parts)
							}
						}
					}
				}
			}
		}
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"REPOSITORY", "TAG", "IMAGE ID", "CREATED", "SIZE"})
	for _, info := range imageInfo {
		if len(info) >= 5 {
			t.AppendRow(table.Row{
				strings.TrimSpace(info[0]),
				strings.TrimSpace(info[1]),
				strings.TrimSpace(info[2]),
				strings.TrimSpace(info[3]),
				strings.TrimSpace(info[4]),
			})
		}
	}
	t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
	t.Render()
}

func HandleLogs(quadctl *util.Quadctl, quadlets []*util.Quadlet) []Command {

	var commands []Command

	cmd := []string{"podman", "logs"}
	var containerName string

	// Fetch image info for each container
	psInfo, err := getContainerPS(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return commands
	}

	if len(psInfo) > 0 {
		if len(psInfo) == 1 {
			containerName = psInfo[0][1]
		} else {
			names := []string{}
			for _, info := range psInfo {
				names = append(names, info[1])
			}
			selected, err := util.SelectFromList(names)
			if err != nil {
				fmt.Printf("Error getting container info: %v\n", err)
				os.Exit(1)
			}
			containerName = selected
		}
		cmd = append(cmd, containerName)
	}

	if quadctl.IsPrintOnly {
		c := NewCommand(fmt.Sprintf("Opening podman logs for %s\n", containerName))
		c.Cmd = cmd
		commands = append(commands, c)
	} else {
		runCommand(cmd)
	}
	return commands
}

func HandleList(quadctl *util.Quadctl) error {

	if !quadctl.IsListAll {
		absPath := quadctl.QuadletSrcPath
		if quadctl.IsSystemd {
			if quadctl.IsRootful {
				absPath = quadctl.QuadletRootPath
			} else {
				absPath = quadctl.QuadletUserPath
			}
		}
		return listQuadlets(absPath, quadctl.ListDepth)
	} else {
		err := listQuadlets(quadctl.QuadletSrcPath, quadctl.ListDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing quadlets in search directory: %v\n", err)
			os.Exit(1)
		}
		err = listQuadlets(quadctl.QuadletRootPath, quadctl.ListDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing quadlets in search directory: %v\n", err)
			os.Exit(1)
		}
		err = listQuadlets(quadctl.QuadletUserPath, quadctl.ListDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing quadlets in search directory: %v\n", err)
			os.Exit(1)
		}
		return nil
	}
}

func listQuadlets(absPath string, depth int) error {
	// Verify the path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		//try to create the directory
		if err = os.MkdirAll(absPath, 0660); err != nil {
			fmt.Printf("Error: Failed to stat configured quadlet.src.path: %v", err)
			os.Exit(1)
		}
	} else {
		if !info.IsDir() {
			fmt.Printf("Error: Configured quadlet.src.path is not a directory: %s\n", absPath)
			os.Exit(1)
		}
	}

	lw := list.NewWriter()
	lw.SetStyle(list.StyleConnectedRounded)

	// Append the root directory name
	lw.AppendItem(absPath)

	// Start recursive rendering (root is level 1, its children are level 2)
	lw.Indent()
	err = appendDirItems(lw, absPath, 2, depth)
	if err != nil {
		return err
	}
	lw.UnIndent()

	// Output the rendered list
	fmt.Println(lw.Render())
	return nil
}

// appendDirItems recursively traverses the directory and adds items to the list writer.
func appendDirItems(lw list.Writer, currentPath string, level int, depth int) error {
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Add the current file or directory to the list
		lw.AppendItem(entry.Name())

		// Nest deeper if it's a directory
		lw.Indent()
		if entry.IsDir() && depth > level {
			nextPath := filepath.Join(currentPath, entry.Name())
			if err := appendDirItems(lw, nextPath, level+1, depth); err != nil {
				return err
			}
		}
		lw.UnIndent()
	}

	return nil
}

// generateCreateCommand creates the base 'podman ... create' string.
func generateCreateCommand(quadctl *util.Quadctl, q *util.Quadlet) ([]string, []string) {
	var warnings []string
	var cmd []string

	// Warn about ignored sections
	for sec := range q.Sections {
		// standard systemd sections not used in CLI calls
		if sec == "Install" || sec == "Unit" {
			warnings = append(warnings, fmt.Sprintf("Ignoring [%s] section (Systemd specific)", sec))
		}
	}

	switch q.Type {
	case ".volume":
		//Get the schema for the volume type and use the PodmanTemplateParsed to format the podman option.
		options, ok := quadctl.QuadletSchemas["volume"]
		if !ok {
			warnings = append(warnings, "No volume schema found.")
			return cmd, warnings
		}
		cmd = append(cmd, "podman", "volume", "create")
		if volSec, ok := q.Sections["Volume"]; ok {
			cmd = append(cmd, getRawPodmanArgs(volSec)...)
			for k, vals := range volSec {
				for _, v := range vals {
					switch k {
					case "Type":
						continue // Type is not a Podman CLI option
					case "ServiceName":
						continue // ServiceName is for systemd and does not affect Podman CLI
					case "VolumeName":
						//cmd = append(cmd, "--name", v) // Not sure this is valid. May need to hold the value and append at the end after processing all options to avoid ordering issues with Podman CLI
						// The volume name is specified by the ID and added at the end of the command
						continue
					case "PodmanArgs": // Handled above
						continue
					default:
						podmanOpt, err := util.QuadletOptionToPodman("volume", options, k, v)
						if err != nil {
							warnings = append(warnings, err.Error())
							continue
						}
						// Use Fields to parse space-separated flags
						cmd = append(cmd, util.ParseFields(podmanOpt)...)
					}
				}
			}
		}
		cmd = append(cmd, q.ID)

	case ".network":
		//Get the schema for the network type and use the PodmanTemplateParsed to format the podman option.
		options, ok := quadctl.QuadletSchemas["network"]
		if !ok {
			warnings = append(warnings, "No network schema found.")
			return cmd, warnings
		}
		cmd = append(cmd, "podman", "network", "create")
		if netSec, ok := q.Sections["Network"]; ok {
			cmd = append(cmd, getRawPodmanArgs(netSec)...)
			for k, vals := range netSec {
				for _, v := range vals {
					switch k {
					case "NetworkName":
						continue // NetworkName is for systemd and does not affect Podman CLI
					case "ServiceName":
						continue // ServiceName is for systemd and does not affect Podman CLI
					case "NetworkDeleteOnStop":
						continue // NetworkDeleteOnStop is for systemd and does not affect Podman CLI
					case "PodmanArgs": // Handled above
					default:
						podmanOpt, err := util.QuadletOptionToPodman("network", options, k, v)
						if err != nil {
							warnings = append(warnings, err.Error())
							continue
						}
						// Use Fields to parse space-separated flags
						cmd = append(cmd, util.ParseFields(podmanOpt)...)
					}
				}
			}
		}
		cmd = append(cmd, q.ID)

	case ".pod":
		//Get the schema
		options, ok := quadctl.QuadletSchemas["pod"]
		if !ok {
			warnings = append(warnings, "No pod schema found.")
			return cmd, warnings
		}

		cmd = append(cmd, "podman", "pod", "create", "--name", q.ID)
		if podSec, ok := q.Sections["Pod"]; ok {
			cmd = append(cmd, getRawPodmanArgs(podSec)...)
			for k, vals := range podSec {
				for _, v := range vals {
					switch k {
					case "ServiceName":
						continue // ServiceName is for systemd and does not affect Podman CLI
					case "PodmanArgs": // Handled above
					default:
						podmanOpt, err := util.QuadletOptionToPodman("pod", options, k, v)
						if err != nil {
							warnings = append(warnings, err.Error())
							continue
						}
						// Use Fields to parse space-separated flags
						cmd = append(cmd, util.ParseFields(podmanOpt)...)
					}
				}
			}
		}

	case ".container":
		//Get the schema
		options, ok := quadctl.QuadletSchemas["container"]
		if !ok {
			warnings = append(warnings, "No container schema found.")
			return cmd, warnings
		}

		resName := q.GeneratedNames["container"]
		cmd = append(cmd, "podman", "container", "create", "--name", resName)

		// Map [Service] Restart= to --restart
		if q.RestartPolicy != "" {
			cmd = append(cmd, "--restart", q.RestartPolicy)
		}

		// Map [Container] AutoUpdate= to label
		//if q.GeneratedNames["auto_update"] != "" {
		//	cmd = append(cmd, "--label", "io.containers.autoupdate="+q.GeneratedNames["auto_update"])
		//}

		var image string
		var execCmd string
		if contSec, ok := q.Sections["Container"]; ok {
			configuredPodmanArgs := getRawPodmanArgs(contSec)

			// Special handling for quadctl run command. It's basically same as create, but allows for specifying podman args and a command to execute.
			if quadctl.PodmanArgs != "" {
				// If PodmanArgs were also provided via CLI, we will append them after the ones from the quadlet file.
				// This allows CLI args to override quadlet args if there are conflicts, since in Podman CLI the last specified flag takes precedence.
				configuredPodmanArgs = append(configuredPodmanArgs, util.ParseFields(quadctl.PodmanArgs)...)
			}
			if quadctl.RunCmd != "" {
				execCmd = quadctl.RunCmd
			}

			cmd = append(cmd, configuredPodmanArgs...)
			for k, vals := range contSec {
				opt, ok := options[k]
				if !ok {
					warnings = append(warnings, fmt.Sprintf("Quadlet container option not defined: %s", k))
					continue
				}
				// Check if multiple values and not supported
				if !opt.AllowMultiple && len(vals) > 1 {
					warnings = append(warnings, fmt.Sprintf("Option %s does not accept multiple space-separated values: '%s'\n", k, strings.Join(vals, " ")))
					continue
				}

				if k == "Exec" {
					// Exec is a special case since it's not a Podman CLI option. Append command and args to the end of the create command.
					// Ignore quadlet file Exec option if --exec flag was passed on the CLI
					if execCmd == "" {
						execCmd = strings.Join(vals, " ")
					}
					continue
				}

				for _, v := range vals {
					switch k {
					case "Image":
						image = v
					case "ReloadCmd":
						continue // ReloadCmd is for systemd and does not affect Podman CLI
					case "ReloadSignal":
						continue // ReloadSignal is for systemd and does not affect Podman CLI
					case "ServiceName":
						continue // ServiceName is for systemd and does not affect Podman CLI
					case "StartWithPod":
						continue // StartWithPod is for systemd and does not affect Podman CLI
					case "Volume":
						volSource := strings.Split(v, ":")[0]
						cleanVol := strings.TrimSuffix(volSource, ".volume")
						mapped := strings.Replace(v, volSource, cleanVol, 1)
						cmd = append(cmd, "-v", mapped)
					case "Network":
						cmd = append(cmd, "--network", strings.TrimSuffix(v, ".network"))
					case "PodmanArgs": // Handled above
					default:
						if k == "Pod" {
							v = strings.TrimSuffix(v, ".pod")
						}

						podmanOpt, err := util.QuadletOptionToPodman("container", options, k, v)
						if err != nil {
							warnings = append(warnings, err.Error())
							continue
						}
						// Use Fields to parse space-separated flags
						cmd = append(cmd, util.ParseFields(podmanOpt)...)
					}
				}
			}
		}
		if image == "" {
			warnings = append(warnings, "No Image= specified in [Container]")
			image = "<MISSING_IMAGE>"
		}
		cmd = append(cmd, image)
		if execCmd != "" {
			// If a command to execute is specified for the quadlet, the equivalent podman create command will have it appended at the end.
			cmd = append(cmd, execCmd)
		}
	}
	return cmd, warnings
}

// generateStartupCommand creates necessary 'start' commands based on existence.
func generateStartupCommand(quadctl *util.Quadctl, q *util.Quadlet) ([]string, []string) {
	cmd := []string{}
	warnings := []string{}
	resName := q.ID

	//Kube is a special case. kube play is create and start in one step, so we generate the play command here in "start" phase.
	if q.Type == ".kube" {
		//Get the schema for the kube type and use the PodmanTemplateParsed to format the podman option.
		options, ok := quadctl.QuadletSchemas["kube"]
		if !ok {
			warnings = append(warnings, "No kube schema found.")
			return cmd, warnings
		}

		cmd = append(cmd, "podman", "play", "kube")
		fmt.Printf("generateStartupCommand(%s): %v\n", q.ID, cmd)
		if kubeSec, ok := q.Sections["Kube"]; ok {
			cmd = append(cmd, getRawPodmanArgs(kubeSec)...)
			for k, vals := range kubeSec {
				for _, v := range vals {
					switch k {
					case "Yaml":
						continue // Yaml is parsed ahead of time and is appended at the end as the file argument for podman play kube
					case "ServiceName":
						continue // ServiceName is for systemd and does not affect Podman CLI
					case "PodmanArgs": // Handled above
						continue
					default:
						podmanOpt, err := util.QuadletOptionToPodman("kube", options, k, v)
						if err != nil {
							warnings = append(warnings, err.Error())
							continue
						}
						// Use Fields to parse space-separated flags
						cmd = append(cmd, util.ParseFields(podmanOpt)...)
					}
				}
			}
		}
		cmd = append(cmd, q.KubernetesYaml)
		return cmd, warnings
	}

	// Other startable types are pod and container
	if q.Type == ".container" {
		resName = q.GeneratedNames["container"]
	}

	// 3. Determine if we should start it
	shouldStart := true
	if q.Type == ".container" && q.ParentPod != "" {
		// Prompt: Create start commands ONLY for pods and loose containers
		shouldStart = false
	}

	if shouldStart {
		if q.Type == ".pod" {
			cmd = append(cmd, "podman", "pod", "start", resName)
		} else if q.Type == ".container" {
			cmd = append(cmd, "podman", "container", "start", resName)
		}
	} else if q.Type == ".container" {
		warnings = append(warnings, fmt.Sprintf(" [INFO] Container %s belongs to pod %s, it will start with the pod.\n", resName, q.ParentPod))
	}

	return cmd, warnings
}

// generateStartupCommand creates necessary 'start' commands based on existence.
func generateRunCommand(quadctl *util.Quadctl, q *util.Quadlet) ([]string, []string) {

	if q.Type == ".kube" {
		// For kube type, just reuse the generateStartupCommand for kube types.
		return generateStartupCommand(quadctl, q)
	}

	createCmd, warnings := generateCreateCommand(quadctl, q)
	runCmd := []string{"podman", "run"}
	runCmd = append(runCmd, createCmd[3:]...) // Replace 'podman container create' with 'podman run'

	return runCmd, warnings
}

func generateStopCommand(quadctl *util.Quadctl, q *util.Quadlet) []string {
	cmd := []string{}
	resName := q.ID
	if q.Type == ".container" {
		resName = q.GeneratedNames["container"]
	}

	switch q.Type {
	case ".kube":
		if quadctl.IsRemoveVolumes || quadctl.IsRemoveNetworks || q.Sections["Kube"]["KubeDownForce"][0] == "true" {
			cmd = append(cmd, []string{"podman", "play", "kube", "--down", "--force", q.KubernetesYaml}...)
		} else {
			cmd = append(cmd, []string{"podman", "play", "kube", "--down", q.KubernetesYaml}...)
		}
	case ".pod":
		cmd = append(cmd, []string{"podman", "pod", "stop", resName}...)
	case ".container":
		if q.ParentPod == "" {
			// loose container
			cmd = append(cmd, []string{"podman", "stop", resName}...)
		}
	}
	return cmd
}

// Helper: Get raw PodmanArgs securely
func getRawPodmanArgs(section map[string][]string) []string {
	var args []string
	for _, argStr := range section["PodmanArgs"] {
		// Use Fields to parse space-separated flags
		args = append(args, util.ParseFields(argStr)...)
	}
	return args
}

func resourceExists(qType string, name string) bool {
	inspectCmd := []string{"podman"}
	switch qType {
	case ".container":
		inspectCmd = append(inspectCmd, "container", "inspect", name)
	case ".pod":
		inspectCmd = append(inspectCmd, "pod", "inspect", name)
	case ".network":
		inspectCmd = append(inspectCmd, "network", "inspect", name)
	case ".volume":
		inspectCmd = append(inspectCmd, "volume", "inspect", name)
	default:
		return false
	}
	return runCommandSilently(inspectCmd) == nil
}

func listSystemdInstalledQuadlets(quadlets []*util.Quadlet) ([][]string, error) {
	cmd := []string{"podman", "quadlet", "list", "--format", "{{.Name}},{{.Path}},{{.UnitName}},{{.Status}}"}
	output, err := runCommandCapture(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching installed quadlets: %v\n", err)
		return nil, err
	}
	//.Printf("podman quadlet list:\n%s\n", output)
	lines := strings.Split(output, "\n")
	var info [][]string
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}
		//filter for our quadlets
		for _, q := range quadlets {
			name := filepath.Base(q.Filepath)
			if strings.TrimSpace(parts[0]) == name {
				info = append(info, parts)
				break
			}
		}
	}
	return info, nil
}

func getContainerPS(quadlets []*util.Quadlet) ([][]string, error) {
	cmd := []string{"podman", "ps", "-a", "--format", "{{.ID}}|{{.Names}}|{{.PodName}}|{{.Status}}|{{.Ports}}|{{.Image}}|{{.Created}}"}
	output, err := runCommandCapture(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(output, "\n")
	var psInfo [][]string
	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 7 {
			continue
		}
		//filter for containers that match our quadlet definitions by name or parent pod
		for _, q := range quadlets {
			//if q.Type == ".container" && strings.HasSuffix(parts[1], q.GeneratedNames["container"]) || (q.ParentPod != "" && strings.HasSuffix(parts[2], q.ParentPod)) {
			if q.Type == ".container" && strings.HasSuffix(parts[1], q.GeneratedNames["container"]) || (q.ParentPod != "" && strings.HasSuffix(parts[2], q.GeneratedNames["pod_name"])) {
				psInfo = append(psInfo, parts)
				break
			}
			if q.Type == ".kube" {
				for _, res := range q.KubeResources {
					if res["type"] == "container" && strings.HasSuffix(parts[1], res["name"].(string)) || (res["pod"] != nil && strings.HasSuffix(parts[2], res["pod"].(string))) {
						psInfo = append(psInfo, parts)
						break
					}
				}
			}
		}
	}
	return psInfo, nil
}

func displayListOfSystemdInstalledQuadlets(quadlets []*util.Quadlet) error {
	/*
		//podman quadlet list --format "{{.Name}}|{{.UnitName}}|{{.Path}}|{{.Status}}\n"
		cmd := []string{"podman", "quadlet", "list", "--format", "{{.Name}}|{{.UnitName}}|{{.Path}}|{{.Status}}"}
		output, err := runCommandCapture(cmd)
		if err != nil {
			return err
		}
		info := [][]string{}
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			parts := strings.Split(line, "|")
			if len(parts) < 4 {
				continue
			}
			info = append(info, parts)
		}
	*/
	info, err := listSystemdInstalledQuadlets(quadlets)
	if err != nil {
		return err
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "UNIT NAME", "PATH", "STATUS"})
	for _, quadletInfo := range info {
		if len(quadletInfo) >= 4 {
			t.AppendRow(table.Row{
				strings.TrimSpace(quadletInfo[0]),
				strings.TrimSpace(quadletInfo[1]),
				strings.TrimSpace(quadletInfo[2]),
				strings.TrimSpace(quadletInfo[3]),
			})
		}
	}
	t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
	t.Render()
	return nil
}
