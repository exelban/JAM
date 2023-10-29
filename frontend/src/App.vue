<template>
  <div class="app w h column">
    <c-list :services="services"/>
    <c-footer/>
  </div>
</template>

<script>
import footer from "@/components/footer.vue"
import list from "@/components/list.vue"

export default {
  name: "app",
  components: {"c-footer": footer, "c-list": list},
  data: () => ({
    services: [],
  }),
  methods: {
    load() {
      fetch("http://localhost:8080/list").then(res => res.json()).then(data => {
        this.services = data
      }).catch(err => {
        console.log(err)
      })
    }
  },
  beforeCreate() {
    this.interval = setInterval(() => {
      this.load()
    }, 1000)
  },
  beforeMount() {
    this.load()
  },
  beforeUnmount() {
    clearInterval(this.interval)
  }
}
</script>

<style lang="scss">
@import "style.scss";
</style>