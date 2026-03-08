package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Quadlet represents a parsed Quadlet file and its drop-ins.
type Quadlet struct {
	ID       string
	Filepath string
	Type     string
	Sections map[string]map[string][]string
	Deps     []string
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	action := strings.ToLower(os.Args[1])
	fileArgs := os.Args[2:]

	if action != "up" && action != "down" && action != "print" {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", action)
		printUsage()
		os.Exit(1)
	}

	quadlets := make(map[string]*Quadlet)

	// 1. Parse files and their drop-ins
	for _, arg := range fileArgs {
		q, err := parseQuadlet(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", arg, err)
			continue
		}
		quadlets[q.ID] = q
	}

	// 2. Extract dependencies
	for _, q := range quadlets {
		q.Deps = extractDependencies(q, quadlets)
	}

	// 3. Topological Sort
	ordered, err := topologicalSort(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining startup order: %v\n", err)
		os.Exit(1)
	}

	// 4. Execute based on subcommand
	switch action {
	case "print":
		handlePrint(ordered)
	case "up":
		handleUp(ordered)
	case "down":
		handleDown(ordered)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <up|down|print> <file1.container> [file2.pod] ...\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  up    : Create and start all services\n")
	fmt.Fprintf(os.Stderr, "  down  : Stop and remove all services\n")
	fmt.Fprintf(os.Stderr, "  print : Print the podman commands and warnings without executing\n")
}

func handlePrint(ordered []*Quadlet) {
	fmt.Println("# --- Podman CLI Commands ---")
	for _, q := range ordered {
		cmd, warnings := generatePodmanCommand(q)
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "[WARN] %s: %s\n", q.Filepath, w)
		}
		if len(cmd) > 0 {
			fmt.Println(strings.Join(cmd, " "))
		}
	}
}

func handleUp(ordered []*Quadlet) {
	for _, q := range ordered {
		cmdArgs, warnings := generatePodmanCommand(q)
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "[WARN] %s: %s\n", q.Filepath, w)
		}

		if len(cmdArgs) > 0 {
			fmt.Printf("=> Executing: %s\n", strings.Join(cmdArgs, " "))
			runCommand(cmdArgs)
		}
	}
}

func handleDown(ordered []*Quadlet) {
	// Reverse the order for teardown
	for i := len(ordered) - 1; i >= 0; i-- {
		q := ordered[i]
		cmds := generateTeardownCommands(q)
		for _, cmdArgs := range cmds {
			fmt.Printf("=> Executing: %s\n", strings.Join(cmdArgs, " "))
			runCommand(cmdArgs)
		}
	}
}

func runCommand(args []string) {
	if len(args) == 0 {
		return
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Command failed: %v\n", err)
	}
}

func parseQuadlet(path string) (*Quadlet, error) {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	id := strings.TrimSuffix(base, ext)

	q := &Quadlet{
		ID:       id,
		Filepath: path,
		Type:     ext,
		Sections: make(map[string]map[string][]string),
	}

	// Parse main file
	if err := parseIniFile(path, q); err != nil {
		return nil, err
	}

	// Look for drop-in directory (e.g., app.container.d/)
	dropInDir := path + ".d"
	if info, err := os.Stat(dropInDir); err == nil && info.IsDir() {
		files, _ := filepath.Glob(filepath.Join(dropInDir, "*.conf"))
		for _, f := range files {
			_ = parseIniFile(f, q) // Merge drop-ins silently
		}
	}

	return q, nil
}

func parseIniFile(path string, q *Quadlet) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.Trim(line, "[]")
			if _, exists := q.Sections[currentSection]; !exists {
				q.Sections[currentSection] = make(map[string][]string)
			}
			continue
		}

		if currentSection != "" {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				q.Sections[currentSection][key] = append(q.Sections[currentSection][key], val)
			}
		}
	}
	return scanner.Err()
}

func extractDependencies(q *Quadlet, all map[string]*Quadlet) []string {
	depSet := make(map[string]bool)

	if unit, ok := q.Sections["Unit"]; ok {
		for _, key := range []string{"Requires", "After", "Wants"} {
			for _, val := range unit[key] {
				id := strings.TrimSuffix(strings.TrimSuffix(val, ".service"), filepath.Ext(val))
				if _, exists := all[id]; exists {
					depSet[id] = true
				}
			}
		}
	}

	if q.Type == ".container" {
		cont := q.Sections["Container"]
		for _, net := range cont["Network"] {
			id := strings.TrimSuffix(net, ".network")
			if _, exists := all[id]; exists {
				depSet[id] = true
			}
		}
		for _, pod := range cont["Pod"] {
			id := strings.TrimSuffix(pod, ".pod")
			if _, exists := all[id]; exists {
				depSet[id] = true
			}
		}
		for _, vol := range cont["Volume"] {
			id := strings.TrimSuffix(strings.Split(vol, ":")[0], ".volume")
			if _, exists := all[id]; exists {
				depSet[id] = true
			}
		}
	}

	var deps []string
	for k := range depSet {
		deps = append(deps, k)
	}
	return deps
}

func topologicalSort(quadlets map[string]*Quadlet) ([]*Quadlet, error) {
	var ordered []*Quadlet
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(nodeID string) error
	visit = func(nodeID string) error {
		if temp[nodeID] {
			return fmt.Errorf("circular dependency detected involving %s", nodeID)
		}
		if visited[nodeID] {
			return nil
		}

		temp[nodeID] = true
		for _, dep := range quadlets[nodeID].Deps {
			if err := visit(dep); err != nil {
				return err
			}
		}
		temp[nodeID] = false
		visited[nodeID] = true
		ordered = append(ordered, quadlets[nodeID])
		return nil
	}

	for id := range quadlets {
		if !visited[id] {
			if err := visit(id); err != nil {
				return nil, err
			}
		}
	}
	return ordered, nil
}

func generatePodmanCommand(q *Quadlet) ([]string, []string) {
	var warnings []string
	var cmd []string

	for sec := range q.Sections {
		if sec == "Install" || sec == "Service" {
			warnings = append(warnings, fmt.Sprintf("Ignoring [%s] section (Systemd specific)", sec))
		}
	}

	// Helper to extract PodmanArgs securely
	getPodmanArgs := func(section map[string][]string) []string {
		var args []string
		for _, argStr := range section["PodmanArgs"] {
			args = append(args, strings.Fields(argStr)...) // Simple space splitting
		}
		return args
	}

	switch q.Type {
	case ".volume":
		cmd = append(cmd, "podman", "volume", "create")
		if volSec, ok := q.Sections["Volume"]; ok {
			cmd = append(cmd, getPodmanArgs(volSec)...)
			for k := range volSec {
				if k != "PodmanArgs" {
					warnings = append(warnings, fmt.Sprintf("Ignored [Volume] key: %s", k))
				}
			}
		}
		cmd = append(cmd, q.ID)

	case ".network":
		cmd = append(cmd, "podman", "network", "create")
		if netSec, ok := q.Sections["Network"]; ok {
			cmd = append(cmd, getPodmanArgs(netSec)...)
			for k, vals := range netSec {
				for _, v := range vals {
					switch k {
					case "Subnet":
						cmd = append(cmd, "--subnet", v)
					case "Gateway":
						cmd = append(cmd, "--gateway", v)
					case "Label":
						cmd = append(cmd, "--label", v)
					case "PodmanArgs":
						// Handled above
					default:
						warnings = append(warnings, fmt.Sprintf("Ignored [Network] key: %s", k))
					}
				}
			}
		}
		cmd = append(cmd, q.ID)

	case ".pod":
		cmd = append(cmd, "podman", "pod", "create", "--name", q.ID)
		if podSec, ok := q.Sections["Pod"]; ok {
			cmd = append(cmd, getPodmanArgs(podSec)...)
			for k, vals := range podSec {
				for _, v := range vals {
					switch k {
					case "PublishPort":
						cmd = append(cmd, "-p", v)
					case "Network":
						cmd = append(cmd, "--network", strings.TrimSuffix(v, ".network"))
					case "PodmanArgs":
						// Handled above
					default:
						warnings = append(warnings, fmt.Sprintf("Ignored [Pod] key: %s", k))
					}
				}
			}
		}

	case ".container":
		cmd = append(cmd, "podman", "run", "-d", "--name", q.ID)
		var image string
		if contSec, ok := q.Sections["Container"]; ok {
			cmd = append(cmd, getPodmanArgs(contSec)...)
			for k, vals := range contSec {
				for _, v := range vals {
					switch k {
					case "Image":
						image = v
					case "Environment":
						cmd = append(cmd, "-e", v)
					case "PublishPort":
						cmd = append(cmd, "-p", v)
					case "Volume":
						volName := strings.Split(v, ":")[0]
						cleanVol := strings.TrimSuffix(volName, ".volume")
						mapped := strings.Replace(v, volName, cleanVol, 1)
						cmd = append(cmd, "-v", mapped)
					case "Network":
						cmd = append(cmd, "--network", strings.TrimSuffix(v, ".network"))
					case "Pod":
						cmd = append(cmd, "--pod", strings.TrimSuffix(v, ".pod"))
					case "ContainerName":
						for i, part := range cmd {
							if part == "--name" {
								cmd[i+1] = v
							}
						}
					case "PodmanArgs":
						// Handled above
					default:
						warnings = append(warnings, fmt.Sprintf("Ignored [Container] key: %s", k))
					}
				}
			}
		}
		if image == "" {
			warnings = append(warnings, "No Image= specified in [Container] section")
			image = "<MISSING_IMAGE>"
		}
		cmd = append(cmd, image)
	}

	return cmd, warnings
}

func generateTeardownCommands(q *Quadlet) [][]string {
	var cmds [][]string
	switch q.Type {
	case ".container":
		cmds = append(cmds, []string{"podman", "stop", q.ID})
		cmds = append(cmds, []string{"podman", "rm", "-f", q.ID})
	case ".pod":
		cmds = append(cmds, []string{"podman", "pod", "stop", q.ID})
		cmds = append(cmds, []string{"podman", "pod", "rm", "-f", q.ID})
	case ".network":
		cmds = append(cmds, []string{"podman", "network", "rm", q.ID})
	case ".volume":
		cmds = append(cmds, []string{"podman", "volume", "rm", q.ID})
	}
	return cmds
}
