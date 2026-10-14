# 03 · Go + Eino 后端设计

后端在本地环境实现（需可联网 `go get`）。Eino 当前为 alpha 版本，**务必 pin 精确版本**（如 `v0.10.0-alpha.x`），升级前读 changelog。

## 依赖

```
github.com/cloudwego/eino            // 核心：组件抽象 + compose(Graph/Chain) + agents
github.com/cloudwego/eino-ext/...    // OpenAI/Ollama ChatModel、Embedding、向量库、MCP
github.com/cloudwego/eino-ext/libs/acl/openai
ollama/ollama                        // 或任意 OpenAI 兼容端点
```

## 建议目录结构

```
backend/
├── cmd/server/main.go          # 启动：路由 + Eino runtime 装配
├── internal/
│   ├── api/                    # HTTP 层（net/http 或 Hertz），实现 02 契约
│   ├── graph/                  # 每个功能一个编排图
│   │   ├── resume_parse.go     # 简历解析 Workflow
│   │   ├── gap_analysis.go     # 查漏补缺 Graph
│   │   ├── exam_gen.go         # 出题 Graph（含校验环）
│   │   ├── interview.go        # 面试 Agent（ReAct + checkpoint）
│   │   └── radar.go            # 技术雷达 Workflow
│   ├── model/                  # ChatModel 工厂（Ollama/OpenAI 切换）
│   ├── memory/                 # SQLite + 向量记忆存储
│   ├── tool/                   # MCP 客户端 + 自定义 InvokableTool
│   └── store/                  # 画像/考试/会话持久化
└── go.mod
```

## 核心编排设计

### 1. 简历解析（compose.Chain，线性）

```
file → [PDF/DOCX 抽取器] → text → ChatModel(结构化抽取 prompt)
     → schema.Parse → Profile
```
- 抽取 prompt 用 JSON schema 约束输出（skills/level/projects）。
- Eino 回调点在 `OnStart/OnEnd/OnError`，接入日志即可观测每步耗时。

### 2. 查漏补缺（compose.Graph，带分支）

```
START → load_profile → compare_model(规则打分)
      → branch: 分数<60 的维度 → retrieve_docs(MCP 检索/RAG)
      → llm_plan(生成补齐建议) → write_memory → END
```
- `compare_model` 不调 LLM，纯规则（能力模型权重 × 技能分），省钱且确定。

### 3. 出题（Graph + 校验环）

```
START → select_topics → llm_gen_questions(batch)
      → validate(node: 结构/答案完整性校验)
      → 条件边: 不合格 → llm_fix → validate（最多 2 次，防死循环）
      → assemble_exam → END
```

### 4. 模拟面试（agents.ReAct + CheckPoint = HITL）

- 用 Eino `agents` 模块跑多轮对话，system prompt 定义面试官人格（技术面/行为面两套）。
- **状态**：每轮的 transcript、各维度累计分，存 session store。
- **HITL**：利用 Eino Graph 的 checkpoint 能力——面试官触发「深入考察」决策点时中断，状态序列化（Redis/SQLite），等用户 `resume` 后恢复执行。这正好实践你关心的断点续传。

### 5. 技术雷达（Chain + MCP）

```
START → mcp_web_search(query=薄弱维度+近30天)
      → dedup → llm_summarize → rank(与画像相关度) → END
```
- MCP server 用 eino-ext 的 mcp client 接入（stdio/SSE）。

## 记忆层（长期记忆）

| 内容 | 存储 |
|---|---|
| 能力画像（结构化） | SQLite 表 `profile`，每次笔试/面试后按结果微调维度分 |
| 错题/弱项事件流 | `memory_events` 表（append-only），可回放 |
| 语义检索（学习笔记/题库） | 向量库（SQLite-vec 或 Milvus 起步可选） |

更新规则示例：笔试答错 `llm_app` 题 → `llm_app` 分 -2（下限钳制）；面试某维度得分 → 指数平滑合并进画像。这让"越用越懂你"可验证。

## 模型接入

`internal/model/factory.go`（已实现）：OpenAI 兼容端点。env `LLM_BASE_URL`（默认 DashScope compatible-mode）、`LLM_API_KEY`（缺省回退读 `DASHSCOPE_API_KEY`，两者都空则报错）、`LLM_MODEL`（默认 `qwen-plus`）。统一开启 `response_format=json_object`。
本地测试可用 Ollama 跑 `qwen2.5`（`LLM_BASE_URL=http://localhost:11434/v1`）—— 注意这是推理层，与框架性能对比时建议 mock ChatModel 固定输出。

## 实施顺序建议

1. `api` 骨架 + `/dashboard/summary`（先全 mock 硬编码，跑通前端联调）
2. 简历解析 Chain（本地文件，效果直观）
3. 出题 Graph + 判分（纯逻辑，可先规则化）
4. 面试 Agent（最难，放最后，HITL 可选做）
5. 技术雷达（依赖 MCP 配置）
```
```
```
```
