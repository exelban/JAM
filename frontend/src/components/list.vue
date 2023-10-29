<template>
  <main class="p-small">
    <div class="filters row middle">
      <fa :icon="['fas', 'filter']" class="mh-5"/>
      <span class="title">Filter: </span>
      <button class="uk-button shadow-normal border-rounded uk-button-small ml-5" style="background: var(--default-color); color: var(--text-color)" type="button">
        {{ !filter.status ? "Status" : "" }}
        <div v-if="filter.status==='up'" class="row center middle">
          <fa :icon="['fas', 'circle-check']" style="color: #77bb41;"/>
          <span class="ml-5">Up</span>
        </div>
        <div v-else-if="filter.status==='down'" class="row center middle">
          <fa :icon="['fas', 'circle-xmark']" style="color: #f85c5c;"/>
          <span class="ml-5">Down</span>
        </div>
      </button>
      <div uk-dropdown="mode: click" class="statuses border-rounded shadow-normal">
        <div class="column">
          <div class="status border-rounded row middle between" :class="{'active': filter.status==='up'}" @click="filterByStatus('up')">
            <div class="row middle center">
              <fa :icon="['fas', 'circle-check']" style="color: var(--up-color);"/>
              <span class="ml-5">Up</span>
              <span style="margin-left: 4px;">({{upCount}})</span>
            </div>
            <fa class="active-mark" :icon="['fas', 'check']"/>
          </div>
          <div class="status border-rounded row middle between" :class="{'active': filter.status==='down'}" @click="filterByStatus('down')">
            <div class="row middle center">
              <fa :icon="['fas', 'circle-xmark']" style="color: var(--down-color);"/>
              <span class="ml-5">Down</span>
              <span style="margin-left: 4px;">({{downCount}})</span>
            </div>
            <fa class="active-mark" :icon="['fas', 'check']"/>
          </div>
        </div>
      </div>

      <div class="row ml-5">
          <span class="uk-label tag row center middle" v-for="t in filter.tags" uk-tooltip="Remove tag from filter" :style="{backgroundColor: t.color}" @click="closeTag(t.name)">
            {{ t.name }} <fa :icon="['fas', 'xmark']" style="margin-left: 4px;"/>
          </span>
      </div>
    </div>

    <table class="list shadow-normal border-rounded uk-background-primary mt-5 uk-table uk-table-middle">
      <thead class="head">
      <tr>
        <th style="width: 20px;"></th>
        <th class="uk-width-small">Check</th>
        <th class="uk-width-small">Availability</th>
        <th class="uk-width-large" style="text-align: center;">Response time</th>
        <th class="uk-width-small" style="text-align: right;">Tags</th>
      </tr>
      </thead>
      <c-service v-for="c in list" :value="c" @filter-by-tag="filterByTag"/>
    </table>
  </main>
</template>

<script>
import service from "@/components/service.vue"

export default {
  name: "list",
  components: {"c-service": service},
  props: ["services"],
  data: () => ({
    filter: {
      status: "",
      tags: [],
    },
  }),
  computed: {
    list() {
      let list = this.services
      if (this.filter.status) {
        list = this.services.filter(item => item.status.value === this.filter.status)
      }
      if (this.filter.tags.length > 0) {
        list = list.filter(item => {
          let tags = item.tags.map(t => t.name)
          return this.filter.tags.every(t => tags.includes(t.name))
        })
      }
      return list
    },
    tags() {
      let tags = []
      this.list.forEach(item => {
        item.tags.forEach(tag => {
          if (!tags.includes(tag.name)) {
            tags.push(tag.name)
          }
        })
      })
      return tags
    },
    upCount() {
      return this.services.filter(item => item.status.value === "up").length
    },
    downCount() {
      return this.services.filter(item => item.status.value === "down").length
    },
  },
  methods: {
    filterByStatus(status) {
      if (this.filter.status === status) {
        this.filter.status = null
        return
      }
      this.filter.status = status
    },
    filterByTag(tag) {
      if (this.filter.tags.find(t => t.name === tag.name)) {
        return
      }
      this.filter.tags.push(tag)
    },
    closeTag(tag) {
      this.filter.tags = this.filter.tags.filter(t => t.name !== tag)
    },
  },
}
</script>

<style lang="scss">
@import "@/style.scss";

main {
  flex: 1;
  .filters {
    height: 40px;

    .statuses {
      background: var(--default-color);
    }
    .status {
      cursor: pointer;
      padding: 8px;
      transition: background-color 0.1s ease-in-out;
      .active-mark {
        display: none;
      }
      &:hover {
        background: var(--muted-color);
      }
      &.active {
        background: var(--muted-color);
        .active-mark {
          display: block;
        }
      }
    }
  }

  .list {
    background: var(--default-color);
    thead {
      background: var(--muted-color);
      th {
        font-weight: 600;
        &:first-child {
          border-top-left-radius: 5px;
        }
        &:last-child {
          border-top-right-radius: 5px;
        }
      }
    }
    tbody {
      border-top: solid var(--muted-color) 1px;
      &:first-child {
        border-top: none;
      }
    }
  }

  .tag {
    margin: 1px;
    cursor: pointer;
  }
}
</style>