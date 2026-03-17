import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/tests/setup.ts',
  },
  build: {
    rollupOptions: {
      onwarn(warning, warn) {
        // Suppress 'use client' directive warnings from Mantine
        if (warning.message?.includes('use client')) return;
        warn(warning);
      },
    },
  },
})
