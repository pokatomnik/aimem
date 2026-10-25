import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// The Go backend serves this SPA from the binary root (`/`), so every asset
// reference must stay relative: an absolute `/assets/...` URL would break when
// the UI is mounted under a sub-path later.
export default defineConfig({
  plugins: [react()],
  base: "./",
  resolve: {
    alias: {
      // Keep in sync with `compilerOptions.paths` in tsconfig.json.
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    port: 5173,
    // Dev only: talk to a locally running `aimem web` without CORS setup.
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8888",
        changeOrigin: false,
      },
    },
  },
});
