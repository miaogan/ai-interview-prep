<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../api/client'
import { useAppStore } from '../stores/app'
import { DIMS } from '../constants'
import type { DimensionId, Exam, ExamResult } from '../types'

const store = useAppStore()

// 学习路径「去练习」带 ?dims=algo,llm_app 直达时预选维度
const route = useRoute()
const queryDims = typeof route.query.dims === 'string'
  ? route.query.dims.split(',').filter((d): d is DimensionId => DIMS.some((x) => x.id === d))
  : []
const selected = ref<Set<DimensionId>>(new Set(queryDims.length ? queryDims : ['llm_app', 'finetune']))
const difficulty = ref<'easy' | 'medium' | 'hard'>('medium')
const exam = ref<Exam | null>(null)
const answers = reactive<Record<string, string>>({})
const result = ref<ExamResult | null>(null)
const loading = ref(false)

function toggle(id: DimensionId) {
  if (selected.value.has(id)) selected.value.delete(id)
  else selected.value.add(id)
  selected.value = new Set(selected.value) // trigger reactivity
}

async function generate() {
  if (!selected.value.size) return
  loading.value = true
  result.value = null
  exam.value = null
  try {
    const r = await api.generateExam({ dimensionIds: [...selected.value], count: 10, types: ['choice', 'coding', 'system_design'], difficulty: difficulty.value })
    exam.value = r.data
    for (const k of Object.keys(answers)) delete answers[k]
  } finally { loading.value = false }
}

async function submit() {
  if (!exam.value) return
  loading.value = true
  try {
    const r = await api.submitExam(exam.value.examId, { ...answers })
    result.value = r.data
    await store.refreshAnalytics() // 记忆已更新，刷新 Gap/总览
  } finally { loading.value = false }
}

function res(item: { id: string }) { return result.value?.results.find((x) => x.id === item.id) }
</script>

<template>
  <div class="page">
    <h1 class="page-title">笔试练习</h1>
    <p class="page-sub">按目标维度生成题目（选择/编程/系统设计）。提交后判分并回写长期记忆，Gap 分数随之变化。</p>

    <div class="card">
      <div class="card-title">选择练习维度</div>
      <div class="row wrap">
        <button
          v-for="d in DIMS" :key="d.id"
          class="dimbtn" :class="{ on: selected.has(d.id) }"
          @click="toggle(d.id)">{{ d.name }}</button>
      </div>
      <div class="row mt16">
        <span class="muted">难度：</span>
        <select v-model="difficulty" style="width:auto">
          <option value="easy">简单</option>
          <option value="medium">中等</option>
          <option value="hard">困难</option>
        </select>
        <button class="btn" :disabled="!selected.size || loading" @click="generate">
          {{ loading ? '生成中…' : '生成笔试' }}
        </button>
      </div>
    </div>

    <div v-if="!store.profile" class="card center mt16">
      <p class="muted">建议先<RouterLink to="/resume" class="link">上传简历</RouterLink>，系统会针对你的薄弱维度出题。</p>
    </div>

    <div v-if="result && exam" class="card center mt16" style="border-color:var(--brand)">
      <div class="card-title">本次成绩</div>
      <div class="big-score">{{ result.score }}</div>
      <div class="muted">/ {{ result.total }} · 正确 {{ result.results.filter(r=>r.isCorrect).length }}/{{ result.results.length }}</div>
      <div v-if="result.weakPoints.length" class="mt8">
        <span class="muted">仍薄弱：</span>
        <span v-for="w in result.weakPoints" :key="w" class="pill weak" style="margin:2px">{{ DIMS.find(d=>d.id===w)?.name || w }}</span>
      </div>
      <p class="muted mt8" style="font-size:13px">✓ Gap 分数已更新，<RouterLink to="/gap" class="link">去查看 →</RouterLink></p>
    </div>

    <template v-if="exam">
      <div v-for="(q, idx) in exam.questions" :key="q.id" class="card mt16">
        <div class="row spread">
          <strong>Q{{ idx + 1 }}. {{ q.stem }}</strong>
          <span class="pill moderate">{{ q.type === 'choice' ? '选择' : q.type === 'coding' ? '编程' : '系统设计' }}</span>
        </div>

        <!-- 选择题 -->
        <div v-if="q.type === 'choice'" class="mt8">
          <label v-for="opt in q.options" :key="opt" class="opt" :class="{ done: result && res(q) }">
            <input type="radio" :name="q.id" :value="opt[0]" :disabled="!!result" v-model="answers[q.id]" />
            <span>{{ opt }}</span>
          </label>
          <div v-if="result && res(q)" class="mt8">
            <span :class="res(q)!.isCorrect ? 'pill strong' : 'pill missing'">{{ res(q)!.isCorrect ? '✓ 正确' : '✗ 错误' }}</span>
            <span class="muted mt8" style="font-size:13px;display:block">答案：{{ res(q)!.correctAnswer }} · {{ res(q)!.explanation }}</span>
          </div>
        </div>

        <!-- 编程/系统设计 -->
        <div v-else class="mt8">
          <textarea v-model="answers[q.id]" :disabled="!!result" placeholder="写下你的思路 / 代码…"></textarea>
          <div v-if="result && res(q)" class="mt8">
            <span :class="res(q)!.isCorrect ? 'pill strong' : 'pill weak'">{{ res(q)!.isCorrect ? '要点覆盖充分' : '要点不足' }}</span>
            <span class="muted" style="font-size:13px;display:block;margin-top:6px">{{ res(q)!.explanation }}</span>
          </div>
        </div>
      </div>

      <div class="mt16 center">
        <button v-if="!result" class="btn" :disabled="loading" @click="submit">
          {{ loading ? '判分中…' : '提交并判分' }}
        </button>
        <button v-else class="btn ghost" @click="generate">再来一套</button>
      </div>
    </template>
  </div>
</template>

<style scoped>
.dimbtn {
  background: var(--panel-2); border: 1px solid var(--border); color: var(--muted);
  border-radius: 999px; padding: 6px 14px; font-size: 13px; cursor: pointer; transition: all .15s;
}
.dimbtn.on { background: var(--brand); color: #fff; border-color: var(--brand); }
.opt { display: flex; gap: 8px; align-items: center; padding: 8px 10px; border-radius: 8px; cursor: pointer; }
.opt:hover { background: var(--panel-2); }
.opt.done { cursor: default; }
.opt input { width: auto; }
</style>
