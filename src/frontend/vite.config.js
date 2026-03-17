import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173, // 固定前端端口
    proxy: {
      // 配置跨域代理
      '/api': {
        target: 'http://localhost:8080', // 你的 Go 后端服务地址
        changeOrigin: true,
      }
    }
  }
})