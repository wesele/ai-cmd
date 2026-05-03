package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func showHelp() {
	fmt.Println("AI Command - Natural language to shell command converter")
	fmt.Println("\nUsage:")
	fmt.Println("  ai [options] <command>    Convert natural language to a shell command and execute it")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -d, --debug    Enable debug mode (output logs to stderr)")
	fmt.Println("  -c, --config   Run configuration wizard to set up API keys")
	fmt.Println("\nExamples:")
	fmt.Println("  ai list all files in current directory")
	fmt.Println("  ai -d explain current directory structure")
	fmt.Println("  ai show system all usb camera devices")
	fmt.Println("  ai find all log files in c:\\temp including subdirectories")
	fmt.Println("\nConfiguration:")
	fmt.Println("  Environment Variables:")
	fmt.Println("    AI_CMD_API_KEY      Your LLM API key")
	fmt.Println("    AI_CMD_ENDPOINT     LLM API endpoint (default: OpenAI compatible)")
	fmt.Println("    AI_CMD_MODEL        LLM model name")
	fmt.Println("    AI_CMD_PROVIDER     AI provider (openai, baidu)")
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
	// Filter out -d/--debug and -c/--config flags
	args := []string{os.Args[0]}
	debugEnabled := false
	runConfig := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-d", "--debug":
			debugEnabled = true
		case "-c", "--config":
			runConfig = true
		default:
			args = append(args, arg)
		}
	}
	os.Args = args

	if debugEnabled {
		EnableDebug()
		LogInfo("Debug mode enabled")
	}

	// Run config wizard if requested
	if runConfig {
		RunConfigWizard()
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		showHelp()
		os.Exit(0)
	}

	cfg := LoadConfig()

	naturalLang := strings.Join(os.Args[1:], " ")
	LogInfo("Input: %q", naturalLang)

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

	start := time.Now()
	command, danger, err := cfg.ConvertToCommand(naturalLang)
	elapsed := time.Since(start)
	stopSpinner <- true

	if err != nil {
		LogError("API error: %v | Input: %q", err, naturalLang)
		fmt.Printf("\rError: %v\n", err)
		os.Exit(1)
	}

	if strings.HasPrefix(command, "ERROR:") {
		LogError("LLM returned error: %s | Input: %q", command, naturalLang)
		fmt.Printf("\r%s\n", command)
		os.Exit(1)
	}

	if command == "" {
		LogError("Empty command returned | Input: %q", naturalLang)
		fmt.Printf("\rError: empty command returned\n")
		os.Exit(1)
	}

	LogCommand(naturalLang, command, danger, elapsed)
	fmt.Printf("\n> %s\n(Press Enter to execute or Ctrl+C to cancel)\n", colorize(command, danger))

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')

	exitCode := Execute(command)
	LogInfo("Execution finished | Command: %q | Exit code: %d", command, exitCode)
	os.Exit(exitCode)
}
