# AI Interview Prep Agent — 项目设计文档

面向「AI 开发工程师」求职备考的智能体。核心闭环：

```
简历上传 → 能力画像 → 查漏补缺 → 笔试练习 → 模拟面试 → 最新技术雷达
                 ↑__________________________________________|
                        （长期记忆：持续更新你的弱点画像）
```

## 文档索引

| 文档 | 内容 |
|---|---|
| [01-architecture.md](./01-architecture.md) | 系统架构、模块划分、技术选型 |
| [02-api-contract.md](./02-api-contract.md) | 前后端 REST API 契约（后端实现依据） |
| [03-backend-eino.md](./03-backend-eino.md) | Go + Eino 后端设计（Graph 编排、记忆、MCP） |
| [04-data-model.md](./04-data-model.md) | 数据模型与 mock 数据规范 |
| [05-frontend.md](./05-frontend.md) | 前端结构、mock 策略、页面说明 |

## 技术栈

- **前端**：Vue 3 + TypeScript + Vite + Pinia + Vue Router（本仓库 `frontend/` 已实现，mock 数据打通全流程）
- **后端**：Go 1.22 + CloudWeGo Eino（Graph/Workflow 编排）+ Eino-ext（OpenAI/Ollama）
- **记忆**：SQLite + 向量检索（长期记忆 / 用户画像）
- **工具接入**：MCP（联网检索最新技术、题库等）

## 当前状态

- ✅ 设计文档（本目录）
- ✅ 前端 Vue3+TS（mock 数据，全流程可跑）
- ⏳ 后端 Go/Eino：需在可联网拉取 Go 模块的本地环境实现，遵循 `02-api-contract.md` 契约

## 前后端切换

前端通过环境变量 `VITE_USE_MOCK` 控制数据来源：
- `true`（默认）：走内置 mock 服务，无需后端即可跑通全流程
- `false`：走真实后端 `VITE_API_BASE_URL`

契约保持一致，替换后端时前端零改动。
