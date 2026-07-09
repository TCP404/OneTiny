import { defineConfig } from "vite";

export default defineConfig({
  publicDir: "../resource/logo",
  build: {
    outDir: "../internal/gui/webassets/dist",
    emptyOutDir: true,
    rollupOptions: {
      external: ["/wails/runtime.js"],
    },
  },
});
