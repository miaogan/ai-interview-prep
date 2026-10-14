# 02 · API 契约

前后端 REST/JSON 契约。**后端按此实现，前端 mock 严格对齐**，切换真实后端时前端零改动。

约定：
- Base URL：`/api/v1`
- 统一响应包裹：`{ "code": 0, "data": <T>, "message": "ok" }`，`code=0` 为成功
- 认证：MVP 阶段单用户，预留 `Authorization: Bearer <token>`
- 时间：ISO 8601 字符串

---

## 1. 简历 Resume

### POST `/resume/upload`
上传简历文件（multipart/form-data，字段 `file`，支持 pdf/docx/md/txt，≤20MB）。

后端流程（真实模式）：文件 → 文本抽取（pdf 用 ledongthuc/pdf，docx 解 zip 读 word/document.xml，md/txt 直读）→ Eino ChatModel（OpenAI 兼容端点，`response_format=json_object`）结构化抽取 → 校验（姓名非空、技能/项目至少其一、level 截断到 1~5、skills≤12 / projects≤4）→ 写入画像。

**错误映射（失败即报错，不回退到示例画像）：**

| 场景 | code |
| --- | --- |
| 文件无法抽取文本（扫描件/加密/伪 PDF/.doc 二进制/格式不支持） | 40002 |
| 模型返回 JSON 但内容不像简历（缺姓名、无技能无项目） | 40002 |
| 模型调用失败（缺 key、网络、限流、输出非 JSON） | 50001 |

**响应 data：**
```json
{
  "resumeId": "r_20261014_001",
  "fileName": "张三-AI工程师.pdf",
  "parsedProfile": {
    "name": "张三",
    "yearsExp": 2,
    "targetRole": "AI开发工程师",
    "skills": [
      { "name": "Python", "level": 4, "category": "language" },
      { "name": "PyTorch", "level": 3, "category": "framework" },
      { "name": "RAG", "level": 2, "category": "llm_app" }
    ],
    "projects": [
      { "name": "企业知识库问答系统", "techs": ["LangChain","FAISS"], "desc": "..." }
    ]
  },
  "uploadedAt": "2026-10-14T08:00:00Z"
}
```

### GET `/resume/profile`
获取当前能力画像（含系统补全的对比结果）。字段同上 `parsedProfile`，追加 `updatedAt`。

---

## 2. 查漏补缺 Gap Analysis

### GET `/gap/analysis`
基于画像对比「AI 工程师能力模型」，输出差距。

**响应 data：**
```json
{
  "overallScore": 68,
  "dimensions": [
    { "id": "dl_fundamentals", "name": "深度学习基础", "score": 75, "weight": 0.2, "status": "strong" },
    { "id": "llm_app", "name": "LLM应用(RAG/Agent)", "score": 55, "weight": 0.25, "status": "weak" },
    { "id": "finetune", "name": "模型微调", "score": 40, "weight": 0.2, "status": "missing" },
    { "id": "mlops", "name": "MLOps/部署", "score": 60, "weight": 0.15, "status": "weak" },
    { "id": "algo", "name": "算法与数据结构", "score": 80, "weight": 0.2, "status": "strong" }
  ],
  "recommendations": [
    { "dimensionId": "finetune", "priority": "high", "topic": "LoRA/QLoRA 微调实战",
      "resources": [{ "title": "HuggingFace PEFT 文档", "url": "...", "type": "doc" }] }
  ]
}
```
`status` ∈ `strong | moderate | weak | missing`。

---

## 3. 笔试 Exam

### POST `/exam/generate`
**请求体：**
```json
{ "dimensionIds": ["llm_app","finetune"], "count": 10,
  "types": ["choice","coding","system_design"], "difficulty": "medium" }
```
**响应 data：**
```json
{
  "examId": "e_001",
  "title": "LLM应用 & 微调 · 专项笔试",
  "timeLimitSec": 1800,
  "questions": [
    { "id": "q1", "type": "choice", "dimensionId": "llm_app", "difficulty": "medium",
      "stem": "RAG 中用于缓解检索噪声的策略是？",
      "options": ["A. 增大 top_k 后直接拼接","B. Rerank 重排","C. 提高温度","D. 关闭向量检索"],
      "answer": null, "explanation": null }
  ]
}
```
> 生成时 `answer`/`explanation` 为 null，提交后由判分接口返回。

### POST `/exam/{examId}/submit`
**请求体：** `{ "answers": { "q1": "B", "q2": "..." } }`
**响应 data：**
```json
{
  "examId": "e_001", "score": 70, "total": 100,
  "results": [
    { "id": "q1", "yourAnswer": "B", "correctAnswer": "B", "isCorrect": true,
      "explanation": "Rerank 用交叉编码器对候选重排，过滤低相关片段。",
      "dimensionId": "llm_app" }
  ],
  "weakPoints": ["finetune"],
  "submittedAt": "2026-10-14T09:00:00Z"
}
```
判分后应**回写记忆**，更新对应维度弱点权重。

---

## 4. 模拟面试 Interview

### POST `/interview/session`
**请求体：** `{ "type": "technical", "targetRole": "AI开发工程师", "rounds": 8 }`
（`type` ∈ `technical | behavioral`）
**响应 data：** `{ "sessionId": "s_001", "type": "technical", "startedAt": "..." }`

> 技术面会根据当前能力画像的 projects 定制首题（针对具体项目提问）；无画像时回退通用题。

### POST `/interview/{sessionId}/message`
多轮对话。首题通过 `{ "role": "user", "content": "__START__" }` 拉取。
**请求体：** `{ "role": "user", "content": "我的回答是……" }`
**响应 data：**
```json
{
  "role": "interviewer",
  "content": "不错。追问：如果向量库规模到千万级，你会怎么优化检索延迟？",
  "turn": 3,
  "finished": false,
  "interimFeedback": { "score": 78, "hint": "可补充：混合检索（关键词+向量）" },
  "note": {
    "covered": ["Embedding 领域匹配", "Chunk 切分调优"],
    "missed": ["混合检索（关键词+向量）", "Rerank 精排"],
    "tip": "可补充：混合检索（关键词+向量）、Rerank 精排"
  }
}
```
- `note` 是对本轮回答的批注：`covered` 已覆盖要点、`missed` 可补充要点（≤3 条）、`tip` 调整建议，帮助用户快速修正回答策略。
- `score` 由要点覆盖度驱动（55 + 45×覆盖率）。
- `finished=true` 时附带终评。

### GET `/interview/{sessionId}/report`
**响应 data：**
```json
{
  "sessionId": "s_001", "type": "technical",
  "overallScore": 76,
  "dimensions": [{ "name": "技术深度", "score": 80 }, { "name": "表达清晰度", "score": 72 }],
  "highlights": ["Q2 覆盖：Embedding 领域匹配、Chunk 切分调优"],
  "improvements": ["Q4 可补充：rank 过大的风险、取值经验与验证"],
  "transcript": [{ "role": "interviewer", "content": "..." }, { "role": "user", "content": "..." }],
  "review": [
    { "turn": 0, "question": "...", "score": 78,
      "covered": ["技术栈与架构交代", "个人职责与模块"], "missed": ["量化成果"] }
  ]
}
```
`overallScore` 为各轮得分均值；`highlights`/`improvements` 从逐题批注聚合；`review` 逐题复盘优劣。

### HITL 中断（可选，Eino checkpoint）
`POST /interview/{sessionId}/message` 返回 `finished=false, awaitingApproval=true` 时，
前端展示「面试官想深入考察 X，是否继续？」，用户确认后 `POST /interview/{sessionId}/resume` 恢复。

---

## 5. 技术雷达 Tech Radar

### GET `/radar/latest?topic=llm&limit=10`
**响应 data：**
```json
{
  "items": [
    { "id": "t1", "title": "某开源模型发布", "source": "arXiv/GitHub",
      "url": "...", "summary": "……", "publishedAt": "2026-10-10",
      "tags": ["LLM","开源模型"], "relevanceDimension": "llm_app" }
  ],
  "generatedAt": "2026-10-14T08:30:00Z"
}
```
`relevanceDimension` 关联到你的薄弱维度，实现「最新动态 × 个人短板」推荐。

---

## 6. Dashboard 聚合

### GET `/dashboard/summary`
**响应 data：**
```json
{
  "overallScore": 68,
  "weakDimensions": ["finetune","llm_app"],
  "recentExams": [{ "examId": "e_001", "score": 70, "at": "..." }],
  "recentInterviews": [{ "sessionId": "s_001", "score": 76, "at": "..." }],
  "studyStreakDays": 5,
  "radarHighlights": [{ "title": "...", "url": "..." }]
}
```

---

## 错误码

| code | 含义 |
|---|---|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40002 | 文件解析失败 |
| 40401 | 资源不存在 |
| 50001 | 模型调用失败 |
| 50002 | 内部错误 |
