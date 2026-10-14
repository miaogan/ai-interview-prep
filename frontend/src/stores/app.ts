import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '../api/client'
import type { Profile, GapReport, DashboardSummary, RadarResp } from '../types'

export const useAppStore = defineStore('app', () => {
  const profile = ref<Profile | null>(null)
  const gap = ref<GapReport | null>(null)
  const dashboard = ref<DashboardSummary | null>(null)
  const radar = ref<RadarResp | null>(null)
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

  async function refreshAnalytics() {
    await Promise.all([fetchGap(), fetchDashboard()])
  }

  return { profile, gap, dashboard, radar, loading, hasProfile, overallScore,
    uploadResume, fetchProfile, fetchGap, fetchRadar, fetchDashboard, refreshAnalytics }
})
