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
	fmt.Println("  ai -a <question>          Ask AI about your system (AI runs read-only commands to answer)")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println("  -d, --debug      Enable debug mode (output logs to stderr)")
	fmt.Println("  -c, --config     Run configuration wizard to set up API keys")
	fmt.Println("  -i, --info       Show current API configuration")
	fmt.Println("  -a, --ask        Ask AI about your system (assistant mode)")
fmt.Println("  --history        Show command history")
fmt.Println("\nHistory Shortcuts:")
fmt.Println("  !!               Repeat last command")
fmt.Println("  !<number>        Repeat command by ID (e.g., !12)")
fmt.Println("  !<keyword>       Search and repeat command (e.g., !memory)")
	fmt.Println("\nExamples:")
	fmt.Println("  ai list all files in current directory")
	fmt.Println("  ai -d explain current directory structure")
	fmt.Println("  ai show system all usb camera devices")
	fmt.Println("  ai find all log files in c:\\temp including subdirectories")
	fmt.Println("  ai -a 系统里有没有不认识的进程")
	fmt.Println("  ai -a what services are running on my system")
fmt.Println("  ai !!                            # Repeat last command")
fmt.Println("  ai !5                            # Repeat command #5")
fmt.Println("  ai '!docker'                     # Search and run docker command")
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

func showConfig(cfg *Config) {
	fmt.Println("Current Configuration:")
	fmt.Printf("  Provider:  %s\n", cfg.Provider)
	fmt.Printf("  API Key:   %s\n", maskKey(cfg.APIKey))
	fmt.Printf("  Endpoint:  %s\n", cfg.Endpoint)
	fmt.Printf("  Model:     %s\n", cfg.Model)
	if cfg.Secret != "" {
		fmt.Printf("  Secret:    %s\n", maskKey(cfg.Secret))
	}
	configPath, _ := getConfigPath()
	fmt.Printf("  Config:    %s\n", configPath)
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
	// Filter out flags
	args := []string{os.Args[0]}
	debugEnabled := false
	runConfig := false
	showInfo := false
	assistantMode := false
	showHistory := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-d", "--debug":
			debugEnabled = true
		case "-c", "--config":
			runConfig = true
		case "-i", "--info":
			showInfo = true
		case "-a", "--ask":
			assistantMode = true
		case "--history":
			showHistory = true
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

	cfg := LoadConfig()

	if showInfo {
		showConfig(cfg)
		os.Exit(0)
	}

	// Initialize history manager
	historyMgr, err := NewHistoryManager()
	if err != nil {
		LogError("Failed to initialize history: %v", err)
	}

	// Show history if requested
	if showHistory {
		if historyMgr != nil {
			historyMgr.PrintHistory()
		}
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

	naturalLang := strings.Join(os.Args[1:], " ")
	LogInfo("Input: %q", naturalLang)

	// Check for special history commands
	if historyMgr != nil {
		if resolved, ok := historyMgr.ResolveSpecialCommand(naturalLang); ok {
			naturalLang = resolved
			LogInfo("Resolved to: %q", naturalLang)
		}
	}

	if assistantMode {
		RunAssistantMode(cfg, naturalLang)
		return
	}

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
	fmt.Printf("\r> %s", colorize(command, danger))

	if danger != "green" {
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
	}

	exitCode := Execute(command)
	LogInfo("Execution finished | Command: %q | Exit code: %d", command, exitCode)

	// Save to history
	if historyMgr != nil {
		historyMgr.Add(naturalLang, command, exitCode == 0)
	}

	os.Exit(exitCode)
}
