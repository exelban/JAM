<template lang="pug">
.service(:title="this.service.name")
  .row(v-on:click="opened = !opened")
    .info
      i(:class="'status '+this.service.status.value")
      p.name {{this.service.name}}
    .tags
      p(v-for="tag in service.tags", :title="tag.name", :style="{'background-color': tag.color}") {{tag.name}}
  transition(name="slide")
    .details(v-if="opened")
      .stats
        p Last status change: {{ timestamp(this.service.status.timestamp) }}
        p Success count: {{ this.service.success.length }}
        p Failure count: {{ this.service.failure.length }}
      ul.chart
        li(v-for="check in this.service.checks")
          span(
            :class="check.status ? 'up' : 'down'",
            :style="{'height': check.status ? '100%' : '50%'}",
            :title="check.body + 'respond with status:' + check.code"
          )
</template>

<script>
export default {
  name: "service_",
  props: {
    service: {
      type: Object
    }
  },
  data: () => ({
    opened: false,
  }),
  methods: {
    timestamp: (v) => new Date(v).toLocaleString()
  }
}
</script>

<style lang="scss">
.service {
  width: 100%;
  height: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  background: #313131;
  border-radius: 2px;
  margin-top: 10px;
  box-shadow: rgba(0, 0, 0, 0.12) 0 1px 3px, rgba(0, 0, 0, 0.24) 0 1px 2px;

  .unknown {
    background: #dddddd;
  }
  .up {
    background: #35ea0e;
  }
  .down {
    background: #fc0000;
  }

  .row {
    width: calc(100% - 20px);
    height: 24px;
    padding: 10px;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
    pointer-events: initial;
    .info {
      width: auto;
      height: auto;
      display: flex;
      flex-direction: row;
      align-items: center;
      .status {
        width: 12px;
        height: 12px;
        border-radius: 7px;
        margin: 0;
        padding: 0;
      }
      p.name {
        width: auto;
        height: auto;
        margin: 0 0 0 10px;
        font-size: 14px;
      }
    }
    .tags {
      width: auto;
      height: 100%;
      display: flex;
      flex-direction: row;
      align-items: center;
      p {
        width: auto;
        height: auto;
        border-radius: 10px;
        margin: 0 3px;
        padding: 3px 5px;
        display: flex;
        align-items: center;
      }
    }
  }
  .details {
    width: calc(100% - 20px);
    padding: 0 10px 10px 10px;
    .stats {
      width: calc(100% - 12px);
      display: flex;
      flex-direction: column;
      align-items: start;
      padding: 0 6px 6px 6px;
      p {
        margin: 0;
        cursor: default;
      }
    }
    .chart {
      width: 100%;
      height: 120px;
      display: table;
      table-layout: fixed;
      margin: 0;
      padding: 0;
      background: #111111;
      li {
        position: relative;
        display: table-cell;
        vertical-align: bottom;
        height: 100%;
      }
      span {
        margin: 0 2px;
        display: block;
      }
    }
  }

  &:first-child {
    margin-top: 0;
  }
  &:nth-child(2n) {
    background: #383838;
  }

  .slide-enter-active {
    -moz-transition-duration: 0.3s;
    -webkit-transition-duration: 0.3s;
    -o-transition-duration: 0.3s;
    transition-duration: 0.3s;
    -moz-transition-timing-function: ease-in;
    -webkit-transition-timing-function: ease-in;
    -o-transition-timing-function: ease-in;
    transition-timing-function: ease-in;
  }
  .slide-leave-active {
    -moz-transition-duration: 0.3s;
    -webkit-transition-duration: 0.3s;
    -o-transition-duration: 0.3s;
    transition-duration: 0.3s;
    -moz-transition-timing-function: cubic-bezier(0, 1, 0.5, 1);
    -webkit-transition-timing-function: cubic-bezier(0, 1, 0.5, 1);
    -o-transition-timing-function: cubic-bezier(0, 1, 0.5, 1);
    transition-timing-function: cubic-bezier(0, 1, 0.5, 1);
  }
  .slide-enter-to, .slide-leave-from {
    max-height: 1000px;
    overflow: hidden;
    opacity: 1;
  }
  .slide-enter-from, .slide-leave-to {
    overflow: hidden;
    max-height: 0;
    opacity: 0;
  }
}
</style>