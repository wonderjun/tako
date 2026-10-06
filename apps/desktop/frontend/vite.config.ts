import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';
import wails from '@wailsio/runtime/plugins/vite';

// Wails 资产服务约定：dev 由 wails3 注入 WAILS_VITE_PORT（Taskfile 默认 9245），
// 生产经 go:embed 以相对路径加载。
export default defineConfig({
  base: './',
  server: {
    host: '127.0.0.1',
    port: Number(process.env.WAILS_VITE_PORT) || 9245,
    strictPort: true,
  },
  plugins: [vue(), tailwindcss(), wails('./bindings')],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@contracts': fileURLToPath(new URL('../../../packages/contracts/src', import.meta.url)),
    },
  },
  build: {
    // vite 8 默认 minifier（oxc）；浮窗场景追求启动速度，不产 sourcemap
  },
});
