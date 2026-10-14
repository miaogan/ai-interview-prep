<script setup lang="ts">
import { onMounted } from 'vue'
import { useAppStore } from '../stores/app'

const store = useAppStore()
const NAME: Record<string, string> = {
  dl_fundamentals: '深度学习基础', llm_app: 'LLM应用', finetune: '模型微调', mlops: 'MLOps/部署', algo: '算法与数据结构',
}
onMounted(() => { if (!store.radar) store.fetchRadar() })
</script>

<template>
  <div class="page">
    <h1 class="page-title">技术雷达</h1>
    <p class="page-sub">近期 AI 领域动态（mock 为内置样例；真实后端通过 MCP 联网检索），并按你的薄弱维度优先排序。</p>

    <button class="btn ghost" @click="store.fetchRadar()">↻ 刷新</button>

    <div v-if="!store.radar" class="loading">加载中…</div>

    <template v-else>
      <p class="muted mt16" style="font-size:13px">生成于 {{ new Date(store.radar.generatedAt).toLocaleString() }} · 与你薄弱维度相关的条目已置顶</p>

      <div class="card mt16" style="border-left:3px solid var(--brand)">
        <div class="card-title" style="margin-bottom:0">📌 关联你的薄弱环节（优先关注）</div>
        <div v-for="it in store.radar.items.filter(i => store.dashboard?.weakDimensions?.includes(i.relevanceDimension))" :key="it.id" class="item">
          <a :href="it.url" target="_blank" class="link title">{{ it.title }}</a>
          <div class="row wrap" style="font-size:12px">
            <span class="src muted">{{ it.source }}</span>
            <span class="dot muted">·</span>
            <span class="muted">{{ it.publishedAt }}</span>
            <span class="pill weak" style="margin-left:6px">关联：{{ NAME[it.relevanceDimension] }}</span>
          </div>
          <p class="muted summary">{{ it.summary }}</p>
        </div>
        <div v-if="!store.radar.items.some(i => store.dashboard?.weakDimensions?.includes(i.relevanceDimension))" class="muted center" style="padding:8px">
          {{ store.dashboard ? '暂无关联短板（你挺全面的 👍）' : '上传简历后可展示个性化关联' }}
        </div>
      </div>

      <div class="card mt16">
        <div class="card-title">全部动态</div>
        <div v-for="it in store.radar.items" :key="it.id" class="item">
          <div class="row spread">
            <a :href="it.url" target="_blank" class="link title">{{ it.title }}</a>
            <span class="muted" style="font-size:12px">{{ it.publishedAt }}</span>
          </div>
          <div class="row wrap" style="margin-top:4px">
            <span class="src muted" style="font-size:12px">{{ it.source }}</span>
            <span v-for="t in it.tags" :key="t" class="tag">{{ t }}</span>
          </div>
          <p class="muted summary">{{ it.summary }}</p>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.item { padding: 10px 0; border-bottom: 1px solid var(--border); }
.item:last-child { border-bottom: none; }
.title { font-size: 15px; font-weight: 600; }
.summary { font-size: 13px; margin: 4px 0 0; line-height: 1.5; }
.src { font-family: ui-monospace, monospace; }
.dot { margin: 0 4px; }
.tag { background: var(--panel-2); color: var(--muted); border: 1px solid var(--border); border-radius: 6px; font-size: 11px; padding: 1px 8px; margin-left: 6px; }
</style>
