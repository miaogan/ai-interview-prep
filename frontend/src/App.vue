<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { useAppStore } from './stores/app'

const store = useAppStore()
onMounted(() => { store.fetchProfile(); store.fetchDashboard() })

const nav = [
  { to: '/', label: '总览' },
  { to: '/resume', label: '简历' },
  { to: '/gap', label: '查漏补缺' },
  { to: '/exam', label: '笔试' },
  { to: '/interview', label: '模拟面试' },
  { to: '/radar', label: '技术雷达' },
]
</script>

<template>
  <header class="topbar">
    <div class="brand"><span class="logo">◆</span> AI 面霸<span class="muted small"> · Interview Prep</span></div>
    <nav class="nav">
      <RouterLink v-for="n in nav" :key="n.to" :to="n.to" class="navlink">{{ n.label }}</RouterLink>
    </nav>
  </header>
  <main><RouterView /></main>
</template>

<style scoped>
.topbar {
  display: flex; align-items: center; justify-content: space-between;
  gap: 16px; padding: 14px 20px; border-bottom: 1px solid var(--border);
  position: sticky; top: 0; background: rgba(15,17,21,.9); backdrop-filter: blur(8px); z-index: 10;
  flex-wrap: wrap;
}
.brand { font-weight: 700; font-size: 17px; }
.logo { color: var(--brand); }
.small { font-size: 13px; font-weight: 400; }
.nav { display: flex; gap: 6px; flex-wrap: wrap; }
.navlink {
  padding: 7px 12px; border-radius: 8px; color: var(--muted); font-size: 14px; font-weight: 600;
  transition: all .15s;
}
.navlink:hover { color: var(--text); background: var(--panel-2); }
.navlink.router-link-active { color: var(--text); background: var(--panel-2); }
</style>
