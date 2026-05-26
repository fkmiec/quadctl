package core

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fkmiec/quadctl/util"
)

type Command struct {
	Label    string
	PreFn    func(*Command)
	RunFn    func(*Command)
	PostFn   func(*Command)
	Spinner  *spinner.Spinner
	Cmd      []string
	Output   []string
	Error    error
	Warnings []string
}

func (c *Command) PreCmd() {
	c.PreFn(c)
}

func (c *Command) RunCmd() {
	c.RunFn(c)
}

func (c *Command) PostCmd() {
	c.PostFn(c)
}

func NewCommand(label string) Command {
	return Command{
		Label:  label,
		PreFn:  DefaultPreFn,
		RunFn:  DefaultRunFn,
		PostFn: DefaultPostFn,
	}
}

func DefaultPreFn(c *Command) {
	if slices.Contains(c.Cmd, "run") && (!slices.Contains(c.Cmd, "-d") || !slices.Contains(c.Cmd, "--detach")) {
		return // Skip spinner for 'run' command since it is interactive and the spinner output can interfere with the container's output.
	}
	c.Spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond) // Build our new spinner
	c.Spinner.Prefix = c.Label + " "
	c.Spinner.Start() // Start the spinner
}

func DefaultRunFn(c *Command) {
	if len(c.Cmd) > 0 {

		// Remove quotes added due to spaces. exec.Command will escape the arg ... but dry run will need it quoted for use in shell
		for i, arg := range c.Cmd {
			if strings.Contains(arg, "\"") {
				//Debug
				//fmt.Printf("Found arg with quotes: %s\n", arg)
				c.Cmd[i] = strings.ReplaceAll(arg, "\"", "") //fmt.Sprintf("%q", arg)
			}
		}

		cmd := exec.Command(c.Cmd[0], c.Cmd[1:]...)
		if slices.Contains(c.Cmd, "run") && (!slices.Contains(c.Cmd, "-d") || !slices.Contains(c.Cmd, "--detach")) {
			fmt.Printf("Running in foreground: %s\n", strings.Join(c.Cmd, " "))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			c.Error = cmd.Run()
		} else {
			output, err := cmd.CombinedOutput()
			c.Output = []string{string(output)}
			c.Error = err

		}
	}
}

func DefaultPostFn(c *Command) {
	if slices.Contains(c.Cmd, "run") && (!slices.Contains(c.Cmd, "-d") || !slices.Contains(c.Cmd, "--detach")) {
		return // Skip stopping the spinner for 'run' command since it is interactive and the spinner output can interfere with the container's output.
	}
	c.Spinner.FinalMSG = fmt.Sprintf("%s... Done\n", c.Label)
	c.Spinner.Stop()
}

// Common handling for dry run / verbose output and command execution for all handlers that generate commands.
func RunCommands(quadctl *util.Quadctl, commands []Command) {

	if quadctl.IsVerbose {
		isHeaderPrinted := false
		for _, c := range commands {
			if len(c.Warnings) > 0 {
				if !isHeaderPrinted {
					fmt.Printf("\n# --- WARNINGS ---\n\n")
					isHeaderPrinted = true
				}
				for _, w := range c.Warnings {
					fmt.Printf("[WARN] %s\n", w)
				}
			}
		}
	}
	if quadctl.IsPrintOnly && len(commands) > 0 {
		fmt.Printf("\n# --- DRY-RUN MODE: Commands that would be executed ---\n\n")
		for _, c := range commands {
			if len(c.Cmd) > 0 {
				fmt.Printf("  %s\n", strings.Join(c.Cmd, " "))
			} else {
				fmt.Printf("  %s\n", c.Label)
				for _, line := range c.Output {
					fmt.Println("   => " + line)
				}
			}
		}
	} else if len(commands) > 0 {
		for _, c := range commands {
			c.PreCmd()
			c.RunCmd()
			c.PostCmd()

			if c.Output != nil {
				for _, o := range c.Output {
					fmt.Fprintf(os.Stdout, "%s\n", o)
				}
			}
			if c.Error != nil {
				fmt.Fprintf(os.Stderr, "Error executing command:\n\n  %s\n\n%s\n", strings.Join(c.Cmd, " "), c.Error.Error())
			}
		}
	}
}

func runCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func runCommandSilently(args []string) error {
	//if isRootful && args[0] != "sudo" {
	//	args = append([]string{"sudo"}, args...)
	//}
	cmd := exec.Command(args[0], args[1:]...)
	// Discard output
	err := cmd.Run()
	return err
}

func runCommandCapture(args []string) (string, error) {
	//if isRootful && args[0] != "sudo" {
	//	args = append([]string{"sudo"}, args...)
	//}

	//fmt.Printf("=> Running command: %s\n", strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.Output()
	return string(output), err
}
