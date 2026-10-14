<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { api } from '../api/client'
import { useAppStore } from '../stores/app'
import type { InterviewType, Msg, InterviewMsgResp, InterviewReport, AnswerNote } from '../types'

const store = useAppStore()

interface ThreadItem { msg: Msg; note?: AnswerNote }

const started = ref(false)
const type = ref<InterviewType>('technical')
const sessionId = ref('')
const thread = ref<ThreadItem[]>([])
const draft = ref('')
const sending = ref(false)
const awaiting = ref(false)
const approvalPrompt = ref('')
const finished = ref(false)
const report = ref<InterviewReport | null>(null)
const lastFeedback = ref<{ score: number; hint: string } | null>(null)
const scroller = ref<HTMLElement | null>(null)

async function scrollDown() {
  await nextTick()
  if (scroller.value) scroller.value.scrollTop = scroller.value.scrollHeight
}

async function start() {
  started.value = true
  finished.value = false
  report.value = null
  thread.value = []
  const r = await api.createInterview(type.value)
  sessionId.value = r.data.sessionId
  const first = await api.interviewFirstMsg(sessionId.value)
  thread.value = [{ msg: { role: 'interviewer', content: first.data.content } }]
  await scrollDown()
}

async function send() {
  if (!draft.value.trim() || sending.value) return
  const content = draft.value.trim()
  thread.value.push({ msg: { role: 'user', content } })
  draft.value = ''
  sending.value = true
  await scrollDown()
  try {
    const r = await api.sendInterview(sessionId.value, content)
    apply(r.data)
  } finally { sending.value = false }
  await scrollDown()
}

function apply(d: InterviewMsgResp) {
  // 批注挂在最近一条用户回答下方
  if (d.note) {
    const lastUser = [...thread.value].reverse().find((t) => t.msg.role === 'user')
    if (lastUser) lastUser.note = d.note
  }
  if (d.awaitingApproval) {
    awaiting.value = true
    approvalPrompt.value = d.approvalPrompt || ''
    if (d.interimFeedback) lastFeedback.value = d.interimFeedback
    return
  }
  if (d.content) thread.value.push({ msg: { role: 'interviewer', content: d.content } })
  if (d.interimFeedback) lastFeedback.value = d.interimFeedback
  if (d.finished) { onFinished() }
}

async function skip() {
  awaiting.value = false
  sending.value = true
  try {
    const r = await api.skipInterview(sessionId.value)
    apply(r.data)
  } finally { sending.value = false }
  await scrollDown()
}

async function resume() {
  awaiting.value = false
  sending.value = true
  try {
    const r = await api.resumeInterview(sessionId.value)
    apply(r.data)
  } finally { sending.value = false }
  await scrollDown()
}

async function onFinished() {
  finished.value = true
  await store.refreshAnalytics()
  const r = await api.interviewReport(sessionId.value)
  report.value = r.data
}
</script>

<template>
  <div class="page">
    <h1 class="page-title">模拟面试</h1>
    <p class="page-sub">AI 面试官多轮问答 + 追问 + HITL「深入考察」确认。结束后生成评估报告，并回写长期记忆。</p>

    <div v-if="!started" class="card center">
      <div class="row wrap center" style="justify-content:center;gap:10px">
        <button class="btn ghost" :class="{ onsel: type === 'technical' }" @click="type = 'technical'">技术面</button>
        <button class="btn ghost" :class="{ onsel: type === 'behavioral' }" @click="type = 'behavioral'">行为面</button>
      </div>
      <p class="muted mt16">{{ type === 'technical' ? '考察 LLM/RAG/微调/系统设计等技术深度' : '考察项目协作、抗压、职业规划' }}</p>
      <button class="btn mt16" @click="start">开始面试 →</button>
      <div v-if="!store.profile" class="muted mt8" style="font-size:13px">提示：先<RouterLink to="/resume" class="link">上传简历</RouterLink>可让提问更贴合你的背景。</div>
    </div>

    <template v-else>
      <div class="card chat">
        <div class="card-title row spread">
          <span>{{ type === 'technical' ? '技术面试官' : 'HR 面试官' }}</span>
          <span class="pill moderate">{{ thread.filter(t => t.msg.role === 'user').length }} 轮</span>
        </div>
        <div ref="scroller" class="thread">
          <template v-for="(t, i) in thread" :key="i">
            <div class="msg" :class="t.msg.role">
              <div class="bubble">{{ t.msg.content }}</div>
            </div>
            <div v-if="t.note" class="note">
              <span class="note-tag">📝 批注</span>
              <span v-if="t.note.covered.length" class="note-ok">✓ {{ t.note.covered.join('、') }}</span>
              <span v-if="t.note.missed.length" class="note-warn">△ 可补充：{{ t.note.missed.join('、') }}</span>
            </div>
          </template>
          <div v-if="sending" class="msg interviewer"><div class="bubble typing">思考中…</div></div>
        </div>

        <div v-if="awaiting" class="hitl">
          <div>⏸ {{ approvalPrompt }}</div>
          <div class="row mt8">
            <button class="btn" @click="resume">继续深入</button>
            <button class="btn ghost" @click="skip">跳过</button>
          </div>
        </div>

        <div v-else-if="!finished" class="composer">
          <textarea v-model="draft" placeholder="回答面试官的问题…" @keydown.enter.exact.prevent="send"></textarea>
          <div class="row spread mt8">
            <span class="muted" style="font-size:12px">Enter 发送 · 每题都有实时反馈</span>
            <button class="btn" :disabled="!draft.trim() || sending" @click="send">发送</button>
          </div>
        </div>

        <div v-if="lastFeedback" class="feedback mt8">
          <span class="pill moderate">本轮评分 {{ lastFeedback.score }}</span>
          <span class="muted" style="font-size:13px">· {{ lastFeedback.hint }}</span>
        </div>
      </div>

      <div v-if="report" class="card mt16" style="border-color:var(--brand)">
        <div class="card-title">面试评估报告</div>
        <div class="row">
          <div class="big-score">{{ report.overallScore }}</div>
          <div class="muted">/ 100 综合评分</div>
        </div>
        <div class="grid cols-3 mt16">
          <div v-for="d in report.dimensions" :key="d.name" class="subcard">
            <div class="muted" style="font-size:13px">{{ d.name }}</div>
            <div class="row mt8"><div class="bar grow"><span :style="{width: d.score + '%'}"></span></div><span class="kbd">{{ d.score }}</span></div>
          </div>
        </div>
        <h3 class="mt16" style="font-size:14px">📋 逐题复盘</h3>
        <div v-for="r in report.review" :key="r.turn" class="review-item">
          <div class="row spread">
            <span class="review-q">Q{{ r.turn + 1 }}. {{ r.question }}</span>
            <span class="kbd">{{ r.score }}</span>
          </div>
          <div v-if="r.covered.length" class="note-ok">✓ {{ r.covered.join('、') }}</div>
          <div v-if="r.missed.length" class="note-warn">△ {{ r.missed.join('、') }}</div>
        </div>
        <h3 class="mt16" style="font-size:14px;color:var(--ok)">✓ 亮点</h3>
        <ul><li v-for="h in report.highlights" :key="h">{{ h }}</li></ul>
        <h3 style="font-size:14px;color:var(--warn)">△ 改进</h3>
        <ul><li v-for="im in report.improvements" :key="im">{{ im }}</li></ul>
        <RouterLink to="/gap"><button class="btn ghost mt8">查看更新后的 Gap →</button></RouterLink>
      </div>

      <button v-if="finished" class="btn ghost mt16" @click="started = false">再来一场</button>
    </template>
  </div>
</template>

<style scoped>
.onsel { background: var(--brand); color: #fff; border-color: var(--brand); }
.chat { padding: 0; overflow: hidden; }
.chat .card-title { padding: 14px 16px; border-bottom: 1px solid var(--border); }
.thread { height: 340px; overflow-y: auto; padding: 16px; display: flex; flex-direction: column; gap: 12px; }
.msg { display: flex; }
.msg.interviewer { justify-content: flex-start; }
.msg.user { justify-content: flex-end; }
.bubble {
  max-width: 75%; padding: 10px 14px; border-radius: 14px; font-size: 14px; line-height: 1.5;
}
.msg.interviewer .bubble { background: var(--panel-2); border: 1px solid var(--border); border-top-left-radius: 4px; }
.msg.user .bubble { background: linear-gradient(135deg, var(--brand), var(--brand-2)); color: #fff; border-top-right-radius: 4px; }
.bubble.typing { color: var(--muted); font-style: italic; }
.composer { padding: 12px 16px 16px; border-top: 1px solid var(--border); }
.hitl { margin: 0 16px 16px; padding: 14px; border: 1px dashed var(--warn); border-radius: 10px; background: rgba(251,191,36,.06); }
.feedback { padding: 0 16px 14px; }
.subcard { background: var(--panel-2); border: 1px solid var(--border); border-radius: 10px; padding: 12px; }
.note {
  display: flex; flex-wrap: wrap; gap: 6px 12px; align-items: baseline;
  margin: -6px 8px 0 auto; max-width: 75%; padding: 6px 10px;
  border: 1px dashed var(--border); border-radius: 8px; background: rgba(255,255,255,.02);
  font-size: 12px;
}
.note-tag { color: var(--muted); flex-shrink: 0; }
.note-ok { color: var(--ok); font-size: 12px; }
.note-warn { color: var(--warn); font-size: 12px; }
.review-item { border-top: 1px solid var(--border); padding: 10px 0; display: flex; flex-direction: column; gap: 4px; }
.review-q { font-size: 13px; color: var(--muted); margin-right: 12px; }
ul { margin: 6px 0 12px; padding-left: 20px; }
li { font-size: 14px; margin: 3px 0; }
</style>
