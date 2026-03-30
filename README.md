# DayLens

工作活动追踪与智能分析系统。Tauri 桌面客户端 + Go 服务端。

后台静默记录你的工作活动（前台应用、窗口标题、截图、OCR），通过 AI 生成日报周报，帮你搞清楚时间花在了哪。

## 功能

- **采集**：前台窗口监控 + 截屏 + OCR，空闲/锁屏自动暂停
- **隐私**：本地脱敏、应用黑名单、关键词过滤、域名屏蔽
- **统计**：仪表盘、时间线、小时摘要、应用排行
- **AI 日报**：自动生成 Markdown 日报/周报，支持 Ollama / OpenAI / Claude / Gemini
- **工作智能**：会话聚合、意图分类（9 种）、TODO 自动提取、深度工作检测
- **搜索**：窗口标题 + OCR 全文搜索，AI 问答
- **同步**：WebSocket 实时推送，断网自动缓冲恢复

## 快速开始

### 前置要求

Node.js 18+、Rust 1.75+、Go 1.24+、Docker 24+

> 客户端目前仅支持 Windows，暂不支持 macOS。服务端不限平台。

### 启动服务端

```bash
cd server
docker-compose up -d          # 启动 Server + PostgreSQL
docker-compose logs -f server # 查看日志，控制台会输出自动生成的 Token
```

本地开发：

```bash
docker-compose up -d db       # 仅启动数据库
go run ./cmd/server           # 本地运行
```

### 启动客户端

```bash
cd client
npm install
npm run tauri dev             # 开发模式
npm run tauri build           # 生产构建
```

启动后在设置页填入服务端 URL（默认 `http://localhost:8080`）和控制台输出的 Token。

## 配置

优先级：环境变量 > config.yaml > 默认值

### 服务端（config.yaml）

```yaml
auth:
  token: ""                    # 留空自动生成

database:
  url: "postgres://daylens:daylens_secret@localhost:5432/daylens?sslmode=disable"

ai:
  provider: "ollama"           # ollama / openai / claude / gemini
  endpoint: "http://localhost:11434"
  model: "qwen2.5"
  api_key: ""

storage:
  retention_days: 30
  max_storage_mb: 2048
```

### 环境变量

`DAYLENS_` 前缀覆盖任意配置项：

```bash
DAYLENS_DATABASE_URL=postgres://user:pass@host:5432/db
DAYLENS_AUTH_TOKEN=your-token
DAYLENS_AI_PROVIDER=openai
DAYLENS_AI_API_KEY=sk-xxxx
```

### 客户端

配置文件：`%APPDATA%/daylens/config.json`

- 截屏间隔：默认 30 秒
- 空闲超时：默认 3 分钟
- OCR：默认开启
- 离线缓冲：默认开启

## 生产部署

```bash
cd server
docker-compose --profile production up -d   # 含 Nginx TLS 代理
```

## 测试

```bash
cd server && go test ./...            # 服务端
cd client/src-tauri && cargo test     # 客户端
```
