# AI Command (ai-cmd)

[![Releases](https://img.shields.io/github/v/release/wesele/ai-cmd?include_prereleases)](https://github.com/wesele/ai-cmd/releases)
[![Download](https://img.shields.io/badge/download-latest_binaries-blue)](https://github.com/wesele/ai-cmd/releases/latest)

A lightweight CLI tool that converts natural language instructions into executable system commands (PowerShell/Bash) using Large Language Models.

[English](#english) | [中文](#中文)

---

<a name="english"></a>
## English

### Features

- **Natural Language Processing**: Convert phrases like "list all log files in temp folder" into precise shell commands.
- **Assistant Mode (`-a`)**: Ask AI questions about your system. The AI autonomously runs commands to gather information and provides a clear answer — with full transparency on every command executed and its purpose. Tool call output is displayed in gray to distinguish it from the final answer.
  - **Three-tier Safety Mechanism**:
    - 🟢 **Read-only**: Auto-executed (e.g., `Get-Process`, `ping`)
    - 🟡 **Non-read-only**: Requires user confirmation (yellow highlight)
    - 🔴 **Dangerous**: Requires user confirmation (red highlight)
- **Cross-Platform**: Automatically detects OS and generates appropriate commands for **Windows PowerShell** or **Linux Bash**.
- **Danger Level Highlighting**: Visual color-coded warnings based on the command's potential impact:
  - <kbd>🟢 Green</kbd>: Harmless / Read-only (e.g., `ls`, `Get-Process`)
  - <kbd>🟡 Yellow</kbd>: Low danger / Creating content (e.g., `mkdir`, `touch`)
  - <kbd>🟠 Orange</kbd>: Medium danger / Modifying individual items (e.g., `rm file.txt`)
  - <kbd>🔴 Red</kbd>: High danger / Batch deletion or system changes (e.g., `rm -rf`, `format`)
- **Multi-Provider Support**: Compatible with OpenAI API and Baidu Wenxin, easily extensible to other AI platforms.
- **Configuration Wizard**: Interactive setup with `ai -c` — each option defaults to the current saved value, press Enter to keep it.
- **Show Config**: Use `ai -i` to quickly view current API configuration (keys are masked for security).
- **Debug Mode**: Enable detailed logging with `ai -d` for troubleshooting.
- **Optimized**: Compact binary (~2.3MB) thanks to build optimizations and UPX compression.

### Installation

#### Building from Source
Requires Go 1.26 or higher.
```bash
git clone https://github.com/yourusername/ai-cmd.git
cd ai-cmd
go build -ldflags="-s -w" -trimpath -o ai.exe .
```

### Configuration

You can configure the tool via environment variables, a configuration file at `~/.ai-cmd.json`, or the interactive config wizard.

**Quick Setup:**
```bash
ai -c
```

**Environment Variables:**
- `AI_CMD_API_KEY`: Your LLM API key.
- `AI_CMD_ENDPOINT`: API endpoint (OpenAI compatible, e.g., `https://api.openai.com/v1`).
- `AI_CMD_MODEL`: The model name to use (e.g., `gpt-4o`).
- `AI_CMD_PROVIDER`: AI provider (`openai` or `baidu`).

**Config File (`~/.ai-cmd.json`):**
```json
{
  "provider": "openai",
  "api_key": "your-key-here",
  "endpoint": "https://api.openai.com/v1",
  "model": "gpt-4o"
}
```

For Baidu Wenxin:
```json
{
  "provider": "baidu",
  "api_key": "your-api-key",
  "secret": "your-secret-key",
  "endpoint": "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions",
  "model": "ernie-4.0-turbo-8k"
}
```

### Usage

To use `ai` from any directory, add the folder containing the `ai` binary to your system's `PATH` environment variable.

**Options:**
- `ai <command>` - Convert natural language to command and execute
- `ai -a <question>` - Ask AI about your system (assistant mode with safety checks)
- `ai -d <command>` - Enable debug mode (output logs to stderr)
- `ai -c` - Run configuration wizard (defaults to current values)
- `ai -i` - Show current API configuration
- `ai -h` - Show help message

```bash
ai <your natural language command>
```

**Examples:**
- `ai show all usb cameras`
- `ai -d explain current directory structure`
- `ai find all files larger than 100MB in c:\data`
- `ai kill the process using port 8080`
- `ai -a are there any unrecognized processes running on my system`
- `ai -a what services are running on my system`

---

<a name="中文"></a>
## 中文

AI Command 是一个轻量级的命令行工具，利用大语言模型（LLM）将自然语言指令转换为可执行的系统命令（PowerShell/Bash）。

### 功能特性

- **自然语言处理**：将"列出 temp 文件夹下所有的日志文件"之类的短语转换为精确的 shell 命令。
- **助手模式 (`-a`)**：向 AI 提问关于系统的问题。AI 会自动执行命令收集信息并给出清晰回答 —— 每条执行的命令和调用目的都完全透明。工具调用过程以灰色输出显示，与最终回答区分。
  - **三级安全机制**：
    - 🟢 **只读命令**：自动执行（如 `Get-Process`, `ping`）
    - 🟡 **非只读命令**：需用户回车确认（黄色高亮）
    - 🔴 **危险命令**：需用户回车确认（红色高亮）
- **跨平台支持**：自动检测操作系统，并为 **Windows PowerShell** 或 **Linux Bash** 生成相应的命令。
- **危险等级高亮**：根据命令的潜在影响进行视觉颜色预警：
  - <kbd>🟢 绿色</kbd>：无害 / 只读类（如 `ls`, `Get-Process`）
  - <kbd>🟡 黄色</kbd>：低风险 / 创建内容（如 `mkdir`, `touch`）
  - <kbd>🟠 橙色</kbd>：中风险 / 修改单个对象（如 `rm file.txt`）
  - <kbd>🔴 红色</kbd>：高风险 / 批量删除或系统级更改（如 `rm -rf`, `format`）
- **多平台支持**：兼容 OpenAI API 和百度文心一言，可轻松扩展至其他 AI 平台。
- **配置向导**：使用 `ai -c` 交互式配置 —— 每个选项默认显示当前值，直接回车即可保留。
- **查看配置**：使用 `ai -i` 快速查看当前 API 配置（密钥自动脱敏）。
- **调试模式**：使用 `ai -d` 启用详细日志输出，便于排查问题。
- **极小体积**：经过编译优化和 UPX 压缩，二进制文件仅约 2.3MB。

### 安装

#### 从源码编译
需要 Go 1.26 或更高版本。
```bash
git clone https://github.com/yourusername/ai-cmd.git
cd ai-cmd
go build -ldflags="-s -w" -trimpath -o ai.exe .
```

### 配置

您可以通过环境变量、位于 `~/.ai-cmd.json` 的配置文件或交互式配置向导进行配置。

**快速设置：**
```bash
ai -c
```

**环境变量：**
- `AI_CMD_API_KEY`: 您的 LLM API 密钥。
- `AI_CMD_ENDPOINT`: API 端点 (兼容 OpenAI 格式)。
- `AI_CMD_MODEL`: 使用的模型名称。
- `AI_CMD_PROVIDER`: AI 提供商 (`openai` 或 `baidu`)。

**配置文件 (`~/.ai-cmd.json`):**
```json
{
  "provider": "openai",
  "api_key": "您的密钥",
  "endpoint": "https://api.openai.com/v1",
  "model": "gpt-4o"
}
```

百度文心一言配置：
```json
{
  "provider": "baidu",
  "api_key": "您的API Key",
  "secret": "您的Secret Key",
  "endpoint": "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions",
  "model": "ernie-4.0-turbo-8k"
}
```

### 使用方法

为了在任意目录下使用 `ai` 命令，请将 `ai` 二进制文件所在的文件夹添加到系统的 `PATH` 环境变量中。

**命令行选项：**
- `ai <指令>` - 转换自然语言并执行命令
- `ai -a <问题>` - 向 AI 提问关于系统的问题（助手模式，带安全检查）
- `ai -d <指令>` - 启用调试模式（输出日志到 stderr）
- `ai -c` - 运行配置向导（默认值为当前配置）
- `ai -i` - 查看当前 API 配置
- `ai -h` - 显示帮助信息

```bash
ai <自然语言指令>
```

**示例：**
- `ai 显示系统内所有 USB 摄像头设备`
- `ai -d 解释当前目录结构`
- `ai 查找 c:\data 中大于 100MB 的所有文件`
- `ai 杀掉占用 8080 端口的进程`
- `ai -a 系统里有没有不认识的进程`
- `ai -a 当前系统运行了哪些服务`

---

### Disclaimer / 免责声明

This tool executes commands generated by AI. Always review the command carefully before pressing Enter. The author is not responsible for any data loss or system damage.
本工具执行由 AI 生成的命令。在按下回车键之前，请务必仔细检查命令。作者对任何数据丢失或系统损坏不承担责任。
