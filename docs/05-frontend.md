# 05 · 前端设计

## 技术栈

- Vue 3.5（Composition API + `<script setup>`）
- TypeScript 严格模式（`noUnusedLocals/noUnusedParameters/erasableSyntaxOnly`）
- Vite 8
- Pinia 4 状态管理
- Vue Router 4

## 分层

```
views/          页面组件（6 个：Dashboard/Resume/Gap/Exam/Interview/Radar）
stores/         Pinia store（app.ts 集中管理 profile/gap/dashboard/radar）
api/
  ├── client.ts  适配层，按 VITE_USE_MOCK 分发
  └── mock.ts    真实业务逻辑（Gap 打分、判分、面试状态机、记忆回写）
types/          与后端契约一一对齐的 TS 类型
router/         6 条路由
```

## Mock 设计原则

`api/mock.ts` **不是静态假数据**，而是把后端应该做的事在前端先实现一遍：

- Gap 分数由"画像技能命中能力模型锚 × 记忆事件修正"计算得出
- 判分真实比对答案（choice 精确匹配、编程/简答走关键词命中率）
- 面试是多轮状态机，包含 HITL 中断/恢复/跳过
- 每次练习/面试都 append 到 `memoryEvents`，Gap 分数据此变化

好处：产品逻辑在 mock 阶段就能验证，后端只需替换成同契约的 HTTP 实现。

## 记忆持久化

mock 状态写入 localStorage（`ai_prep_mock_state`），刷新页面不丢，可以完整演示"跨会话记忆"。想重置：调 `resetMock()` 或清 localStorage。

## 交互亮点

- **笔试** 提交 → 成绩卡 → **Gap 页刷新即看到分数变化**，直观演示记忆回写
- **面试** 第 2 题后触发 HITL 暂停 → 用户可选"继续深入/跳过"，两条路径都能走到终评
- **技术雷达** 与 Gap 联动，按你的薄弱维度优先排序

## 后端接入

改 `.env` 里 `VITE_USE_MOCK=false` + `VITE_API_BASE_URL` 即可。`client.ts` 每个方法都有 mock 与真实两条分支，切换无感。
