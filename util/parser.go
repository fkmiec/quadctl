package util

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/fkmiec/quadctl/schema"
)

var (
	extensions = map[string]bool{
		".container": true,
		".pod":       true,
		".network":   true,
		".volume":    true,
	}
)

type Quadctl struct {
	QuadletSchemas map[string]map[string]schema.SchemaOption
	Config         map[string]string

	IsRootful         bool
	IsSystemd         bool
	IsPrintOnly       bool
	IsVerbose         bool
	IsFile            bool
	ListDepth         int
	Subcommand        string
	SearchDir         string
	PodmanArgs        string
	RunCmd            string
	DotQuadletsPath   string
	QuadletSrcPath    string // Path to the user's source directory containing quadlet folders or files
	UseSubdirectories bool   // Default to installing quadlets in a subdirectory to keep them organized
	UseSymbolicLinks  bool   // Default to copying files for installation to avoid potential issues with source files being moved or deleted, but can be configured to use symbolic links for a more dynamic setup
	IsReloadSystemd   bool   // Default to reloading systemd after installation to apply changes immediately
	IsRemoveVolumes   bool   // Default to removing volumes on uninstall since they are often not needed after uninstall and can be left behind if not removed, but can be configured to keep volumes for data persistence.
	IsRemoveNetworks  bool   // Default to removing networks on uninstall since they are often not needed after uninstall and can be left behind if not removed, but can be configured to keep volumes for data persistence.
	SystemdStartTmpl  *template.Template
	SystemdStopTmpl   *template.Template
	SystemdStatusTmpl *template.Template
	SystemdReloadTmpl *template.Template
	SystemdLogsTmpl   *template.Template
	QuadletRootPath   string
	QuadletUserPath   string
}

// Quadlet represents a parsed Quadlet file and its relationships.
type Quadlet struct {
	ID             string // Base name without extension (e.g., "my-app")
	Filepath       string
	Type           string // .container, .pod, .network, .volume
	Sections       map[string]map[string][]string
	Deps           []string          // IDs of other quadlets that must run first
	ParentPod      string            // If this is a container, the ID of its parent pod
	RestartPolicy  string            // [Service] Restart=
	GeneratedNames map[string]string // Key: name type, Value: specific name (useful for ps filters)
	ServiceName    string            // The name of the systemd unit (from quadlet file or default to <id>-<type>)
}

type Option struct {
	Key   string
	Value string
}

func InitQuadlets(quadctl *Quadctl) []*Quadlet {
	// Discover, parse and resolve dependencies
	quadlets, err := discoverAndParseQuadlets(quadctl, quadctl.SearchDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing quadlets in %s: %v\n", quadctl.SearchDir, err)
		os.Exit(1)
	}

	// If user specified the -f flag, the path provided should be a quadlet file, rather than directory. Only process the specified file and its dependencies.
	var selectedQuadlets []*Quadlet
	if quadctl.IsFile {
		// If a file was specified, find the corresponding quadlet
		tmp := strings.TrimSuffix(flag.Arg(1), filepath.Ext(flag.Arg(1)))
		selected := quadlets[tmp]
		if selected != nil {
			selectedQuadlets = append(selectedQuadlets, selected)
			if len(selected.Deps) > 0 {
				// Add dependencies to the selected quadlets
				for _, dep := range selected.Deps {
					if depQuadlet := quadlets[dep]; depQuadlet != nil {
						selectedQuadlets = append(selectedQuadlets, depQuadlet)
					}
				}
			}
			// Replace the original quadlets with the selected ones
			selectedQuadletsMap := make(map[string]*Quadlet)
			for _, q := range selectedQuadlets {
				selectedQuadletsMap[q.ID] = q
			}
			quadlets = selectedQuadletsMap
		}
	}

	//quadlets := util.InitQuadlets()

	// Topologically sort quadlets based on dependencies
	ordered, err := topologicalSort(quadlets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining ordering: %v\n", err)
		os.Exit(1)
	}

	return ordered
}

// --- PARSING AND GENERATION LOGIC ---

func discoverAndParseQuadlets(quadctl *Quadctl, searchDir string) (map[string]*Quadlet, error) {

	quadlets := make(map[string]*Quadlet)

	if info, err := os.Stat(searchDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("search path is not a directory: %s", searchDir)
	}

	dir, err := os.Open(searchDir)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	/*
	   Proposed modification to support single file format (.quadlet):
	   - Loop over searchDir and check for a .quadlets file extension (single file format for quadlets)
	   - If found one or more .quadlets file(s)
	   -   Create temp directory and extract .quadlets into separate files with their indicated filenames and extensions
	   -   Set DotQuadletsPath
	   - If DotQuadletsPath != "" (ie. there were .quadlets files):
	   -   Loop over files again and look for .container, .pod, .volume, .network and if found, copy to the temp directory
	   -   Also copy any drop in directories
	   -   Set searchDir = DotQuadletsPath (aka tempDir) and read files again from new searchDir

	   - Loop over files and look for .container, .pod, .volume, .network and if found, parse quadlets
	*/
	for _, f := range files {
		//fmt.Println(f.Name(), f.IsDir())
		path := filepath.Join(searchDir, f.Name())
		ext := filepath.Ext(path)
		if ".quadlets" == ext {
			//parseDotQuadlets extracts individual quadlets into separate files in a temp directory
			tempDir, err := parseDotQuadlets(path)
			if err != nil {
				return nil, err
			}

			//Save the DotQuadletsPath (location .quadlets file was extracted to) in case needed for systemd install
			quadctl.DotQuadletsPath = tempDir

			//tempQuadlets, err := discoverAndParseQuadlets(quadctl, tempDir)
			//if err != nil {
			//	return nil, err
			//}
			//for k, v := range tempQuadlets {
			//	quadlets[k] = v
			//}
		}
	}

	// If there were .quadlets files, then we copy other dot files to the temp directory where .quadlets were extracted
	if quadctl.DotQuadletsPath != "" {
		for _, f := range files {
			if f.IsDir() {
				path := filepath.Join(searchDir, f.Name())
				newPath := filepath.Join(quadctl.DotQuadletsPath, f.Name())
				if err := CopyDir(path, newPath); err != nil {
					fmt.Fprintf(os.Stderr, " Error copying drop-in directory %s to %s: %v\n", path, newPath, err)
					os.Exit(1)
				}
				continue
			}
			//fmt.Println(f.Name(), f.IsDir())
			path := filepath.Join(searchDir, f.Name())
			ext := filepath.Ext(path)
			if extensions[ext] {
				newPath := filepath.Join(quadctl.DotQuadletsPath, f.Name())
				if err := CopyFile(path, newPath); err != nil {
					fmt.Fprintf(os.Stderr, " Error copying %s to temporary .quadlets processing path %s: %v\n", path, newPath, err)
					os.Exit(1)
				}
			}
		}
		searchDir = quadctl.DotQuadletsPath
		dir, err = os.Open(searchDir)
		if err != nil {
			return nil, err
		}
		files, err = dir.Readdir(0)
		if err != nil {
			return nil, err
		}
	}

	// Below will process all .container, .pod, .volume, .network files
	// If there were .quadlets files, all were extracted to a temp directory and all other files and subdirectories were copied to the temp directory

	for _, f := range files {
		//fmt.Println(f.Name(), f.IsDir())
		path := filepath.Join(searchDir, f.Name())
		ext := filepath.Ext(path)
		if extensions[ext] {
			q, err := parseQuadlet(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, " Error parsing %s: %v\n", path, err)
				os.Exit(1)
			} else {
				quadlets[q.ID] = q
			}
		}
	}

	// 2nd pass: Extract dependencies (after all have IDs)
	for _, q := range quadlets {
		extractDependencies(q, quadlets)
	}

	return quadlets, nil
}

// Split quadlets by "---" on a separate new line and find filenames specified as "# FileName=<name>"
func parseDotQuadlets(path string) (string, error) {
	// Extract the .quadlets file into a temp directory with the same name as the original quadctl.SearchDir in the system temp directory.
	//base := filepath.Base(path)
	//id := strings.TrimSuffix(base, ".quadlets")

	id := filepath.Base(filepath.Dir(path))
	tempDir := filepath.Join(os.TempDir(), id)

	//fmt.Printf("Temp Dir for .quadlet: %s\n", tempDir)

	// Create temp directory
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp directory: %v\n", err)
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	baseQuadletFilename := ""
	quadletText := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		//fmt.Println("READING: " + line)
		if strings.HasPrefix(line, "#") && strings.Contains(strings.TrimSpace(line), "Filename") {
			//fmt.Println("Found Filename...")
			prop := strings.Split(line, "=")
			if len(prop) > 1 {
				baseQuadletFilename = strings.TrimSpace(prop[1])
				//fmt.Println("Filename: " + baseQuadletFilename)
				continue
			}
		}
		// Save file when hit the separator
		if "---" == strings.TrimSpace(line) {
			//fmt.Println("SAVING file...")

			err := WriteFile(filepath.Join(tempDir, baseQuadletFilename), quadletText)
			if err != nil {
				return "", err
			}
			baseQuadletFilename = ""
			quadletText = ""
			continue
		}
		quadletText += line + "\n"
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading .quadlets file: %v\n", err)
		os.Exit(1)
	}

	// Save file if reach end of .quadlet file with a filename and quadlet text
	if len(baseQuadletFilename) > 0 && len(quadletText) > 0 {
		//fmt.Println("SAVING FINAL FILE...")
		err := WriteFile(filepath.Join(tempDir, baseQuadletFilename), quadletText)
		if err != nil {
			return "", err
		}
	}

	return tempDir, nil
}

func parseQuadlet(path string) (*Quadlet, error) {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	id := strings.TrimSuffix(base, ext)

	q := &Quadlet{
		ID:             id,
		Filepath:       path,
		Type:           ext,
		Sections:       make(map[string]map[string][]string),
		GeneratedNames: make(map[string]string),
	}

	if err := parseIniFile(path, q); err != nil {
		return nil, err
	}

	// Set service name ... For container, use ServiceName if provided, otherwise {id}. For others, ServiceName or {id}-{type}
	var confServiceName string
	switch q.Type {
	case ".container":
		q.GeneratedNames["container"] = id
		vals := q.Sections["Container"]["ServiceName"]
		if len(vals) > 0 {
			confServiceName = vals[0]
		}
	case ".pod":
		vals := q.Sections["Pod"]["ServiceName"]
		if len(vals) > 0 {
			confServiceName = vals[0]
		}
	case ".volume":
		vals := q.Sections["Volume"]["ServiceName"]
		if len(vals) > 0 {
			confServiceName = vals[0]
		}
	case ".network":
		vals := q.Sections["Network"]["ServiceName"]
		if len(vals) > 0 {
			confServiceName = vals[0]
		}
	}
	if confServiceName == "" {
		if q.Type == ".container" {
			q.ServiceName = id
		} else {
			q.ServiceName = id + "-" + strings.TrimPrefix(q.Type, ".")
		}
	} else {
		q.ServiceName = confServiceName
	}

	// Merge systemd-style drop-ins from filename.d/*.conf
	dropInDir := path + ".d"
	if info, err := os.Stat(dropInDir); err == nil && info.IsDir() {
		files, _ := filepath.Glob(filepath.Join(dropInDir, "*.conf"))
		for _, f := range files {
			_ = parseIniFile(f, q) // Merge drop-ins silently
		}
	}

	// Specific checks based on parsing
	if contSec, ok := q.Sections["Container"]; ok {
		if val, ok := contSec["ContainerName"]; ok && len(val) > 0 {
			q.GeneratedNames["container"] = val[0]
		}
		if val, ok := contSec["Pod"]; ok && len(val) > 0 {
			q.ParentPod = strings.TrimSuffix(val[0], ".pod")
		}
		if val, ok := contSec["AutoUpdate"]; ok && len(val) > 0 {
			q.GeneratedNames["auto_update"] = val[0]
		}
	}

	if svcSec, ok := q.Sections["Service"]; ok {
		if val, ok := svcSec["Restart"]; ok && len(val) > 0 {
			q.RestartPolicy = strings.ToLower(val[0])
		}
	}

	return q, nil
}

// Simple INI parser
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

				//Handle options specified using a multiple space-separated value format
				values := ParseFields(val)
				for _, v := range values {
					q.Sections[currentSection][key] = append(q.Sections[currentSection][key], v)
				}

				//q.Sections[currentSection][key] = append(q.Sections[currentSection][key], val)
			}
		}
	}
	return scanner.Err()
}

// extractDependencies determines implicit and explicit requirements
func extractDependencies(q *Quadlet, all map[string]*Quadlet) {
	depSet := make(map[string]bool)

	// Explicit Systemd dependencies [Unit] After=/Requires=
	if unit, ok := q.Sections["Unit"]; ok {
		for _, key := range []string{"Requires", "After"} {
			for _, val := range unit[key] {
				// Strip systemd.service ext, and optional quadlet ext, map back to ID
				id := strings.TrimSuffix(val, ".service")
				id = strings.TrimSuffix(id, filepath.Ext(id))
				if _, exists := all[id]; exists {
					depSet[id] = true
				}
			}
		}
	}

	// Implicit dependencies [Container/Pod] Network=/Volume=/Pod=
	if q.Type == ".container" {
		cont := q.Sections["Container"]
		if pod, ok := cont["Pod"]; ok && len(pod) > 0 {
			depSet[strings.TrimSuffix(pod[0], ".pod")] = true
		}

		for _, net := range cont["Network"] {
			id := strings.TrimSuffix(net, ".network")
			if _, exists := all[id]; exists {
				depSet[id] = true
			}
		}

		for _, vol := range cont["Volume"] {
			// Vol format source.volume:/path
			sourceVol := strings.TrimSuffix(strings.Split(vol, ":")[0], ".volume")
			if _, exists := all[sourceVol]; exists {
				depSet[sourceVol] = true
			}
		}
	} else if q.Type == ".pod" {
		podSec := q.Sections["Pod"]
		for _, net := range podSec["Network"] {
			id := strings.TrimSuffix(net, ".network")
			if _, exists := all[id]; exists {
				depSet[id] = true
			}
		}
	}

	deps := []string{}
	for k := range depSet {
		deps = append(deps, k)
	}
	q.Deps = deps
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
			if _, exists := quadlets[dep]; !exists {
				return fmt.Errorf("%s depends on unknown quadlet %s", nodeID, dep)
			}
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

// parseFields splits a space-separated string into a slice,
// preserving spaces within quoted values.
func ParseFields(input string) []string {
	var fields []string
	if len(strings.TrimSpace(input)) == 0 {
		return fields
	}

	var currentToken strings.Builder
	inQuotes := false

	for _, r := range input {
		switch r {
		case '"':
			inQuotes = !inQuotes
			// We skip writing the quote character to the builder.
			// This automatically strips out the quotes while keeping the contents.
			// TEMPORARY - see if writing the quotes back is the way to go
			currentToken.WriteRune(r)
		case ' ':
			if inQuotes {
				currentToken.WriteRune(r)
			} else {
				// Space outside of quotes terminates the current key=value pair
				if currentToken.Len() > 0 {
					fields = append(fields, currentToken.String())
					currentToken.Reset()
				}
			}
		default:
			currentToken.WriteRune(r)
		}
	}

	// Catch the final pair if the string doesn't end with a trailing space
	if currentToken.Len() > 0 {
		fields = append(fields, currentToken.String())
	}

	/*
		//Quote any unquoted (previously quoted) strings with spaces
		for i, f := range fields {
			if strings.Contains(f, " ") {
				fields[i] = fmt.Sprintf("%q", f)
			}
		}
	*/

	return fields
}

func QuadletOptionToPodman(qType string, options map[string]schema.SchemaOption, k string, v string) (string, error) {
	var buf bytes.Buffer
	if opt, ok := options[k]; ok {
		option := Option{Key: opt.PodmanKey, Value: v}
		err := opt.PodmanTemplateParsed.Execute(&buf, option)
		if err != nil {
			return "", fmt.Errorf("Error formatting %s option %s: %w", qType, k, err)
		}
		return buf.String(), nil
	}
	return "", fmt.Errorf("Quadlet %s option not defined: %s", qType, k)
}
