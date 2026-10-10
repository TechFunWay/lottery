import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'lottery-web',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        assetFileNames: (assetInfo) => {
          // 将js/css等静态资源放到lottery-web目录
          if (/\.(css)$/.test(assetInfo.name || '')) {
            return `lottery-web/[name]-[hash][extname]`
          }
          if (/\.(js)$/.test(assetInfo.name || '')) {
            return `lottery-web/[name]-[hash][extname]`
          }
          return `lottery-web/[name]-[hash][extname]`
        },
        chunkFileNames: 'lottery-web/[name]-[hash].js',
        entryFileNames: 'lottery-web/[name]-[hash].js'
      }
    }
  },
  server: {
    host: '0.0.0.0',
    port: 5176,
    allowedHosts: true
  }
})
