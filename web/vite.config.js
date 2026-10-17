import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  build: { outDir: 'dist', emptyOutDir: true, chunkSizeWarningLimit: 900 },
  server: {
    // npm run dev: a API continua no `iz ui` (porta 8080)
    proxy: { '/api': { target: 'http://127.0.0.1:8080', ws: true } }
  }
});
