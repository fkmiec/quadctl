package core

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/briandowns/spinner"
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
		cmd := exec.Command(c.Cmd[0], c.Cmd[1:]...)
		if slices.Contains(c.Cmd, "run") && (!slices.Contains(c.Cmd, "-d") || !slices.Contains(c.Cmd, "--detach")) {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Run()
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
