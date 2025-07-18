import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  build: {
    outDir: '.build',
    assetsDir: 'static',
    emptyOutDir: true,
    minify: 'false',
    sourcemap: 'true',
  },
  plugins: [react()],
})
