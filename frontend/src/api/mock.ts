// Mock 引擎：真实可运行的产品逻辑，非静态假数据
// 演示完整闭环：简历解析 -> Gap分析 -> 出题/判分 -> 面试 -> 记忆更新
import type {
  Profile, ResumeUploadResp, GapReport, Dimension, DimensionId, Status,
  Exam, ExamGenerateReq, ExamResult, Question, QuestionType, ResultItem,
  InterviewSession, InterviewMsgResp, InterviewReport, Msg, AnswerNote, QuestionReview,
  RadarResp, DashboardSummary, StudyPlan, PlanPhase,
} from '../types'

const now = () => new Date().toISOString()

// ---------- 内置能力模型锚 ----------
const MODEL: Record<DimensionId, { name: string; weight: number; skills: string[] }> = {
  dl_fundamentals: { name: '深度学习基础', weight: 0.2, skills: ['PyTorch', 'TensorFlow', 'CNN', 'RNN', 'Transformer'] },
  llm_app: { name: 'LLM应用(RAG/Agent)', weight: 0.25, skills: ['RAG', 'LangChain', 'Agent', 'Prompt', 'MCP', '向量检索'] },
  finetune: { name: '模型微调', weight: 0.2, skills: ['LoRA', 'QLoRA', 'SFT', 'RLHF', 'DPO', 'PEFT', '微调'] },
  mlops: { name: 'MLOps/部署', weight: 0.15, skills: ['Docker', 'K8s', '量化', 'vLLM', '部署', '服务化'] },
  algo: { name: '算法与数据结构', weight: 0.2, skills: ['算法', '数据结构', '系统设计', 'LeetCode'] },
}
const DIM_IDS = Object.keys(MODEL) as DimensionId[]

// ---------- 三个预置画像（弱/中/强），上传时按文件名哈希选择，便于演示不同结果 ----------
const PRESET_PROFILES: Profile[] = [
  { name: '李雷', yearsExp: 1, targetRole: 'AI开发工程师',
    skills: [ {name:'Python',level:3,category:'language'}, {name:'PyTorch',level:2,category:'framework'},
      {name:'RAG',level:1,category:'llm_app'}, {name:'数据分析',level:4,category:'other'} ],
    projects: [ {name:'课程作业：图像分类',techs:['CNN'],desc:'基于ResNet的猫狗分类'} ] },
  { name: '韩梅', yearsExp: 3, targetRole: 'AI开发工程师',
    skills: [ {name:'Python',level:4,category:'language'}, {name:'PyTorch',level:4,category:'framework'},
      {name:'Transformer',level:3,category:'framework'}, {name:'LangChain',level:3,category:'llm_app'},
      {name:'RAG',level:2,category:'llm_app'}, {name:'Docker',level:3,category:'mlops'},
      {name:'LoRA',level:2,category:'finetune'}, {name:'算法',level:4,category:'algo'} ],
    projects: [ {name:'企业知识库问答',techs:['LangChain','FAISS','RAG'],desc:'RAG检索问答'},
      {name:'垂域模型微调',techs:['LoRA','PEFT'],desc:'基于LoRA的行业微调'} ] },
  { name: '王强', yearsExp: 5, targetRole: 'AI开发工程师',
    skills: [ {name:'Python',level:5,category:'language'}, {name:'PyTorch',level:5,category:'framework'},
      {name:'Transformer',level:5,category:'framework'}, {name:'RAG',level:5,category:'llm_app'},
      {name:'Agent',level:5,category:'llm_app'}, {name:'MCP',level:4,category:'llm_app'},
      {name:'LoRA',level:5,category:'finetune'}, {name:'RLHF',level:4,category:'finetune'},
      {name:'DPO',level:4,category:'finetune'}, {name:'vLLM',level:4,category:'mlops'},
      {name:'K8s',level:4,category:'mlops'}, {name:'系统设计',level:4,category:'algo'},
      {name:'算法',level:5,category:'algo'} ],
    projects: [ {name:'生产级Agent平台',techs:['Agent','MCP','vLLM'],desc:'多Agent编排平台'},
      {name:'大模型对齐',techs:['RLHF','DPO'],desc:'偏好对齐训练'} ] },
]

// ---------- 记忆：画像 + 事件流，驱动"越练越准" ----------
interface State {
  profile: Profile | null
  memoryEvents: { type: 'exam' | 'interview'; dimensionId: DimensionId; delta: number }[]
}
let state: State = { profile: null, memoryEvents: [] }

// 从 localStorage 恢复，刷新不丢闭环
const LS = 'ai_prep_mock_state'
function persist() { localStorage.setItem(LS, JSON.stringify(state)) }
function load() {
  const raw = localStorage.getItem(LS)
  if (raw) { try { state = JSON.parse(raw) } catch { /* ignore */ } }
}
load()

export function resetMock() { state = { profile: null, memoryEvents: [] }; persist() }

// ---------- 1. 简历上传（解析为预置画像） ----------
export function mockUploadResume(fileName: string): ResumeUploadResp {
  const idx = [...fileName].reduce((a, c) => a + c.charCodeAt(0), 0) % PRESET_PROFILES.length
  const profile = structuredClone(PRESET_PROFILES[idx])
  state.profile = profile
  persist()
  return { resumeId: 'r_' + Date.now(), fileName, parsedProfile: profile, uploadedAt: now() }
}

export function mockGetProfile() { return state.profile }

// 计算某维度分：技能匹配 + 记忆事件修正
function scoreDimension(id: DimensionId): number {
  if (!state.profile) return 0
  const anchors = MODEL[id].skills
  let sum = 0
  for (const s of state.profile.skills) {
    const hit = anchors.some((a) => s.name.toLowerCase().includes(a.toLowerCase()) || a.toLowerCase().includes(s.name.toLowerCase()))
    if (hit) sum += s.level * 20
  }
  // 记忆修正：每次相关笔试/面试结果微调
  const mem = state.memoryEvents.filter((e) => e.dimensionId === id).reduce((a, e) => a + e.delta, 0)
  const base = Math.min(sum, 100)
  return Math.max(0, Math.min(100, Math.round(base + mem)))
}

function statusOf(score: number): Status {
  if (score >= 75) return 'strong'
  if (score >= 60) return 'moderate'
  if (score >= 40) return 'weak'
  return 'missing'
}

// ---------- 2. 查漏补缺 ----------
export function mockGapAnalysis(): GapReport {
  const dims: Dimension[] = DIM_IDS.map((id) => {
    const score = scoreDimension(id)
    return { id, name: MODEL[id].name, score, weight: MODEL[id].weight, status: statusOf(score) }
  })
  const overall = Math.round(dims.reduce((a, d) => a + d.score * d.weight, 0))
  const recMap: Record<DimensionId, { topic: string; resources: { title: string; url: string; type: string }[] }> = {
    dl_fundamentals: { topic: 'Transformer 原理与手撕注意力', resources: [{ title: 'The Illustrated Transformer', url: 'https://jalammar.github.io/illustrated-transformer/', type: 'blog' }] },
    llm_app: { topic: 'RAG 进阶：混合检索 + Rerank + Agent', resources: [{ title: 'LlamaIndex RAG', url: 'https://docs.llamaindex.ai/', type: 'doc' }, { title: 'MCP 官方', url: 'https://modelcontextprotocol.io/', type: 'doc' }] },
    finetune: { topic: 'LoRA/QLoRA 微调实战', resources: [{ title: 'HuggingFace PEFT', url: 'https://huggingface.co/docs/peft', type: 'doc' }] },
    mlops: { topic: 'vLLM 部署 + 量化推理', resources: [{ title: 'vLLM 文档', url: 'https://docs.vllm.ai/', type: 'doc' }] },
    algo: { topic: '系统设计 + 手撕代码高频题', resources: [{ title: 'System Design Primer', url: 'https://github.com/donnemartin/system-design-primer', type: 'repo' }] },
  }
  const recommendations = dims
    .filter((d) => d.status === 'weak' || d.status === 'missing')
    .sort((a, b) => b.weight * (100 - b.score) - a.weight * (100 - a.score))
    .map((d) => ({ dimensionId: d.id, priority: (d.weight >= 0.2 && d.status === 'missing' ? 'high' : 'medium') as 'high' | 'medium', topic: recMap[d.id].topic, resources: recMap[d.id].resources }))
  return { overallScore: overall, dimensions: dims, recommendations }
}

// ---------- 2.5 学习路径（mock：按薄弱维度确定性合成，结构同真实 LLM 输出） ----------
const PLAN_CATALOG: Record<DimensionId, { topic: string; resources: { title: string; url: string; type: string }[] }> = {
  dl_fundamentals: { topic: 'Transformer 原理与手撕注意力', resources: [{ title: 'The Illustrated Transformer', url: 'https://jalammar.github.io/illustrated-transformer/', type: 'blog' }] },
  llm_app: { topic: 'RAG 进阶：混合检索 + Rerank + Agent', resources: [{ title: 'LlamaIndex RAG', url: 'https://docs.llamaindex.ai/', type: 'doc' }, { title: 'MCP 官方', url: 'https://modelcontextprotocol.io/', type: 'doc' }] },
  finetune: { topic: 'LoRA/QLoRA 微调实战', resources: [{ title: 'HuggingFace PEFT', url: 'https://huggingface.co/docs/peft', type: 'doc' }] },
  mlops: { topic: 'vLLM 部署 + 量化推理', resources: [{ title: 'vLLM 文档', url: 'https://docs.vllm.ai/', type: 'doc' }] },
  algo: { topic: '系统设计 + 手撕代码高频题', resources: [{ title: 'System Design Primer', url: 'https://github.com/donnemartin/system-design-primer', type: 'repo' }] },
}

let planState: StudyPlan | null = null

export function mockGeneratePlan(): StudyPlan {
  const dims = DIM_IDS.map((id) => ({ id, score: scoreDimension(id) }))
    .sort((a, b) => a.score - b.score)
    .filter((d) => d.score < 75)
  const phases: PlanPhase[] = []
  const chunk = 2
  for (let i = 0; i < dims.length; i += chunk) {
    const part = dims.slice(i, i + chunk)
    const phaseIdx = phases.length
    phases.push({
      name: `阶段${['一', '二', '三'][phaseIdx] || phaseIdx + 1}：${part.map((d) => MODEL[d.id].name).join(' + ')}`,
      weeks: `第${phaseIdx * 2 + 1}-${phaseIdx * 2 + 2}周`,
      items: part.map((d) => ({
        dimensionId: d.id,
        topic: PLAN_CATALOG[d.id].topic,
        goal: `系统过一遍核心概念并完成配套练习（当前 ${d.score} 分）`,
        acceptance: `笔试「${MODEL[d.id].name}」维度达到 70 分`,
        estDays: 10,
        resources: PLAN_CATALOG[d.id].resources,
      })),
    })
  }
  planState = { createdAt: now(), targetRole: state.profile?.targetRole || 'AI工程师', totalWeeks: Math.max(2, phases.length * 2), phases }
  return planState
}

export function mockGetPlan(): StudyPlan | null { return planState }

// ---------- 3. 笔试 ----------
export function mockGenerateExam(req: ExamGenerateReq): Exam {
  const q: Question[] = []
  const pool: Omit<Question, 'id' | 'answer' | 'explanation'>[] = []
  for (const dim of req.dimensionIds) {
    if (dim === 'llm_app') pool.push(
      { type: 'choice', dimensionId: dim, difficulty: req.difficulty, stem: 'RAG 中用于缓解检索噪声、提升答案质量的关键策略是？', options: ['A. 无限增大 top_k 直接拼接', 'B. Rerank 对候选重排', 'C. 提高采样温度', 'D. 关闭向量检索'] },
      { type: 'coding', dimensionId: dim, difficulty: req.difficulty, stem: '用伪代码/Python 描述：你如何为向量库（千万级）设计低延迟检索方案？（关键词：ANN/IVF/HNSW、分片、Rerank、缓存）' },
    )
    if (dim === 'finetune') pool.push(
      { type: 'choice', dimensionId: dim, difficulty: req.difficulty, stem: 'QLoRA 相比标准 LoRA 的核心改进是？', options: ['A. 全参数微调', 'B. 4-bit 量化底座 + LoRA 适配器', 'C. 去掉反向传播', 'D. 只用 CPU 训练'] },
      { type: 'system_design', dimensionId: dim, difficulty: req.difficulty, stem: '设计一个支持 DPO 偏好对齐的训练流水线，需覆盖数据、训练、评估、部署。（关键词：偏好数据、ref model、beta、奖励、评估集）' },
    )
    if (dim === 'dl_fundamentals') pool.push(
      { type: 'choice', dimensionId: dim, difficulty: req.difficulty, stem: 'Transformer 中缩放点积注意力里除以 sqrt(dk) 的目的是？', options: ['A. 加速收敛', 'B. 防止点积过大导致 softmax 梯度过小', 'C. 减少参数量', 'D. 增加模型容量'] },
    )
    if (dim === 'mlops') pool.push(
      { type: 'choice', dimensionId: dim, difficulty: req.difficulty, stem: 'vLLM 提升推理吞吐的核心技术是？', options: ['A. 静态 batching', 'B. PagedAttention', 'C. int8 训练', 'D. 关闭 KV cache'] },
    )
    if (dim === 'algo') pool.push(
      { type: 'coding', dimensionId: dim, difficulty: req.difficulty, stem: '实现函数 topK(nums, k) 返回频率最高的 k 个元素。（关键词：哈希计数、堆/桶排序、O(n)）' },
    )
  }
  const limit = Math.min(req.count || 10, pool.length)
  for (let i = 0; i < limit; i++) q.push({ ...pool[i], id: 'q' + (i + 1) })
  const title = req.dimensionIds.map((d) => MODEL[d].name.split('(')[0]).join(' & ')
  return { examId: 'e_' + Date.now(), title: title + ' · 专项笔试', timeLimitSec: 1800, questions: q }
}

// 参考答案与判分逻辑（简答/编程用关键词匹配）
const KEYS: Record<DimensionId, string[]> = {
  dl_fundamentals: ['softmax', '梯度', 'dk', '方差', '缩放'],
  llm_app: ['rerank', 'ann', 'hnsw', 'ivf', '分片', '缓存', '重排'],
  finetune: ['偏好', 'ref', 'beta', 'dpo', '奖励', '评估'],
  mlops: ['paged', 'attention', 'kv', '吞吐', 'batch'],
  algo: ['哈希', '堆', '桶', 'o(n)', 'count'],
}
function gradeAnswer(dim: DimensionId, ans: string): boolean {
  const a = (ans || '').toLowerCase()
  if (!a.trim()) return false
  const keys = KEYS[dim]
  const hit = keys.filter((k) => a.includes(k)).length
  return hit >= Math.min(2, keys.length)
}
const REF: Record<string, { answer: string; explanation: string }> = {
  'llm_app_c': { answer: 'B', explanation: 'Rerank 用交叉编码器对候选重排，过滤低相关片段，缓解噪声。' },
  'finetune_c': { answer: 'B', explanation: 'QLoRA 将底座量化到 4-bit 再训 LoRA 适配器，显存大幅下降。' },
  'dl_fundamentals_c': { answer: 'B', explanation: '除以 sqrt(dk) 防止点积过大让 softmax 进入饱和区、梯度消失。' },
  'mlops_c': { answer: 'B', explanation: 'PagedAttention 像操作系统分页一样管理 KV cache，显著提升吞吐。' },
}
function refKey(type: QuestionType, dim: DimensionId) { return dim + '_' + (type === 'choice' ? 'c' : 'x') }

export function mockSubmitExam(examId: string, answers: Record<string, string>): ExamResult {
  const exam = mockGenerateExam({ dimensionIds: DIM_IDS, count: 99, types: ['choice', 'coding', 'system_design'], difficulty: 'medium' })
  const target = currentExam || exam
  const results: ResultItem[] = []
  let correct = 0
  for (const q of target.questions) {
    const your = answers[q.id] || ''
    let isCorrect = false, ref = '', expl = ''
    if (q.type === 'choice') {
      const r = REF[refKey(q.type, q.dimensionId)] || { answer: 'A', explanation: '' }
      ref = r.answer; expl = r.explanation; isCorrect = your.trim().toUpperCase() === ref
    } else {
      ref = '(关键词命中即得分)'; expl = '参考要点：' + KEYS[q.dimensionId].join(' / '); isCorrect = gradeAnswer(q.dimensionId, your)
    }
    if (isCorrect) correct++
    results.push({ id: q.id, yourAnswer: your, correctAnswer: ref, isCorrect, explanation: expl, dimensionId: q.dimensionId })
  }
  const score = Math.round((correct / Math.max(1, target.questions.length)) * 100)
  // 记忆回写：每个维度按对错微调
  const perDim: Record<string, { c: number; t: number }> = {}
  for (const r of results) {
    perDim[r.dimensionId] ||= { c: 0, t: 0 }
    perDim[r.dimensionId].t++
    if (r.isCorrect) perDim[r.dimensionId].c++
    state.memoryEvents.push({ type: 'exam', dimensionId: r.dimensionId, delta: r.isCorrect ? 2 : -2 })
  }
  persist()
  const weakPoints = Object.keys(perDim).filter((d) => perDim[d].c / perDim[d].t < 0.6) as DimensionId[]
  examHistory.push({ examId, score, at: now() })
  return { examId, score, total: 100, results, weakPoints, submittedAt: now() }
}

let currentExam: Exam | null = null
const examHistory: { examId: string; score: number; at: string }[] = []
export function mockGenerateExamAndCache(req: ExamGenerateReq): Exam {
  currentExam = mockGenerateExam(req)
  return currentExam
}

// ---------- 4. 模拟面试（项目定制脚本 + 批注 + 逐题复盘 + HITL） ----------
interface InterviewState {
  type: string; script: string[]; transcript: Msg[]; idx: number
  awaiting: boolean; finished: boolean
  reviews: QuestionReview[]; scores: number[]; answers: string[]
}
const sessions: Record<string, InterviewState> = {}
const interviewHistory: { sessionId: string; score: number; at: string }[] = []

const TECH_SCRIPT = [
  '请用 2-3 分钟介绍一个你做过的 LLM 相关项目，以及你负责的模块。',
  '你提到用了 RAG，那检索召回不佳时你怎么排查和优化？',
  '我想深入考察向量库选型：HNSW 和 IVF 的区别与适用场景，以及 Rerank 放在哪一步？',
  '模型微调方面，LoRA 的 rank 设大了或设小了分别有什么问题？如何选？',
  '最后一题：如果要上线一个日均百万请求的推理服务，你会怎么设计保证低延迟？',
]
const BEHAV_SCRIPT = [
  '讲一次你和同事在技术方案上有分歧的经历，你怎么处理的？',
  '描述一个你在 deadline 压力下完成任务的例子。',
  '你最近主动学的新技术是什么？怎么学的？',
  '为什么想做 AI 开发工程师，你的规划是什么？',
]
const DEEP_IDX = 2 // technical 的 HITL 触发点

function buildScript(type: string): string[] {
  if (type !== 'technical') return [...BEHAV_SCRIPT]
  const s = [...TECH_SCRIPT]
  const proj = state.profile?.projects?.[0]
  if (proj) {
    s[0] = `你在《${proj.name}》中用了 ${proj.techs.join('、')}，请介绍这个项目的整体架构、你负责的核心模块，以及最有挑战的部分。`
    const t = proj.techs.map((x) => x.toLowerCase()).join(' ')
    if (!/rag|langchain|向量|检索|faiss/.test(t)) {
      s[1] = '在这个项目里，你遇到的最难的技术问题是什么？怎么定位和解决的？'
    }
  }
  return s
}

interface RubricPoint { label: string; keys: string[] }
const TECH_RUBRIC: RubricPoint[][] = [
  [
    { label: '技术栈与架构交代', keys: ['架构', '技术栈', '框架', 'langchain', 'rag', 'python', 'pytorch', 'faiss', 'vllm', '检索', '向量'] },
    { label: '个人职责与模块', keys: ['负责', '模块', '设计', '实现', '搭建', '开发', '主导'] },
    { label: '难点与挑战', keys: ['挑战', '难点', '问题', '坑', '优化', '排查', '解决'] },
    { label: '量化成果', keys: ['%', '提升', '降低', 'qps', '延迟', '准确率', '覆盖率', '上线', '用户'] },
  ],
  [
    { label: 'Embedding 领域匹配', keys: ['embedding', '向量模型', '领域', 'bge', 'm3e', '微调模型'] },
    { label: 'Chunk 切分调优', keys: ['chunk', '切分', '分块', '粒度'] },
    { label: '混合检索（关键词+向量）', keys: ['混合检索', 'bm25', '关键词', 'hybrid', '稀疏'] },
    { label: 'Rerank 精排', keys: ['rerank', '重排', '精排'] },
    { label: '召回评估指标', keys: ['召回率', '命中率', 'hit', '评估', '指标', 'badcase'] },
  ],
  [
    { label: 'HNSW 图索引特性', keys: ['hnsw', '图索引', '图'] },
    { label: 'IVF 倒排聚类', keys: ['ivf', '聚类', '倒排'] },
    { label: '内存/延迟/召回权衡', keys: ['内存', '延迟', '召回率', '权衡', 'trade', '省'] },
    { label: 'Rerank 所处阶段', keys: ['rerank', '重排', '召回之后', '精排', '召回后'] },
  ],
  [
    { label: 'rank 过大的风险', keys: ['过拟合', '显存', '遗忘', '参数多'] },
    { label: 'rank 过小的局限', keys: ['表达能力', '欠拟合', '容量', '学不到'] },
    { label: '取值经验与验证', keys: ['8', '16', '32', '64', '验证集', '实验', '起步', '试'] },
  ],
  [
    { label: '连续批处理', keys: ['批处理', 'batch', 'continuous', '动态批'] },
    { label: 'KV cache 管理', keys: ['kv', 'pagedattention', '显存'] },
    { label: '量化推理', keys: ['量化', 'int8', 'int4', 'fp8', 'awq', 'gptq'] },
    { label: '弹性扩容', keys: ['扩容', '水平', '副本', '负载均衡', '弹性'] },
    { label: '缓存与 SLA 指标', keys: ['缓存', 'cache', 'p99', 'sla', '延迟要求', '吞吐'] },
  ],
]
const BEHAV_RUBRIC: RubricPoint[][] = [
  [
    { label: '倾听与理解对方', keys: ['倾听', '理解', '对方', '立场', '沟通'] },
    { label: '用事实/数据论证', keys: ['数据', '事实', '实验', '验证', 'benchmark', '对比'] },
    { label: '达成共识的方式', keys: ['共识', '妥协', '折中', '对齐', '讨论', '拍板'] },
    { label: '结果与反思', keys: ['结果', '最终', '效果', '复盘', '教训'] },
  ],
  [
    { label: '任务拆解与优先级', keys: ['拆解', '优先级', '拆分', '里程碑', '排期'] },
    { label: '风险沟通与同步', keys: ['沟通', '同步', '上报', '风险', '对齐'] },
    { label: '执行与按时交付', keys: ['赶工', '加班', '交付', '按时', '完成', '上线'] },
    { label: '事后复盘', keys: ['复盘', '改进', '总结', '教训'] },
  ],
  [
    { label: '学习动机', keys: ['因为', '为了', '兴趣', '需要', '驱动', '工作'] },
    { label: '系统化方法', keys: ['文档', '源码', '课程', '官方', '论文', '书'] },
    { label: '动手实践', keys: ['实践', '项目', 'demo', '练手', '落地', '动手'] },
    { label: '输出与沉淀', keys: ['输出', '分享', '博客', '笔记', '总结', '沉淀'] },
  ],
  [
    { label: '目标岗位认知', keys: ['ai', '工程师', '岗位', '方向', '目标', '大模型'] },
    { label: '能力差距分析', keys: ['差距', '短板', '不足', '提升', '弥补'] },
    { label: '具体行动计划', keys: ['计划', '学习', '半年', '一年', '规划', '步骤'] },
    { label: '长期发展愿景', keys: ['长期', '三年', '五年', '成长', '发展', '深耕'] },
  ],
]

function rubricFor(type: string, idx: number): RubricPoint[] | null {
  const r = type === 'technical' ? TECH_RUBRIC : BEHAV_RUBRIC
  return idx >= 0 && idx < r.length ? r[idx] : null
}

function makeNote(answer: string, points: RubricPoint[] | null): { note?: AnswerNote; score: number } {
  if (!points || !points.length) return { note: undefined, score: 70 }
  const a = (answer || '').toLowerCase()
  const covered: string[] = [], missed: string[] = []
  for (const p of points) {
    if (p.keys.some((k) => a.includes(k.toLowerCase()))) covered.push(p.label)
    else if (missed.length < 3) missed.push(p.label)
  }
  const tip = missed.length ? '可补充：' + missed.join('、') : '要点覆盖全面，保持这个回答结构。'
  const score = 55 + Math.round(45 * (covered.length / points.length))
  return { note: { covered, missed, tip }, score }
}

export function mockCreateInterview(type: string): InterviewSession {
  const sessionId = 's_' + Date.now()
  const script = buildScript(type)
  sessions[sessionId] = {
    type, script, transcript: [{ role: 'interviewer', content: script[0] }], idx: 0,
    awaiting: false, finished: false, reviews: [], scores: [], answers: [],
  }
  return { sessionId, type: type as InterviewSession['type'], startedAt: now() }
}

export function mockInterviewFirstMsg(sessionId: string): InterviewMsgResp {
  const s = sessions[sessionId]
  return { role: 'interviewer', content: s.transcript[0].content, turn: 0, finished: false }
}

export function mockSendInterview(sessionId: string, content: string): InterviewMsgResp {
  const s = sessions[sessionId]
  if (s.finished) return { role: 'interviewer', content: '本场已结束。', turn: s.idx, finished: true }
  if (content) {
    s.transcript.push({ role: 'user', content })
    s.answers.push(content)
  }
  const { note, score } = makeNote(content, rubricFor(s.type, s.idx))
  s.reviews.push({ turn: s.idx, question: s.script[s.idx], score, covered: note?.covered || [], missed: note?.missed || [] })
  s.scores.push(score)
  const fb = { score, hint: note?.tip || '' }

  // 答完第 2 题(idx=1)，下一题(idx=2)是深入考察 → 先 HITL 暂停，暂不透露
  if (s.type === 'technical' && s.idx === 1 && !s.awaiting) {
    s.awaiting = true
    return { role: 'interviewer', content: '', turn: s.idx, finished: false, awaitingApproval: true, approvalPrompt: '面试官想就此深入考察技术细节，是否继续？', interimFeedback: fb, note }
  }
  return advance(s, fb, note)
}

// resume：确认深入，揭示深入题
export function mockResumeInterview(sessionId: string): InterviewMsgResp {
  const s = sessions[sessionId]
  s.awaiting = false
  s.idx = DEEP_IDX
  const reply = s.script[s.idx]
  s.transcript.push({ role: 'interviewer', content: reply })
  return { role: 'interviewer', content: reply, turn: s.idx, finished: false, interimFeedback: { score: 74, hint: '继续展开。' } }
}

// skip：跳过深入题，直接到下一题
export function mockSkipInterview(sessionId: string): InterviewMsgResp {
  const s = sessions[sessionId]
  s.awaiting = false
  s.idx = DEEP_IDX // 把深入题当作已跳过，advance 到 idx=3
  return advance(s, undefined, undefined)
}

function advance(s: InterviewState, fb?: { score: number; hint: string }, note?: AnswerNote): InterviewMsgResp {
  const next = s.idx + 1
  if (next >= s.script.length) {
    s.finished = true
    s.transcript.push({ role: 'interviewer', content: '好的，本轮面试结束，稍后给你评估报告。' })
    persistInterviewMemory(s)
    const overall = s.scores.length ? Math.round(s.scores.reduce((a, x) => a + x, 0) / s.scores.length) : 0
    interviewHistory.push({ sessionId: sidOf(s), score: overall, at: now() })
    return { role: 'interviewer', content: '感谢参与！面试已结束。', turn: s.idx, finished: true, interimFeedback: { score: overall, hint: '报告已生成，去查看。' }, note }
  }
  s.idx = next
  const reply = s.script[s.idx]
  s.transcript.push({ role: 'interviewer', content: reply })
  return { role: 'interviewer', content: reply, turn: s.idx, finished: false, interimFeedback: fb, note }
}

function sidOf(s: InterviewState) { return Object.keys(sessions).find((k) => sessions[k] === s) || 's_0' }

function persistInterviewMemory(s: InterviewState) {
  const dim: DimensionId = s.type === 'technical' ? 'llm_app' : 'algo'
  state.memoryEvents.push({ type: 'interview', dimensionId: dim, delta: 3 })
  persist()
}

function reviewSummary(reviews: QuestionReview[], wantHighlight: boolean): string[] {
  const out: string[] = []
  for (const rv of reviews) {
    if (wantHighlight && rv.covered.length >= 2) out.push(`Q${rv.turn + 1} 覆盖：${rv.covered.join('、')}`)
    if (!wantHighlight && rv.missed.length) out.push(`Q${rv.turn + 1} 可补充：${rv.missed.join('、')}`)
    if (out.length === 3) break
  }
  if (!out.length) return wantHighlight ? ['全部题目均有作答，态度积极'] : ['各题要点覆盖全面，可挑战更高难度']
  return out
}

export function mockInterviewReport(sessionId: string): InterviewReport {
  const s = sessions[sessionId]
  const isTech = s.type === 'technical'
  const avg = (xs: number[]) => (xs.length ? Math.round(xs.reduce((a, x) => a + x, 0) / xs.length) : 0)
  const latter = s.scores.length > 2 ? s.scores.slice(-2) : s.scores
  const avgLen = s.answers.length ? Math.round(s.answers.join('').length / s.answers.length) : 0
  let structure = avgLen >= 40 ? 78 : avgLen >= 20 ? 70 : 62
  const starHits = ['首先', '然后', '接着', '最后', '结果'].filter((w) => s.answers.some((a) => a.includes(w))).length
  if (!isTech && starHits >= 2) structure = Math.min(85, structure + 6)
  const overall = avg(s.scores)
  return {
    sessionId, type: s.type as InterviewReport['type'], overallScore: overall,
    dimensions: isTech
      ? [{ name: '技术深度', score: avg(s.scores) }, { name: '系统设计', score: avg(latter) }, { name: '表达清晰度', score: structure }]
      : [{ name: '沟通协作', score: avg(s.scores) }, { name: '职业规划', score: avg(latter) }, { name: 'STAR 结构', score: structure }],
    highlights: reviewSummary(s.reviews, true),
    improvements: reviewSummary(s.reviews, false),
    transcript: s.transcript,
    review: s.reviews,
  }
}

// ---------- 5. 技术雷达 ----------
export function mockRadar(): RadarResp {
  const weak = mockGapAnalysis().recommendations.map((r) => r.dimensionId)
  const feed = [
    { id: 't1', title: 'vLLM 发布新调度优化，长上下文吞吐提升', source: 'GitHub', url: 'https://github.com/vllm-project/vllm', summary: '改进的 KV cache 管理显著提升长文本推理吞吐。', publishedAt: '2026-10-10', tags: ['推理', 'vLLM'], relevanceDimension: 'mlops' as DimensionId },
    { id: 't2', title: 'MCP 生态爆发：主流 Agent 框架全面支持', source: 'ModelContextProtocol', url: 'https://modelcontextprotocol.io/', summary: 'MCP 成为工具接入统一标准，Agent 平台标配。', publishedAt: '2026-10-12', tags: ['MCP', 'Agent'], relevanceDimension: 'llm_app' as DimensionId },
    { id: 't3', title: '开源模型 Qwen3 系列发布，长上下文能力增强', source: 'arXiv', url: 'https://arxiv.org/', summary: '新一代开源模型在推理与长文本上进步明显。', publishedAt: '2026-10-09', tags: ['开源模型'], relevanceDimension: 'dl_fundamentals' as DimensionId },
    { id: 't4', title: 'DPO/偏好对齐新方法降低训练成本', source: 'arXiv', url: 'https://arxiv.org/', summary: '无需显式奖励模型的对齐方法，适合资源受限微调。', publishedAt: '2026-10-08', tags: ['对齐', '微调'], relevanceDimension: 'finetune' as DimensionId },
    { id: 't5', title: 'GraphRAG 实践：图结构增强检索', source: 'GitHub', url: 'https://github.com/', summary: '以图组织知识，提升多跳问答召回。', publishedAt: '2026-10-11', tags: ['RAG'], relevanceDimension: 'llm_app' as DimensionId },
  ]
  const items = feed.sort((a, b) => (weak.includes(b.relevanceDimension) ? 1 : 0) - (weak.includes(a.relevanceDimension) ? 1 : 0))
  return { items, generatedAt: now() }
}

// ---------- 6. Dashboard ----------
export function mockDashboard(): DashboardSummary {
  const gap = mockGapAnalysis()
  return {
    overallScore: gap.overallScore,
    weakDimensions: gap.dimensions.filter((d) => d.status === 'weak' || d.status === 'missing').map((d) => d.id),
    recentExams: examHistory.slice(-3).reverse(),
    recentInterviews: interviewHistory.slice(-3).reverse(),
    studyStreakDays: 5,
    radarHighlights: mockRadar().items.slice(0, 3).map((i) => ({ title: i.title, url: i.url })),
  }
}
