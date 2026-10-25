import { defineConfig, mergeConfig } from "vitest/config";

import viteConfig from "./vite.config";

// Tests are hermetic: jsdom provides a DOM, and every HTTP call goes through an
// injected fake `Api`. No network, no filesystem, no environment reads.
// The Vite config is merged in so the `@/` alias has a single source of truth.
export default mergeConfig(
  viteConfig,
  defineConfig({
    test: {
      environment: "jsdom",
      globals: true,
      setupFiles: ["./src/test/setup.ts"],
      restoreMocks: true,
      clearMocks: true,
      mockReset: true,
      include: ["src/**/*.test.{ts,tsx}"],
      css: false,
    },
  }),
);
