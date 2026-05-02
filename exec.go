package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func Execute(command string) int {
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

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
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		return 1
	}
	return 0
}
