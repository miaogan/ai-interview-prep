import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'
import type { Profile, GapReport, DashboardSummary, RadarResp, StudyPlan } from '../types'

export const useAppStore = defineStore('app', () => {
  const profile = ref<Profile | null>(null)
  const gap = ref<GapReport | null>(null)
  const dashboard = ref<DashboardSummary | null>(null)
  const radar = ref<RadarResp | null>(null)
  const plan = ref<StudyPlan | null>(null)
  const planLoading = ref(false)
  const loading = ref(false)

  const hasProfile = computed(() => !!profile.value)
  const overallScore = computed(() => gap.value?.overallScore ?? null)

  async function fetchProfile() {
    const r = await api.getProfile()
    profile.value = r.data
  }
  async function uploadResume(file: File) {
    loading.value = true
    try {
      const r = await api.uploadResume(file)
      profile.value = r.data.parsedProfile
      await refreshAnalytics()
      return r.data
    } finally { loading.value = false }
  }
  async function fetchGap() {
    loading.value = true
    try { const r = await api.gapAnalysis(); gap.value = r.data }
    finally { loading.value = false }
  }
  async function fetchRadar() { const r = await api.radar(); radar.value = r.data }
  async function fetchDashboard() { const r = await api.dashboard(); dashboard.value = r.data }
  async function fetchPlan() { const r = await api.getPlan(); plan.value = r.data }
  async function generatePlan() {
    planLoading.value = true
    try { const r = await api.generatePlan(); plan.value = r.data }
    finally { planLoading.value = false }
  }

  async function refreshAnalytics() {
    await Promise.all([fetchGap(), fetchDashboard()])
  }

  return { profile, gap, dashboard, radar, plan, planLoading, loading, hasProfile, overallScore,
    uploadResume, fetchProfile, fetchGap, fetchRadar, fetchDashboard, fetchPlan, generatePlan, refreshAnalytics }
})
