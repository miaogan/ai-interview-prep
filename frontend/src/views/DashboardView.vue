<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useAppStore } from '../stores/app'

const store = useAppStore()
onMounted(() => { if (!store.dashboard) store.fetchDashboard() })
</script>

<template>
  <div class="page">
    <h1 class="page-title">备考总览</h1>
    <p class="page-sub">面向 AI 开发工程师 · 简历 → 查漏补缺 → 笔试 → 面试 的完整闭环</p>

    <div v-if="!store.hasProfile" class="card center mt24">
      <h3>还没有能力画像</h3>
      <p class="muted">上传简历，系统会解析你的技能栈并生成 Gap 分析。</p>
      <RouterLink to="/resume"><button class="btn mt8">上传简历 →</button></RouterLink>
    </div>

    <template v-else>
      <div class="grid cols-3">
        <div class="card center">
          <div class="card-title">综合准备度</div>
          <div class="big-score">{{ store.dashboard?.overallScore ?? '—' }}</div>
          <div class="muted small">/ 100</div>
        </div>
        <div class="card center">
          <div class="card-title">连续备考</div>
          <div class="big-score">{{ store.dashboard?.studyStreakDays ?? 0 }}</div>
          <div class="muted small">天</div>
        </div>
        <div class="card">
          <div class="card-title">重点补齐维度</div>
          <div class="row wrap">
            <RouterLink v-for="d in store.dashboard?.weakDimensions" :key="d" to="/gap" class="pill weak" style="margin:2px">{{ d }}</RouterLink>
            <span v-if="!store.dashboard?.weakDimensions?.length" class="muted">暂无明显短板 👍</span>
          </div>
        </div>
      </div>

      <div class="grid cols-2 mt16">
        <div class="card">
          <div class="card-title">最近笔试</div>
          <div v-if="!store.dashboard?.recentExams?.length" class="muted">还没有记录。<RouterLink to="/exam" class="link">去练一套 →</RouterLink></div>
          <div v-for="e in store.dashboard?.recentExams" :key="e.examId" class="row spread" style="padding:6px 0;border-bottom:1px solid var(--border)">
            <span class="kbd">{{ e.examId.slice(0,8) }}</span><span class="pill moderate">{{ e.score }}</span>
          </div>
        </div>
        <div class="card">
          <div class="card-title">最近面试</div>
          <div v-if="!store.dashboard?.recentInterviews?.length" class="muted">还没有记录。<RouterLink to="/interview" class="link">模拟一场 →</RouterLink></div>
          <div v-for="i in store.dashboard?.recentInterviews" :key="i.sessionId" class="row spread" style="padding:6px 0;border-bottom:1px solid var(--border)">
            <span class="kbd">{{ i.sessionId.slice(0,8) }}</span><span class="pill moderate">{{ i.score }}</span>
          </div>
        </div>
      </div>

      <div class="card mt16">
        <div class="card-title">技术雷达速览</div>
        <div v-for="r in store.dashboard?.radarHighlights" :key="r.title" style="padding:5px 0">
          <a :href="r.url" target="_blank" class="link">• {{ r.title }}</a>
        </div>
      </div>
    </template>
  </div>
</template>
