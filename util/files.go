package util

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func InitConfig(quadctl *Quadctl) {
	// Read config
	config, err := GetConfig(quadctl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}
	if val, ok := config["use_subdirectories"]; ok && (val == "false" || val == "0") {
		quadctl.UseSubdirectories = false
	}
	if val, ok := config["use_symbolic_links"]; ok && (val == "true" || val == "1") {
		quadctl.UseSymbolicLinks = true
	}
	if val, ok := config["auto_reload_systemd"]; ok && (val == "false" || val == "0") {
		quadctl.IsReloadSystemd = false
	}
	if val, ok := config["remove_volumes"]; ok && (val == "false" || val == "0") {
		quadctl.IsRemoveVolumes = false
	}
	if val, ok := config["remove_networks"]; ok && (val == "false" || val == "0") {
		quadctl.IsRemoveNetworks = false
	}
	if val, ok := config["quadlet.src.path"]; ok && val != "" {
		quadctl.QuadletSrcPath = val
	}
	if val, ok := config["quadlet.root.path"]; ok && val != "" {
		quadctl.QuadletRootPath = val
	}
	if val, ok := config["quadlet.user.path"]; ok && val != "" {
		quadctl.QuadletUserPath = val
	}
	if val, ok := config["systemd.start"]; ok && val != "" {
		quadctl.SystemdStartTmpl = template.Must(template.New("systemdStart").Parse(val))
	}
	if val, ok := config["systemd.stop"]; ok && val != "" {
		quadctl.SystemdStopTmpl = template.Must(template.New("systemdStop").Parse(val))
	}
	if val, ok := config["systemd.status"]; ok && val != "" {
		quadctl.SystemdStatusTmpl = template.Must(template.New("systemdStatus").Parse(val))
	}
	if val, ok := config["systemd.reload"]; ok && val != "" {
		quadctl.SystemdReloadTmpl = template.Must(template.New("systemdReload").Parse(val))
	}
	if val, ok := config["systemd.logs"]; ok && val != "" {
		quadctl.SystemdLogsTmpl = template.Must(template.New("systemdLogs").Parse(val))
	}
}

func GetConfig(quadctl *Quadctl) (map[string]string, error) {

	config := make(map[string]string)
	var path string
	if quadctl.IsRootful {
		path = os.Getenv("QUADCTL_CONFIG_DIR")
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			err = fmt.Errorf("Invalid config path: %s\nWhen running as root, ensure QUADCTL_CONFIG_DIR is set and points to a valid directory.\nTo set root config same as user:\n\necho \"QUADCTL_CONFIG_DIR=$HOME/.config/quadctl\" | sudo tee -a /etc/environment > /dev/null", path)
			return nil, err
		}
	} else {
		path = os.Getenv("XDG_CONFIG_HOME")
		if path == "" {
			path = os.Getenv("HOME") + "/.config"
		}
		path = filepath.Join(path, "quadctl")
	}

	path = filepath.Join(path, "quadctl.conf")

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			config[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}
	return config, nil
}

func WriteFile(path string, text string) error {
	//fmt.Println("WRITING: \n" + text)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}

func CopyFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := os.Chmod(dst, 0644); err != nil {
		d.Close()
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

func CopyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() {
			continue // Don't handle recursive dirs for drop-ins
		}
		if err := CopyFile(filepath.Join(src, f.Name()), filepath.Join(dst, f.Name())); err != nil {
			return err
		}
	}
	return nil
}
