import { defineConfig } from "vite";

export default defineConfig({
  build: {
    outDir: "../internal/gui/webassets/dist",
    emptyOutDir: true,
    rollupOptions: {
      external: ["/wails/runtime.js"],
    },
  },
});
