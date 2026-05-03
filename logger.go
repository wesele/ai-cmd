package main

import (
	"fmt"
	"os"
)

var debugMode bool

func EnableDebug() {
	debugMode = true
}

func LogInfo(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", args...)
	}
}

func LogError(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
	}
}

func LogCommand(input, command, danger string, duration interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[INFO] Input: %s | Command: %s | Danger: %s\n", input, command, danger)
	}
}
