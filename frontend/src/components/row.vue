<template>
  <tbody class="check">
  <tr :ref="target.id">
    <td style="cursor: pointer;" @click="toggleDetails()">
      <fa v-if="target.status.value === 'up'" :icon="['fas', 'circle-check']" size="lg" style="color: var(--up-color);" class="uk-preserve-width" uk-tooltip="UP"/>
      <fa v-else-if="target.status.value === 'down'" :icon="['fas', 'circle-xmark']" size="lg" style="color: var(--down-color);" class="uk-preserve-width" uk-tooltip="Down"/>
      <fa v-else :icon="['fas', 'circle-question']" size="lg" style="color: var(--unknown-color);" class="uk-preserve-width" uk-tooltip="Unknown"/>
    </td>
    <td style="cursor: pointer;" @click="toggleDetails()">
      <span class="name">{{ target.name }}</span>
    </td>
    <td style="cursor: pointer;" @click="toggleDetails()"><span class="availability">{{availability}}%</span></td>
    <td style="padding:0;" class="responseTimeColumn">
      <div :style="{height: chartHeight}">
        <Line :data="responseTime" :options="options"/>
      </div>
    </td>
    <td ref="tags">
      <div class="row wrap right">
        <span class="uk-label tag" uk-tooltip="Filter by tag" v-for="t in target.tags" :style="{backgroundColor: t.color}" @click="filterByTag(t)">{{t.name}}</span>
      </div>
    </td>
  </tr>
  <Transition>
    <tr v-if="details">
      <td colspan="5" style="background: var(--background-color)">
        <div class="details column">
          <div class="box shadow-normal border-rounded p-4 row middle between">
            <div class="statusChart row wrap">
              <span v-for="p in uptime" :class="p.status === undefined ? '' : p.status ? 'up' : 'down'" :uk-tooltip="p.status === undefined ? 'Unknown' : p.status ? 'UP' : 'Down'"></span>
            </div>
          </div>
        </div>
        <div class="details column">
          <div class="box shadow-normal border-rounded p-medium row middle between">
            <dl class="uk-description-list">
              <dt>Status</dt>
              <dd class="status" :class="target.status.value">{{ target.status.value }}</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Name</dt>
              <dd>{{ target.name }}</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Availability</dt>
              <dd>{{availability}}%</dd>
            </dl>
            <dl class="uk-description-list ml-5">
              <dt>Downtime</dt>
              <dd>{{downtime}}%</dd>
            </dl>
            <dl class="uk-description-list ml-5 responseTimeColumn">
              <dt>Avg response time</dt>
              <dd>{{averageResponseTime}} ms</dd>
            </dl>
          </div>
        </div>
        <div class="details column">
          <div class="box shadow-normal border-rounded p-medium row middle">
            <div class="w h information p-2">
              <span>Retry</span>
              <span>{{target.retry}}</span>
              <span>Timeout</span>
              <span>{{target.timeout}}</span>
              <span>Initial delay</span>
              <span>{{target.initialDelay}}</span>
              <span>Success threshold</span>
              <span>{{target.successThreshold}}</span>
              <span>Failure threshold</span>
              <span>{{target.failureThreshold}}</span>
              <span>Success</span>
              <span>{{target.success ?? "default"}}</span>
              <span>History</span>
              <span>{{target.history}}</span>
              <span>Headers</span>
              <span>{{target.headers ?? "none"}}</span>
            </div>
          </div>
        </div>
        <div class="details column">
          <div class="box shadow-normal border-rounded p-3 row middle">
            <button class="uk-button uk-button-small" style="background: var(--down-color);color: var(--secondary-color);width: 25%;" type="button" @click="deleteTarget">Delete</button>
            <button class="uk-button uk-button-small" style="background: var(--primary-color);color: var(--secondary-color);flex-grow: 1;" type="button" @click="editTarget">Edit</button>
          </div>
        </div>
      </td>
    </tr>
  </Transition>
  </tbody>
</template>

<script>
import {Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip} from "chart.js"
import { Line } from "vue-chartjs"
import UIkit from "uikit"

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip)

export default {
  props: {
    target: Object
  },
  components: {
    Line
  },
  data: () => ({
    chartWidth: 0,
    chartHeight: "40px",
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: {
        duration: 0
      },
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
    responseTime() {
      const checks = this.target.checks.slice(0, 10)
      const labels = checks.map(c => new Date(c.timestamp).toLocaleString())
      const times = checks.map(c => c.time)
      return {
        labels: labels,
        datasets: [
          {
            data: times,
            pointRadius: 0,
            lineTension: 0.1,
            fill: true,
            borderColor: '#0a8bf1',
            backgroundColor: "rgba(10, 139, 241, 0.2)",
            borderWidth: 1
          }
        ]
      }
    },
    averageResponseTime() {
      if (this.target.checks.length === 0) {
        return 0
      }
      return Math.floor(this.target.checks.map(c => c.time).reduce( ( p, c ) => p + c, 0 ) / this.target.checks.length)
    },
    uptime() {
      let list = [...this.target.checks]
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
    },
    availability() {
      if (this.target.checks.length === 0) {
        return 0
      }
      const checks = this.target.checks.filter(c => c.status).length
      return parseInt(checks*100/this.target.checks.length)
    },
    downtime() {
      const checks = this.target.checks.filter(c => !c.status).length
      return parseInt(checks*100/this.target.checks.length)
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
        this.chartHeight = this.$refs.tags.offsetHeight - 1 + "px"
      }
      if (this.$refs[this.target.id]) {
        this.chartWidth = this.$refs[this.target.id].offsetWidth - 24
      }
    },
    editTarget() {
      this.$store.commit("setForm", this.target)
      UIkit.modal(document.getElementById("edit-target-dialog")).show()
    },
    deleteTarget() {
      this.$store.commit("setTarget", this.target)
      UIkit.modal(document.getElementById("delete-target-dialog")).show()
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

  .details {
    margin: 10px 0 0 0;
    &:first-child {
      margin: 0;
    }

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

      .information {
        display: grid;
        grid-template-columns: 1fr 3fr;
        align-items: start;
        gap: 4px;

        span:nth-child(2n) {
          font-weight: 600;
        }
      }
    }

    .statusChart {
      height: 12px;
      padding: 0;
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