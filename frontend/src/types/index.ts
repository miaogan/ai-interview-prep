// 与 docs/02-api-contract.md、docs/04-data-model.md 一一对应

export type DimensionId =
  | 'dl_fundamentals'
  | 'llm_app'
  | 'finetune'
  | 'mlops'
  | 'algo'

export type Status = 'strong' | 'moderate' | 'weak' | 'missing'

export interface Skill {
  name: string
  level: number // 1-5
  category: string
}

export interface Project {
  name: string
  techs: string[]
  desc: string
}

export interface Profile {
  name: string
  yearsExp: number
  targetRole: string
  skills: Skill[]
  projects: Project[]
  updatedAt?: string
}

export interface ResumeUploadResp {
  resumeId: string
  fileName: string
  parsedProfile: Profile
  uploadedAt: string
}

export interface Dimension {
  id: DimensionId
  name: string
  score: number
  weight: number
  status: Status
}

export interface Resource {
  title: string
  url: string
  type: string
}

export interface Recommendation {
  dimensionId: DimensionId
  priority: 'high' | 'medium' | 'low'
  topic: string
  resources: Resource[]
}

export interface GapReport {
  overallScore: number
  dimensions: Dimension[]
  recommendations: Recommendation[]
}

export type QuestionType = 'choice' | 'coding' | 'system_design'

export interface Question {
  id: string
  type: QuestionType
  dimensionId: DimensionId
  difficulty: 'easy' | 'medium' | 'hard'
  stem: string
  options?: string[]
  answer?: string | null
  explanation?: string | null
}

export interface Exam {
  examId: string
  title: string
  timeLimitSec: number
  questions: Question[]
}

export interface ResultItem {
  id: string
  yourAnswer: string
  correctAnswer: string
  isCorrect: boolean
  explanation: string
  dimensionId: DimensionId
}

export interface ExamResult {
  examId: string
  score: number
  total: number
  results: ResultItem[]
  weakPoints: DimensionId[]
  submittedAt: string
}

export interface ExamGenerateReq {
  dimensionIds: DimensionId[]
  count: number
  types: QuestionType[]
  difficulty: 'easy' | 'medium' | 'hard'
}

export type InterviewType = 'technical' | 'behavioral'

export interface InterviewSession {
  sessionId: string
  type: InterviewType
  startedAt: string
}

export interface Msg {
  role: 'interviewer' | 'user'
  content: string
}

export interface AnswerNote {
  covered: string[]
  missed: string[]
  tip: string
}

export interface QuestionReview {
  turn: number
  question: string
  score: number
  covered: string[]
  missed: string[]
}

export interface InterviewMsgResp {
  role: 'interviewer'
  content: string
  turn: number
  finished: boolean
  awaitingApproval?: boolean
  approvalPrompt?: string
  interimFeedback?: { score: number; hint: string }
  note?: AnswerNote
}

export interface InterviewReport {
  sessionId: string
  type: InterviewType
  overallScore: number
  dimensions: { name: string; score: number }[]
  highlights: string[]
  improvements: string[]
  transcript: Msg[]
  review: QuestionReview[]
}

export interface RadarItem {
  id: string
  title: string
  source: string
  url: string
  summary: string
  publishedAt: string
  tags: string[]
  relevanceDimension: DimensionId
}

export interface RadarResp {
  items: RadarItem[]
  generatedAt: string
}

export interface DashboardSummary {
  overallScore: number
  weakDimensions: DimensionId[]
  recentExams: { examId: string; score: number; at: string }[]
  recentInterviews: { sessionId: string; score: number; at: string }[]
  studyStreakDays: number
  radarHighlights: { title: string; url: string }[]
}

// 统一响应包裹
export interface ApiResponse<T> {
  code: number
  data: T
  message: string
}
