package main

import (
	"errors"
	"os"
	"os/exec"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return -1
	}

	args := []string{}
	if len(cmd) > 1 {
		args = cmd[1:]
	}

	for name, value := range env {
		if err := os.Unsetenv(name); err != nil {
			return -1
		}
		if !value.NeedRemove {
			if err := os.Setenv(name, value.Value); err != nil {
				return -1
			}
		}
	}

	command := exec.Command(cmd[0], args...) // #nosec G204
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode()
		}
		return -1
	}
	return 0
}
