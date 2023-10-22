<template>
  <footer class="p-medium row middle between">
    <a href="https://github.com/exelban/uptime">v0.0.0</a>
    <fa :icon="['fas', 'circle-half-stroke']" size="lg" class="pointer" uk-tooltip="Change theme" @click="toggleTheme"/>
  </footer>
</template>

<script>
import storage from "@/plugins/storage.js"

export default {
  name: "c-footer",
  data: () => ({
    theme: storage.get("theme") || "light"
  }),
  methods: {
    toggleTheme() {
      const newTheme = this.theme === "light" ? "dark" : "light"
      document.body.classList.remove(this.theme)
      this.theme = newTheme
      storage.set("theme", newTheme)
      document.body.classList.add(this.theme)
    }
  },
  mounted() {
    document.body.classList.add(this.theme)
  }
}
</script>

<style>
footer > a {
  color: #5e5e5e;
  text-decoration: none;
  font-weight: 600;
  font-size: 13px;
}
</style>