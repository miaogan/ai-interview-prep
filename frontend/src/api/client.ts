// API 适配层：VITE_USE_MOCK=true 走 mock，否则走真实后端（契约一致，后端接入零改动）
import axios from 'axios'
import type {
  Profile, ResumeUploadResp, GapReport, Exam, ExamGenerateReq, ExamResult, StudyPlan,
  InterviewSession, InterviewMsgResp, InterviewReport, RadarResp, DashboardSummary, ApiResponse,
} from '../types'
import * as mock from './mock'

const USE_MOCK = import.meta.env.VITE_USE_MOCK !== 'false'
const http = axios.create({ baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1', timeout: 60000 })

// mock 延迟，模拟网络、驱动 loading 态
const delay = <T>(v: T, ms = 500): Promise<ApiResponse<T>> =>
  new Promise((res) => setTimeout(() => res({ code: 0, data: v, message: 'ok' }), ms))

async function get<T>(url: string, params?: any): Promise<ApiResponse<T>> {
  const r = await http.get<ApiResponse<T>>(url, { params })
  return r.data
}
async function post<T>(url: string, body?: any): Promise<ApiResponse<T>> {
  const r = await http.post<ApiResponse<T>>(url, body)
  return r.data
}

export const api = {
  // 简历
  uploadResume(file: File): Promise<ApiResponse<ResumeUploadResp>> {
    if (USE_MOCK) return delay(mock.mockUploadResume(file.name))
    const fd = new FormData(); fd.append('file', file)
    return http.post<ApiResponse<ResumeUploadResp>>('/resume/upload', fd, { timeout: 120000 }).then((r) => r.data)
  },
  getProfile(): Promise<ApiResponse<Profile | null>> {
    if (USE_MOCK) return delay(mock.mockGetProfile(), 200)
    return get<Profile | null>('/resume/profile')
  },

  // 查漏补缺
  gapAnalysis(): Promise<ApiResponse<GapReport>> {
    if (USE_MOCK) return delay(mock.mockGapAnalysis())
    return get<GapReport>('/gap/analysis')
  },

  // 学习路径
  generatePlan(): Promise<ApiResponse<StudyPlan>> {
    if (USE_MOCK) return delay(mock.mockGeneratePlan(), 1200)
    return post<StudyPlan>('/plan/generate')
  },
  getPlan(): Promise<ApiResponse<StudyPlan | null>> {
    if (USE_MOCK) return delay(mock.mockGetPlan(), 200)
    return get<StudyPlan | null>('/plan')
  },

  // 笔试
  generateExam(req: ExamGenerateReq): Promise<ApiResponse<Exam>> {
    if (USE_MOCK) return delay(mock.mockGenerateExamAndCache(req))
    return post<Exam>('/exam/generate', req)
  },
  submitExam(examId: string, answers: Record<string, string>): Promise<ApiResponse<ExamResult>> {
    if (USE_MOCK) return delay(mock.mockSubmitExam(examId, answers), 800)
    return post<ExamResult>(`/exam/${examId}/submit`, { answers })
  },

  // 面试
  createInterview(type: string): Promise<ApiResponse<InterviewSession>> {
    if (USE_MOCK) return delay(mock.mockCreateInterview(type))
    return post<InterviewSession>('/interview/session', { type, rounds: 8 })
  },
  interviewFirstMsg(sessionId: string): Promise<ApiResponse<InterviewMsgResp>> {
    if (USE_MOCK) return delay(mock.mockInterviewFirstMsg(sessionId), 300)
    return post<InterviewMsgResp>(`/interview/${sessionId}/message`, { role: 'user', content: '__START__' })
  },
  sendInterview(sessionId: string, content: string): Promise<ApiResponse<InterviewMsgResp>> {
    if (USE_MOCK) return delay(mock.mockSendInterview(sessionId, content), 900)
    return post<InterviewMsgResp>(`/interview/${sessionId}/message`, { role: 'user', content })
  },
  resumeInterview(sessionId: string): Promise<ApiResponse<InterviewMsgResp>> {
    if (USE_MOCK) return delay(mock.mockResumeInterview(sessionId), 600)
    return post<InterviewMsgResp>(`/interview/${sessionId}/resume`)
  },
  skipInterview(sessionId: string): Promise<ApiResponse<InterviewMsgResp>> {
    if (USE_MOCK) return delay(mock.mockSkipInterview(sessionId), 500)
    return post<InterviewMsgResp>(`/interview/${sessionId}/skip`)
  },
  interviewReport(sessionId: string): Promise<ApiResponse<InterviewReport>> {
    if (USE_MOCK) return delay(mock.mockInterviewReport(sessionId))
    return get<InterviewReport>(`/interview/${sessionId}/report`)
  },

  // 技术雷达
  radar(): Promise<ApiResponse<RadarResp>> {
    if (USE_MOCK) return delay(mock.mockRadar())
    return get<RadarResp>('/radar/latest', { topic: 'llm', limit: 10 })
  },

  // Dashboard
  dashboard(): Promise<ApiResponse<DashboardSummary>> {
    if (USE_MOCK) return delay(mock.mockDashboard())
    return get<DashboardSummary>('/dashboard/summary')
  },
}

export { USE_MOCK }
