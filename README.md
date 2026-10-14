# AI 面霸 · Interview Prep

面向 **AI 开发工程师** 求职备考的 agent 前端。当前用 **Vue3 + TypeScript** 实现，走 mock 数据完整打通产品闭环；后端按 `docs/` 契约用 Go + Eino 实现后，切一次环境变量即可对接。

## 核心闭环

**简历上传 → 能力画像 → 查漏补缺 → 笔试生成/判分 → 模拟面试（含 HITL）→ 技术雷达**

每一次练习/面试结果都会**回写记忆**，Gap 分数随之动态变化，验证"越用越懂你"的 agent 特性。

## 快速开始

```bash
cd frontend
npm install
npm run dev          # 默认 mock，打开 http://localhost:5173
npm run build        # 严格 vue-tsc 检查 + 生产构建
```

## 切换到真实后端

编辑 `frontend/.env`：

```
VITE_USE_MOCK=false
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

启动后端（Go 1.26+，零框架；简历解析走真实 LLM）：

```bash
cd backend
export DASHSCOPE_API_KEY=sk-...   # 或用 LLM_API_KEY；可选 LLM_BASE_URL / LLM_MODEL（默认 DashScope 兼容端点 + qwen-plus）
go run ./cmd/server               # 监听 :8080，ADDR 环境变量可改
```

契约见 `docs/02-api-contract.md`，前端零改动。简历上传为真实解析：PDF/DOCX/MD/TXT 文本抽取 → Eino ChatModel（JSON 约束输出）→ 结构化画像；解析失败返回 40002、模型失败返回 50001，不做静默降级。

## 目录

```
├── docs/                     设计文档（架构、契约、后端设计、数据模型）
│   ├── README.md             索引
│   ├── 01-architecture.md    系统架构
│   ├── 02-api-contract.md    REST 契约
│   ├── 03-backend-eino.md    Go + Eino 后端设计
│   ├── 04-data-model.md      数据模型 & mock 规范
│   └── 05-frontend.md        前端结构
├── frontend/                 Vue3 + TS 前端（mock 已可用）
│   └── src/
│       ├── types/            与契约一一对应的 TS 类型
│       ├── api/mock.ts       Mock 引擎（真实业务逻辑，非静态数据）
│       ├── api/client.ts     API 适配层（mock/真实 一键切换）
│       ├── stores/app.ts     Pinia 状态
│       ├── router/           6 条路由
│       └── views/            6 个页面
└── backend/                  Go 后端（已实现全部 13 个端点；简历解析为 Eino + 真实 LLM，其余为内存版业务逻辑）
    ├── cmd/server/           入口
    ├── internal/api/         路由 + handlers（契约实现）
    ├── internal/domain/      画像等核心结构
    ├── internal/model/       ChatModel 工厂（OpenAI 兼容端点）
    ├── internal/resume/      PDF/DOCX 文本抽取
    ├── internal/graph/       Eino 解析链
    └── testdata/             示例简历（md/docx/pdf）
```

## 体验路径

1. **简历** → 点"用示例简历"或上传任意文件（mock 会按文件名哈希匹配三档画像）
2. **查漏补缺** → 看 5 个维度打分、优先级建议；点**「生成学习路径」**由 LLM 输出分阶段计划（主题/目标/验收/资源），每项可一键「去练习」直达对应维度笔试
3. **笔试** → 选维度生成题，提交 → 判分 → 回到 Gap 页面**刷新看分数变化**（这是记忆回写的直观体现）
4. **模拟面试** → 技术面首题针对你的简历项目提问；每轮回答给出**批注**（已覆盖/可补充要点）；第 2 题后触发 **HITL "深入考察"** 暂停 → 选择"继续深入"或"跳过"；结束生成**逐题复盘 + 优劣总结**
5. **技术雷达** → 已按你的薄弱维度置顶排序
6. **总览** → 综合分数、最近记录、雷达摘要

## 后端交付说明

后端代码在本地按 `docs/03-backend-eino.md` 落地即可，注意：
- Eino 是 alpha 版，pin 精确版本
- 沙箱环境拉不到 Go 模块，需你本地 `go mod tidy` 联网
- 建议按 `docs/03` 的实施顺序：先 mock 硬编码 `/dashboard/summary` 联调，再逐个替换成真实 Eino 图

## 与前端契约对齐要点

`frontend/src/types/index.ts` 严格对应 `docs/02-api-contract.md`；mock 的每一个字段都来自契约。后端返回只要照契约给，前端就能用。
