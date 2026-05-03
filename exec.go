package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Execute(command string) int {
	LogInfo("Executing: %q", command)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			LogError("Command failed with exit code %d: %q", exitErr.ExitCode(), command)
			return exitErr.ExitCode()
		}
		LogError("Execution error: %v | Command: %q", err, command)
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		return 1
	}

	LogInfo("Command executed successfully: %q", command)
	return 0
}
