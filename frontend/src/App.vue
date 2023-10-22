<template>
  <div class="app w h column">
    <c-list/>
    <c-footer/>
  </div>
</template>

<script>
import footer from "@/components/footer.vue"
import list from "@/components/list.vue"

export default {
  name: "app",
  components: {"c-footer": footer, "c-list": list},
}
</script>

<style lang="scss">
@import "style.scss";
</style>