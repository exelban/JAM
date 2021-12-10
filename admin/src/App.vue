<template lang="pug">
header
  p Cheks &nbsp;
    a(href="https://github.com/exelban/cheks", target="_blank") {{ version }}

main
  service_(v-for="service in list", :service="service")
</template>

<script>
import "main.scss"
import service from "service"

export default {
  name: "app",
  components: {
    "service_": service
  },
  data: () => ({
    list: []
  }),
  computed: {
    version: () => process.env.VERSION
  },
  beforeCreate() {
    fetch("/list").then(res => res.json()).then((res) => {
      this.list = res
    }).catch((err) => {
      console.log(err)
    })
  }
}
</script>

<style lang="scss">
header {
  width: calc(100% - 28px);
  height: 30px;
  background: #111111;
  padding: 14px;
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;

  p {
    font-size: 18px;
    font-weight: bold;
    cursor: default;
    a {
      color: #969696;
      font-size: 11px;
      font-weight: normal;
      text-decoration: none;
      &:hover {
        text-decoration: underline;
      }
    }
  }
}

main {
  width: calc(100% - 20px);
  height: auto;
  padding: 10px;
}
</style>