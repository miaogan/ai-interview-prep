# 04 · 数据模型与 Mock 规范

## 核心实体（后端持久化 / 前端类型共用）

```
Profile { name, yearsExp, targetRole, skills: Skill[], projects: Project[], updatedAt }
Skill   { name, level(1-5), category }
GapReport { overallScore, dimensions: Dimension[], recommendations: Recommendation[] }
Dimension { id, name, score, weight, status }   // status: strong|moderate|weak|missing
Exam    { examId, title, timeLimitSec, questions: Question[] }
Question { id, type(choice|coding|system_design), dimensionId, difficulty, stem, options?, answer?, explanation? }
ExamResult { examId, score, total, results: ResultItem[], weakPoints[] }
InterviewSession { sessionId, type(technical|behavioral), startedAt, transcript[], finished }
InterviewReport { sessionId, overallScore, dimensions[], highlights[], improvements[], transcript[], review[] }
QuestionReview { turn, question, score, covered[], missed[] }   // 逐题复盘
AnswerNote { covered[], missed[], tip }                          // 每轮回答批注（message 响应下发）
RadarItem { id, title, source, url, summary, publishedAt, tags[], relevanceDimension }
```

TS 类型定义见 `frontend/src/types/index.ts`，与本文档和 02 契约一一对应。

## 「AI 工程师能力模型」内置锚（Gap 分析用）

| 维度 id | 名称 | 权重 | 关键技能 |
|---|---|---|---|
| dl_fundamentals | 深度学习基础 | 0.20 | 反向传播、CNN/RNN、优化器、正则化 |
| llm_app | LLM应用(RAG/Agent) | 0.25 | RAG、Agent、Prompt、向量检索、MCP |
| finetune | 模型微调 | 0.20 | LoRA/QLoRA、SFT、RLHF/DPO、PEFT |
| mlops | MLOps/部署 | 0.15 | 模型服务化、量化推理、监控、K8s |
| algo | 算法与数据结构 | 0.20 | 手撕代码、复杂度、系统设计 |

## Mock 数据规范（前端 `src/api/mock.ts`）

设计原则：**mock 不是假数据堆砌，而是能演示完整闭环**。

- 三个预置人物画像（弱/中/强），上传简历时"解析"为其中之一，方便演示不同 Gap 结果。
- Gap 分析：根据画像 skills 真实计算维度分（不是写死），能体现"查漏补缺"逻辑。
- 笔试：内置一道 RAG 选择 + 一道 LoRA 简答；提交后真实判分（choice 比对答案，coding/简答走关键词匹配）→ 返回成绩和 weakPoints。
- 面试：脚本化多轮（预设追问），技术面首题按画像 projects 定制；每轮回答对照考察要点生成批注（covered/missed/tip），分数由覆盖度驱动；到轮次自动结束并给出终评报告（逐题复盘 + 优劣聚合）；可演示 HITL「深入考察」中断确认。
- 记忆：把每次笔试/面试结果累加进 `memoryEvents`，Gap 分数据此微调，能体现"越练越准"。

这些规则让 mock 阶段就能验证产品逻辑，后端接入时只需把 mock 换成 HTTP，行为一致。
