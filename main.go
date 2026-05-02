package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func showHelp() {
	fmt.Println("AI Command - Natural language to shell command converter")
	fmt.Println("\nUsage:")
	fmt.Println("  ai <command>    Convert natural language to a shell command and execute it")
	fmt.Println("  ai -h, --help   Show this help message")
	fmt.Println("\nExamples:")
	fmt.Println("  ai list all files in current directory")
	fmt.Println("  ai show system all usb camera devices")
	fmt.Println("  ai find all log files in c:\\temp including subdirectories")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Environment Variables:")
	fmt.Println("    AI_CMD_API_KEY      Your LLM API key")
	fmt.Println("    AI_CMD_ENDPOINT     LLM API endpoint (default: OpenAI compatible)")
	fmt.Println("    AI_CMD_MODEL        LLM model name")
	fmt.Println("\n  Config File:")
	configPath, _ := getConfigPath()
	fmt.Printf("    %s\n", configPath)
}

func colorize(text, level string) string {
	reset := "\033[0m"
	color := ""
	switch level {
	case "green":
		color = "\033[32m" // Green
	case "yellow":
		color = "\033[33m" // Yellow
	case "orange":
		color = "\033[38;5;208m" // Orange (256-color)
	case "red":
		color = "\033[31;1m" // Bold Red
	default:
		return text
	}
	return color + text + reset
}

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		os.Exit(0)
	}

	cfg := LoadConfig()

	stopSpinner := make(chan bool)
	go func() {
		frames := []string{"|", "/", "-", "\\"}
		for i := 0; ; i++ {
			select {
			case <-stopSpinner:
				return
			default:
				fmt.Printf("\r%s", frames[i%len(frames)])
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	naturalLang := strings.Join(os.Args[1:], " ")
	command, danger, err := cfg.ConvertToCommand(naturalLang)
	stopSpinner <- true

	if err != nil {
		fmt.Printf("\rError: %v\n", err)
		os.Exit(1)
	}

	if strings.HasPrefix(command, "ERROR:") {
		fmt.Printf("\r%s\n", command)
		os.Exit(1)
	}

	if command == "" {
		fmt.Printf("\rError: empty command returned\n")
		os.Exit(1)
	}

	fmt.Printf("\r> %s (Enter or Ctrl + C)", colorize(command, danger))
	os.Exit(Execute(command))
}
