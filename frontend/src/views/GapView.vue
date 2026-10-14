<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAppStore } from '../stores/app'
import { DIMS } from '../constants'
import type { Status } from '../types'

const store = useAppStore()
const planError = ref('')
onMounted(() => {
  if (!store.gap) store.fetchGap()
  if (!store.plan) store.fetchPlan()
})

const LABEL: Record<Status, string> = { strong: '扎实', moderate: '合格', weak: '薄弱', missing: '缺失' }

async function genPlan() {
  planError.value = ''
  try {
    await store.generatePlan()
  } catch (e: any) {
    planError.value = e?.response?.data?.message || '学习路径生成失败，请重试'
  }
}
</script>

<template>
  <div class="page">
    <h1 class="page-title">查漏补缺</h1>
    <p class="page-sub">能力画像 × 「AI 工程师能力模型」逐项对比。分数会随你的笔试/面试结果动态更新（长期记忆）。</p>

    <button v-if="store.profile" class="btn ghost" @click="store.fetchGap()">↻ 重新分析</button>

    <div v-if="store.loading && !store.gap" class="loading">分析中…</div>

    <div v-else-if="!store.profile" class="card center mt16">
      <p class="muted">请先<RouterLink to="/resume" class="link">上传简历</RouterLink>生成能力画像。</p>
    </div>

    <template v-else-if="store.gap">
      <div class="card center mt16">
        <div class="card-title">综合准备度</div>
        <div class="big-score">{{ store.gap.overallScore }}</div>
        <div class="muted">/ 100</div>
      </div>

      <div class="grid cols-2 mt16">
        <div v-for="d in store.gap.dimensions" :key="d.id" class="card">
          <div class="row spread">
            <strong>{{ d.name }}</strong>
            <span class="pill" :class="d.status">{{ LABEL[d.status] }}</span>
          </div>
          <div class="row mt8">
            <div class="bar grow"><span :style="{width: d.score + '%'}"></span></div>
            <span class="kbd">{{ d.score }}</span>
          </div>
          <div class="muted small mt8" style="font-size:12px">权重 {{ (d.weight * 100).toFixed(0) }}%</div>
        </div>
      </div>

      <div v-if="store.gap.recommendations.length" class="card mt16">
        <div class="card-title">补齐建议（按优先级）</div>
        <div v-for="(r, i) in store.gap.recommendations" :key="r.topic" class="mt8" :class="{mt16: i === 0}">
          <div class="row">
            <span class="pill" :class="r.priority === 'high' ? 'missing' : 'weak'">{{ r.priority === 'high' ? '高' : '中' }}</span>
            <strong>{{ r.topic }}</strong>
          </div>
          <div class="row wrap mt8">
            <a v-for="res in r.resources" :key="res.title" :href="res.url" target="_blank" class="link" style="font-size:13px">
              📄 {{ res.title }}
            </a>
          </div>
        </div>
      </div>
      <div v-else class="card center mt16"><p>🎉 没有明显短板，去挑战更难的题吧。</p></div>

      <!-- 学习路径（LLM 生成） -->
      <div class="card mt16" style="border-color:var(--brand)">
        <div class="card-title row spread">
          <span>个性化学习路径</span>
          <button class="btn" :disabled="store.planLoading" @click="genPlan">
            {{ store.planLoading ? '生成中（约 10~30 秒）…' : store.plan ? '↻ 重新生成' : '生成学习路径' }}
          </button>
        </div>
        <p v-if="planError" class="mt8" style="color:var(--bad)">{{ planError }}</p>
        <p v-if="!store.plan && !store.planLoading" class="muted" style="font-size:13px">
          基于画像、各维度得分与近期练习记录，由 LLM 生成分阶段学习计划：学什么、学到什么程度、怎么验收。
        </p>

        <template v-if="store.plan">
          <div class="row wrap mt8" style="gap:8px">
            <span class="pill moderate">{{ store.plan.targetRole }}</span>
            <span class="pill moderate">共 {{ store.plan.totalWeeks }} 周</span>
            <span class="muted" style="font-size:12px">生成于 {{ store.plan.createdAt.slice(0, 10) }}</span>
          </div>
          <div v-for="ph in store.plan.phases" :key="ph.name" class="phase">
            <div class="row spread">
              <strong>{{ ph.name }}</strong>
              <span class="kbd">{{ ph.weeks }}</span>
            </div>
            <div v-for="it in ph.items" :key="it.topic" class="plan-item">
              <div class="row spread">
                <strong style="font-size:14px">{{ it.topic }}</strong>
                <span class="pill moderate">{{ DIMS.find(d => d.id === it.dimensionId)?.name || it.dimensionId }}</span>
              </div>
              <div class="muted mt8" style="font-size:13px">🎯 {{ it.goal }}</div>
              <div class="muted" style="font-size:13px">✓ 验收：{{ it.acceptance }} · 约 {{ it.estDays }} 天</div>
              <div class="row wrap spread mt8">
                <span class="row wrap" style="gap:10px">
                  <a v-for="res in it.resources" :key="res.title" :href="res.url" target="_blank" class="link" style="font-size:13px">📄 {{ res.title }}</a>
                </span>
                <RouterLink :to="'/exam?dims=' + it.dimensionId" class="link" style="font-size:13px">去练习 →</RouterLink>
              </div>
            </div>
          </div>
        </template>
      </div>
    </template>
  </div>
</template>

<style scoped>
.small { font-size: 12px; }
.phase { margin-top: 14px; border-top: 1px solid var(--border); padding-top: 10px; }
.plan-item { background: var(--panel-2); border: 1px solid var(--border); border-radius: 10px; padding: 10px 12px; margin-top: 8px; }
</style>
