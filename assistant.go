package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var readOnlyCommands = map[string]bool{
	"Get-Process":         true,
	"Get-Service":         true,
	"Get-EventLog":        true,
	"Get-WinEvent":        true,
	"Get-Content":         true,
	"Select-String":       true,
	"Get-ChildItem":       true,
	"Get-Item":            true,
	"Get-ItemProperty":    true,
	"Get-WmiObject":       true,
	"Get-CimInstance":     true,
	"Get-ComputerInfo":    true,
	"Get-HotFix":          true,
	"Get-Volume":          true,
	"Get-Disk":            true,
	"Get-Partition":       true,
	"Get-ScheduledTask":   true,
	"Get-AppxPackage":     true,
	"Get-Package":         true,
	"Get-Module":          true,
	"Get-Command":         true,
	"Get-Help":            true,
	"Get-Date":            true,
	"Test-Path":           true,
	"Get-Location":        true,
	"Get-PSDrive":         true,
	"Get-NetAdapter":      true,
	"Get-NetIPAddress":    true,
	"Get-NetConnectionProfile": true,
	"Get-LocalUser":       true,
	"Get-LocalGroup":      true,
	"Get-NetTCPConnection": true,
	"Get-NetRoute":        true,
	"Get-NetFirewallRule": true,
	"Measure-Object":      true,
	"Sort-Object":         true,
	"Group-Object":        true,
	"Where-Object":        true,
	"Select-Object":       true,
	"Format-List":         true,
	"Format-Table":        true,
	"Get-Counter":         true,
	"Get-FileHash":        true,
	"Get-AuthenticodeSignature": true,
	"Test-Connection":     true,
	"Test-NetConnection":  true,
	"Resolve-DnsName":     true,
	"dir":                 true,
	"ls":                  true,
	"cat":                 true,
	"type":                true,
	"find":                true,
	"findstr":             true,
	"ping":                true,
	"tracert":             true,
	"nslookup":            true,
	"ipconfig":            true,
	"systeminfo":          true,
	"tasklist":            true,
	"netstat":             true,
	"whoami":              true,
	"hostname":            true,
	"date":                true,
	"time":                true,
	"echo":                true,
	"query":               true,
	"driverquery":         true,
	"wmic":                true,
	"reg":                 true,
	"net":                 true,
	"schtasks":            true,
	"sc":                  true,
	"curl":                true,
	"wget":                true,
	"more":                true,
	"less":                true,
	"head":                true,
	"tail":                true,
	"wc":                  true,
	"grep":                true,
	"awk":                 true,
	"sed":                 true,
	"ps":                  true,
	"top":                 true,
	"free":                true,
	"df":                  true,
	"du":                  true,
	"uname":               true,
	"uptime":              true,
	"ifconfig":            true,
	"ss":                  true,
	"lsof":                true,
	"lsblk":               true,
	"lscpu":               true,
	"lsmem":               true,
	"lsusb":               true,
	"lspci":               true,
	"powercfg":            true,
	"Get-ComputerRestorePoint": true,
	"Get-PSRepository":    true,
	"Get-ExecutionPolicy": true,
}

var dangerousPatterns = []string{
	"Remove-Item", "rm -rf", "rm -r", "rm\t-rf", "rm\t-r", "del /s", "del /f", "rmdir /s", "rd /s",
	"Set-Content", "Add-Content", "Clear-Content",
	"format", "diskpart", "chkdsk /f", "chkdsk /r",
	"reg add", "reg delete", "reg copy", "reg restore", "reg load", "reg save",
	"sc create", "sc delete", "sc config",
	"net user /add", "net user /delete", "net localgroup /add", "net localgroup /delete",
	"net share /delete", "net session /delete",
	"schtasks /create", "schtasks /delete", "schtasks /change",
	"shutdown", "restart", "logoff",
	"taskkill", "tskill", "Stop-Process", "kill ",
	"attrib +", "icacls /grant", "icacls /deny", "takeown",
	"bcdedit /set", "bcdedit /delete",
	"fsutil", "cipher /w",
	">>", "| Out-File", "> $env:TEMP", "> /tmp/",
	"Invoke-Expression", "Invoke-WebRequest", "Invoke-RestMethod",
	"Start-BitsTransfer",
	"curl -X POST", "curl -X PUT", "curl -X DELETE", "curl -d",
	"mkfs", "fdisk", "parted", "mount -o rw", "umount",
	"iptables -A", "iptables -F", "firewall-cmd --permanent",
	"crontab -e", "crontab -r",
	"useradd", "userdel", "usermod", "groupadd", "groupdel",
	"passwd", "visudo",
	"systemctl start", "systemctl stop", "systemctl restart", "systemctl enable", "systemctl disable",
	"service start", "service stop", "service restart",
	"apt-get install", "apt-get remove", "apt-get purge", "apt-get upgrade",
	"yum install", "yum remove", "yum update",
	"pip install", "npm install", "go get",
	"dd ", "wipe", "shred",
	"Start-Service", "Stop-Service", "Restart-Service",
	"Set-ExecutionPolicy", "Set-ItemProperty",
	"New-Item -ItemType Directory", "mkdir -p",
}

func isDangerousCommand(cmd string) bool {
	cmdTrimmed := strings.TrimSpace(cmd)
	if cmdTrimmed == "" {
		return false
	}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(cmdTrimmed), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func isReadOnlyCommand(cmd string) bool {
	cmdTrimmed := strings.TrimSpace(cmd)
	if cmdTrimmed == "" {
		return false
	}
	firstWord := cmdTrimmed
	if idx := strings.IndexAny(cmdTrimmed, " \t"); idx != -1 {
		firstWord = cmdTrimmed[:idx]
	}
	if _, ok := readOnlyCommands[firstWord]; ok {
		return true
	}
	if _, ok := readOnlyCommands[strings.ToLower(firstWord)]; ok {
		return true
	}
	return false
}

func truncatePurpose(purpose string) string {
	runes := []rune(purpose)
	if len(runes) > 10 {
		return string(runes[:10]) + "..."
	}
	return purpose
}

func stripMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inDetails := false
	for _, line := range lines {
		if strings.HasPrefix(line, "<details") {
			inDetails = true
			continue
		}
		if inDetails {
			if strings.HasPrefix(line, "</details>") {
				inDetails = false
			}
			continue
		}
		line = strings.TrimPrefix(line, "```")
		line = strings.TrimPrefix(line, "###")
		line = strings.TrimPrefix(line, "##")
		line = strings.TrimPrefix(line, "#")
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "`", "")
		if strings.HasPrefix(line, "Response ID:") || strings.HasPrefix(line, "Request ID:") {
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			line = "- " + strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
		}
		result = append(result, line)
	}
	return strings.TrimSpace(strings.Join(result, "\n"))
}

func executeCommand(command, purpose string) (string, error) {
	readOnly := isReadOnlyCommand(command)
	dangerous := isDangerousCommand(command)

	if !readOnly {
		if dangerous {
			fmt.Printf("\r\033[31m[危险命令] %s\033[0m\n", command)
		} else {
			fmt.Printf("\r\033[33m[需确认] %s\033[0m\n", command)
		}
		fmt.Print("  (Press Enter to execute or Ctrl+C to cancel) ")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n[stderr] " + stderr.String()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("command exited with code %d", exitErr.ExitCode())
		}
		return output, fmt.Errorf("execution error: %v", err)
	}

	return output, nil
}

func RunAssistantMode(cfg *Config, question string) {
	provider := GetProvider(cfg)
	if provider == nil {
		fmt.Fprintf(os.Stderr, "Error: no provider available\n")
		os.Exit(1)
	}

	_, ok := provider.(*OpenAIProvider)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: assistant mode (-a) only supports OpenAI-compatible providers\n")
		os.Exit(1)
	}

	systemPrompt := fmt.Sprintf(`You are a system analysis assistant. The user will ask questions about their system.

You have access to a tool called "execute_command" that can run system commands to gather information and perform safe, non-destructive operations.

CRITICAL LANGUAGE RULE:
- You MUST respond in the EXACT SAME LANGUAGE as the user's question.
- If the user asks in Chinese, your ENTIRE response (including all analysis, recommendations, and explanations) MUST be in Chinese.
- If the user asks in English, respond entirely in English.
- NEVER mix languages in your final answer.

Rules:
- Use the execute_command tool to gather information needed to answer the user's question
- You may call the tool multiple times if needed
- For each tool call, provide a clear "purpose" explaining why you're running the command
- Allowed commands include:
  - Read-only/query commands (e.g., Get-Process, tasklist, netstat, systeminfo, ipconfig, dir, ps, top)
  - Diagnostic/test commands (e.g., ping, Test-Connection, Test-NetConnection, nslookup)
  - Safe informational commands that do not modify system state
- NEVER execute commands that delete, destroy, or irreversibly modify data
- NEVER execute commands that change system configuration, services, users, or permissions
- NEVER execute commands that install/uninstall software
- NEVER execute commands that modify network firewall rules
- If a command fails or returns an error, try a different approach or alternative command — do NOT give up
- If a command is blocked as dangerous, find a safer alternative that achieves the same goal
- After gathering information, you MUST analyze the results and provide a clear answer
- You SHOULD provide analysis, recommendations, and optimization advice based on the data — giving text advice is NOT modifying the system
- If the user asks for suggestions or advice, provide them based on your analysis of the collected data
- If you can answer without running commands, do so
- The user's operating system is: %s

Output format rules:
- NEVER use markdown formatting (no **bold**, no *italic*, no headers, no code blocks, no bullet lists with *)
- Use plain text only
- Use simple dashes (-) for list items, not asterisks
- Use numbers for ordered lists
- Use plain text tables with spaces/tabs for alignment
- Keep formatting simple and terminal-friendly

When you have enough information, provide your final answer in a clear, organized format.`, getOSHint())

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	maxRounds := 10
	for round := 0; round < maxRounds; round++ {
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

		resp, err := provider.ChatWithTools(cfg, messages)
		stopSpinner <- true

		if err != nil {
			fmt.Printf("\rError: %v\n", err)
			os.Exit(1)
		}

		msg := resp.Choices[0].Message

		if len(msg.ToolCalls) > 0 {
			fmt.Printf("\r")

			assistantMsg := Message{
				Role:      "assistant",
				Content:   "",
				ToolCalls: msg.ToolCalls,
			}
			messages = append(messages, assistantMsg)

			gray := "\033[90m"
			reset := "\033[0m"

			for _, tc := range msg.ToolCalls {
				var args ToolCallArgs
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					toolResult := Message{
						Role:       "tool",
						Content:    fmt.Sprintf("Error: failed to parse arguments. Please fix the JSON format and try again. Original: %s", tc.Function.Arguments),
						ToolCallID: tc.ID,
					}
					messages = append(messages, toolResult)
					continue
				}

				purpose := truncatePurpose(args.Purpose)
				readOnly := isReadOnlyCommand(args.Command)
				if readOnly {
					fmt.Printf("\r%s[%s] - %s%s\n", gray, purpose, args.Command, reset)
				}

				result, execErr := executeCommand(args.Command, args.Purpose)

				content := result
				if execErr != nil {
					if strings.Contains(execErr.Error(), "blocked") {
						content = fmt.Sprintf("BLOCKED: This command was rejected by the safety filter. You must use a safer alternative command. Original command: %s", args.Command)
					} else {
						content = fmt.Sprintf("FAILED: %v\nOutput so far:\n%s\nPlease try a different command or adjust your approach.", execErr, result)
					}
				}

				toolResult := Message{
					Role:       "tool",
					Content:    content,
					ToolCallID: tc.ID,
				}
				messages = append(messages, toolResult)
			}

			continue
		}

		if msg.Content != "" {
			fmt.Printf("\n%s\n\n", stripMarkdown(msg.Content))
		}

		return
	}

	fmt.Println("\n[提示] 已达到最大对话轮次限制")
}
