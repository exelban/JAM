import { defineConfig } from "vite"
import vue from "@vitejs/plugin-vue"

import { fileURLToPath, URL } from "node:url"

export default defineConfig({
  build: {
    target: 'esnext'
  },
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
      "vue-i18n": "vue-i18n/dist/vue-i18n.cjs.js"
    },
  }
})
