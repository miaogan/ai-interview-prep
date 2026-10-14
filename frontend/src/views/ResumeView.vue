<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '../stores/app'
import { USE_MOCK } from '../api/client'

const store = useAppStore()
const router = useRouter()
const fileName = ref('')
const error = ref('')

async function onFile(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  error.value = ''
  if (!/\.(pdf|docx|md|txt)$/i.test(file.name)) {
    error.value = USE_MOCK
      ? '暂支持 pdf / docx / md / txt（mock 会按文件名匹配示例画像）'
      : '暂支持 pdf / docx / md / txt（.doc 请另存为 .docx）'
    return
  }
  fileName.value = file.name
  try {
    await store.uploadResume(file)
    router.push('/gap')
  } catch (err: any) {
    error.value = err?.response?.data?.message || '解析失败，请重试'
  }
}

function useSample() {
  const file = new File(['韩梅 示例简历内容'], '韩梅-AI工程师.pdf', { type: 'application/pdf' })
  fileName.value = file.name
  store.uploadResume(file).then(() => router.push('/gap'))
}
</script>

<template>
  <div class="page">
    <h1 class="page-title">简历上传</h1>
    <p v-if="USE_MOCK" class="page-sub">解析技能栈 → 生成能力画像。mock 模式下按文件名哈希，选择"弱/中/强"三档示例画像之一，便于体验不同 Gap 结果。</p>
    <p v-else class="page-sub">上传真实简历（pdf/docx/md/txt），LLM 抽取技能栈与项目经历 → 生成能力画像。解析约需 10~30 秒。</p>

    <div class="card">
      <label class="btn" style="display:inline-block">
        选择文件（pdf/docx/md/txt）
        <input type="file" hidden @change="onFile" accept=".pdf,.doc,.docx,.md,.txt" />
      </label>
      <button v-if="USE_MOCK" class="btn ghost" style="margin-left:10px" @click="useSample">用示例简历</button>
      <p v-if="fileName" class="muted mt8">已选：{{ fileName }}</p>
      <p v-if="error" class="mt8" style="color:var(--bad)">{{ error }}</p>
      <p v-if="store.loading" class="mt8 muted">{{ USE_MOCK ? '解析中…' : 'LLM 解析中，请稍候…' }}</p>
    </div>

    <div v-if="store.profile" class="card mt16">
      <div class="card-title">能力画像</div>
      <div class="row wrap">
        <span class="pill moderate">{{ store.profile.name }}</span>
        <span class="pill moderate">{{ store.profile.yearsExp }} 年经验</span>
        <span class="pill moderate">{{ store.profile.targetRole }}</span>
      </div>

      <h3 class="mt16" style="font-size:15px;color:var(--muted)">技能</h3>
      <table style="width:100%;margin-top:8px;font-size:14px">
        <thead><tr class="muted" style="text-align:left"><th>技能</th><th>类目</th><th>等级</th></tr></thead>
        <tbody>
          <tr v-for="s in store.profile.skills" :key="s.name" style="border-top:1px solid var(--border)">
            <td>{{ s.name }}</td>
            <td class="muted">{{ s.category }}</td>
            <td>
              <div class="row">
                <div class="bar grow"><span :style="{width:(s.level*20)+'%'}"></span></div>
                <span class="kbd">{{ s.level }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <h3 class="mt16" style="font-size:15px;color:var(--muted)">项目</h3>
      <div v-for="p in store.profile.projects" :key="p.name" class="mt8" style="border-left:2px solid var(--brand);padding-left:10px">
        <strong>{{ p.name }}</strong>
        <div class="muted">{{ p.desc }}</div>
        <div class="row wrap mt8">
          <span v-for="t in p.techs" :key="t" class="pill moderate">{{ t }}</span>
        </div>
      </div>

      <RouterLink to="/gap"><button class="btn mt16">查看 Gap 分析 →</button></RouterLink>
    </div>
  </div>
</template>
