<template>
  <tbody class="check">
  <tr>
    <td style="cursor: pointer;" @click="toggleDetails()">
      <fa v-if="value.status.value === 'up'" :icon="['fas', 'circle-check']" size="lg" style="color: var(--up-color);" class="uk-preserve-width" uk-tooltip="UP"/>
      <fa v-else-if="value.status.value === 'down'" :icon="['fas', 'circle-xmark']" size="lg" style="color: var(--down-color);" class="uk-preserve-width" uk-tooltip="Down"/>
    </td>
    <td style="cursor: pointer;" @click="toggleDetails()">
      <div class="column center">
        <span class="name">{{ value.name }}</span>
        <span class="host">{{ value.host }}</span>
      </div>
    </td>
    <td style="cursor: pointer;" @click="toggleDetails()"><span class="availability">100%</span></td>
    <td style="padding:0;">
      <div :style="{height: chartHeight}">
        <Line :data="data" :options="options"/>
      </div>
    </td>
    <td ref="tags">
      <div class="row wrap right">
        <span class="uk-label tag" uk-tooltip="Filter by tag" v-for="t in value.tags" :style="{backgroundColor: t.color}" @click="filterByTag(t)">{{t.name}}</span>
      </div>
    </td>
  </tr>
  <tr>
    <td colspan="5" style="padding: 4px 10px;">
      <div :ref="value.id" class="statusChart column wrap border-rounded">
        <span v-for="p in uptime" :class="p.status" :uk-tooltip="p.status"></span>
      </div>
    </td>
  </tr>
  <Transition>
    <tr v-if="details">
      <td colspan="5" style="background: var(--background-color)">
        <div class="details column">
          <div class="box shadow-normal border-rounded p-medium row middle between">
            <dl class="uk-description-list">
              <dt>Status</dt>
              <dd class="status" :class="value.status.value">{{ value.status.value }}</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Name</dt>
              <dd>{{ value.name }}</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Host</dt>
              <dd>{{ value.host }}</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Availability</dt>
              <dd>100%</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Downtime</dt>
              <dd>0%</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Avg response time</dt>
              <dd>528 ms</dd>
            </dl>
          </div>
<!--          <div class="box p-small ml-5">-->
<!--            <h3>Availability & Performance</h3>-->
<!--            <p>Lorem ipsum...</p>-->
<!--          </div>-->
        </div>
      </td>
    </tr>
  </Transition>
  </tbody>
</template>

<script>
import {Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip} from "chart.js"
import { Line } from "vue-chartjs"

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip)

export default {
  props: {
    value: Object
  },
  components: {
    Line
  },
  data: () => ({
    chartWidth: 0,
    chartHeight: "40px",
    data: {
      labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
      datasets: [
        {
          data: [40, 39, 10, 40, 39, 80, 40],
          pointRadius: 0,
          lineTension: 0.1,
          fill: true,
          borderColor: '#0a8bf1',
          backgroundColor: "rgba(10, 139, 241, 0.2)",
          borderWidth: 1
        }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false
        },
        tooltip: {
          usePointStyle: true,
          displayColors: false,
          callbacks: {
            label: function (ctx) {
              return `${ctx.raw} ms`
            }
          }
        }
      },
      interaction: {
        intersect: false,
        mode: "nearest",
        axis: "x"
      },
      scales: {
        x: {
          display: false
        },
        y: {
          border: {
            display: false
          },
          grid: {
            display: false
          },
          ticks: {
            color: "#8c8c8c",
            font: {
              size: 10
            },
            callback: function (value, index, values) {
              if (index === 0) {
                return this.min
              } else if (index === values.length - 1) {
                return this.max
              }
              return ""
            }
          },
        }
      },
    },
    details: false,
  }),
  computed: {
    uptime() {
      let list = this.value.checks
      let max = Math.floor(this.chartWidth / 13)
      if (list.length > max) {
        list = list.slice(list.length - max)
      } else if (list.length < max) {
        let diff = max - list.length
        for (let i = 0; i < diff; i++) {
          list.unshift({})
        }
      }
      return list
    }
  },
  methods: {
    filterByTag(tag) {
      this.$emit("filter-by-tag", tag)
    },
    toggleDetails() {
      this.details = !this.details
    },
    resizeCallback() {
      if (this.$refs.tags) {
        this.chartHeight = this.$refs.tags.offsetHeight - 8 + "px"
      }
      if (this.$refs[this.value.id]) {
        this.chartWidth = this.$refs[this.value.id].offsetWidth
      }
    }
  },
  mounted() {
    this.resizeCallback()
    window.addEventListener("resize", this.resizeCallback)
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.resizeCallback)
  },
}
</script>

<style lang="scss">
@import "@/style.scss";

.check {
  transition: background-color 0.1s ease-in-out;
  &:hover {
    background: var(--hover-color);
  }

  .host {
    color: #7e7e7e;
  }
  .availability {
    font-weight: 600;
  }
  .statusChart {
    height: 12px;
    padding: 0 0 2px 0;
    overflow: hidden;
    span {
      width: 12px;
      height: 100%;
      background: rgba(227, 227, 227, 0.5);
      border-radius: 2px;
      margin: 0 0 0 1px;
      cursor: default;
      &.up {
        background: var(--up-color);
      }
      &.down {
        background: var(--down-color);
      }
    }
  }

  .details {
    .box {
      background: var(--default-color);

      dl {
        margin: 0;
        padding: 0 8px;

        dt {
          color: #7e7e7e;
          font-size: 12px;
          margin: 0;
        }
        dd {
          font-size: 14px;
          font-weight: 600;
          margin: 0;
          &.status {
            text-transform: uppercase;
            &.up {
              color: var(--up-color);
            }
            &.down {
              color: var(--down-color);
            }
          }
        }
      }
    }
  }

  .v-enter-active, .v-leave-active {
    transition: opacity 0.2s ease;
  }
  .v-enter-from, .v-leave-to {
    opacity: 0;
    transition: opacity 0.1s ease;
  }
}
</style>