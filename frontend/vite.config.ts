import {defineConfig, type Plugin} from 'vite'
import react from '@vitejs/plugin-react'

// The bound Go methods can delete and move files, so the packaged page may only
// run its own scripts. Emotion injects <style> tags, hence 'unsafe-inline' for
// styles. macOS serves the app from wails://, which 'self' may not cover.
// Production only: Vite's dev server relies on inline scripts and HMR sockets.
const csp = [
  "default-src 'self' wails:",
  "script-src 'self' wails:",
  "style-src 'self' wails: 'unsafe-inline'",
  "img-src 'self' wails: data: blob:",
  "font-src 'self' wails: data:",
  "connect-src 'self' wails:",
  "object-src 'none'",
  "base-uri 'none'",
  "form-action 'none'",
].join('; ')

const contentSecurityPolicy = (): Plugin => ({
  name: 'content-security-policy',
  apply: 'build',
  transformIndexHtml: () => [{tag: 'meta', attrs: {'http-equiv': 'Content-Security-Policy', content: csp}, injectTo: 'head-prepend'}],
})

// Büyük kütüphaneler ayrı parçalara bölünür; tek paket 500 kB uyarı sınırını aşmasın.
const vendorChunks: Record<string, string[]> = {
  react: ['react', 'react-dom', 'scheduler'],
  mui: ['@mui/material', '@mui/system', '@mui/utils', '@mui/styled-engine', '@mui/private-theming', '@mui/core-downloads-tracker', '@popperjs/core'],
  'mui-icons': ['@mui/icons-material'],
  emotion: ['@emotion'],
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), contentSecurityPolicy()],
  build: {
    // macOS 13 (Go 1.27's minimum) ships with Safari 16's WKWebView. Vite's default targets newer.
    target: 'safari16',
    cssTarget: 'safari16',
    rollupOptions: {
      output: {
        manualChunks(id) {
          const match = id.match(/node_modules[\\/]((?:@[^\\/]+[\\/])?[^\\/]+)/)
          if (!match) return
          const pkg = match[1].replace('\\', '/')
          for (const [chunk, pkgs] of Object.entries(vendorChunks)) {
            if (pkgs.some(p => pkg === p || pkg.startsWith(p + '/'))) return chunk
          }
          return 'vendor'
        },
      },
    },
  },
})
