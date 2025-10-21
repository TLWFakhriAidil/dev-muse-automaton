import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";
import { componentTagger } from "lovable-tagger";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
  plugins: [
    react(),
    mode === 'development' && componentTagger(),
  ].filter(Boolean),
  // Ensure assets are served from /assets/ directory
  base: '/',
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    // Generate manifest for asset tracking
    manifest: false,
    // Ensure consistent hashing
    rollupOptions: {
      output: {
        // Ensure JS and CSS go to assets/ directory
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    }
  },
  server: {
    host: "::",
    port: 8080,
    allowedHosts: ["nodepath-chat-production.up.railway.app"],
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false
      }
    }
  },
  preview: {
    host: "0.0.0.0",
    port: 4173,
    allowedHosts: ["nodepath-chat-production.up.railway.app"]
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
}));
