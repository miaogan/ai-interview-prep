<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '../stores/app'
import type { Status } from '../types'

const store = useAppStore()
onMounted(() => { if (!store.gap) store.fetchGap() })

const LABEL: Record<Status, string> = { strong: '扎实', moderate: '合格', weak: '薄弱', missing: '缺失' }
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
        <p class="muted mt16" style="font-size:13px">💡 去<RouterLink to="/exam" class="link">笔试</RouterLink>和<RouterLink to="/interview" class="link">模拟面试</RouterLink>练一练，回来刷新看分数变化——这就是长期记忆的效果。</p>
      </div>
      <div v-else class="card center mt16"><p>🎉 没有明显短板，去挑战更难的题吧。</p></div>
    </template>
  </div>
</template>

<style scoped>
.small { font-size: 12px; }
</style>
