import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    watch: {
      // 本机 chokidar 默认监听抓不到外部工具的原子写（临时文件+rename），
      // 导致 HMR 不失效、dev server 提供过期模块；轮询兜底。
      usePolling: true,
      interval: 300,
    },
  },
})
