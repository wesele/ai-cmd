---
name: "⚡ 多模型配置与快速切换"
about: "支持配置多个 AI 模型，支持运行时快速切换不同模型"
labels: ["enhancement", "flexibility"]
---

## 🎯 功能描述
允许用户配置多个 AI 模型（不同提供商、不同模型），支持在命令行中快速切换使用不同的模型。

## 💡 使用场景
- 同时配置 OpenAI、LongCat、本地 Ollama 等多个提供商
- 根据任务复杂度选择不同模型（简单任务用轻量模型，复杂任务用大模型）
- 测试对比不同模型的命令生成质量
- 某个 API 服务不可用时快速切换到备用服务
- 团队共享配置，每人使用自己的 API Key

## 📋 实现建议
1. 配置文件支持多个模型配置：
   ```yaml
   models:
     default: longcat
     longcat:
       api_url: "https://api.longcat.chat/openai/v1/"
       api_key: "ak_xxx"
       model: "LongCat-Flash-Lite"
     openai:
       api_url: "https://api.openai.com/v1/"
       api_key: "sk-xxx"
       model: "gpt-4o-mini"
     local:
       api_url: "http://localhost:11434/v1/"
       api_key: ""
       model: "qwen2.5-coder"
   ```
2. 添加 `-m/--model` 标志指定使用哪个模型配置
3. 添加 `ai models` 子命令管理模型：
   - `ai models list` - 查看所有配置的模型
   - `ai models use <name>` - 切换默认模型
   - `ai models test <name>` - 测试模型连接
4. 环境变量支持：`AI_CMD_MODEL_PROFILE=<name>`
5. 显示当前使用的模型信息（在 debug 模式或配置中）

## ✅ 验收标准
- [ ] 配置文件支持多个模型配置
- [ ] `-m/--model` 标志指定模型
- [ ] `ai models` 子命令管理模型列表
- [ ] 支持快速切换默认模型
- [ ] 支持测试模型连接状态
- [ ] 环境变量支持指定模型配置
- [ ] 帮助信息中显示可用模型
- [ ] 错误提示中显示当前使用的模型

## 📝 使用示例
```bash
# 使用特定模型执行命令
ai -m openai "生成复杂的系统监控命令"
ai -m local "简单的文件查询"  # 使用本地模型，快速响应

# 查看和切换模型
ai models list
ai models use longcat
ai models test openai

# 临时使用环境变量
AI_CMD_MODEL_PROFILE=local ai "查找文件"
```

## 🔗 优先级
中 - 提升灵活性和可靠性
