import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'dashboard', component: () => import('../views/DashboardView.vue'), meta: { title: '总览' } },
  { path: '/resume', name: 'resume', component: () => import('../views/ResumeView.vue'), meta: { title: '简历' } },
  { path: '/gap', name: 'gap', component: () => import('../views/GapView.vue'), meta: { title: '查漏补缺' } },
  { path: '/exam', name: 'exam', component: () => import('../views/ExamView.vue'), meta: { title: '笔试练习' } },
  { path: '/interview', name: 'interview', component: () => import('../views/InterviewView.vue'), meta: { title: '模拟面试' } },
  { path: '/radar', name: 'radar', component: () => import('../views/RadarView.vue'), meta: { title: '技术雷达' } },
]

export default createRouter({ history: createWebHistory(), routes })
