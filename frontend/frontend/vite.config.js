import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';
import { resolve } from 'path';

export default defineConfig({
  plugins: [react()],
  server: {
    fs: {
      strict: false
    }
  },
  build: {
    rollupOptions: {
      input: resolve(__dirname, 'index.html')
    }
  }
});
