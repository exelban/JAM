<template>
  <div class="app w h column">
    <main class="w h row center middle" v-if="!targets || !targets.length">
      <div v-if="targets === undefined" uk-spinner="ratio: 3"></div>
      <div v-else-if="targets === null">Error receiving data from the backend</div>
      <div v-else-if="!targets.length" class="column center middle">
        <button class="uk-button uk-button-primary shadow-normal border-rounded" uk-toggle="target: #edit-target-dialog">Add target</button>
      </div>
    </main>

    <c-list v-else :targets="targets"/>

    <footer class="p-medium row middle between">
      <a href="https://github.com/exelban/uptime" target="_blank">v0.0.0</a>
      <fa :icon="['fas', 'circle-half-stroke']" size="lg" class="pointer" uk-tooltip="Change theme" @click="toggleTheme"/>
    </footer>

    <div id="edit-target-dialog" uk-modal ref="editTarget">
      <div class="uk-modal-dialog uk-margin-auto-vertical">
        <div class="uk-modal-body" style="padding-top: 10px;padding-bottom: 10px;">
          <c-edit ref="edit"/>
        </div>
        <div class="uk-modal-footer uk-text-right">
          <button class="uk-button uk-button-default uk-modal-close" type="button">Cancel</button>
          <button class="uk-button uk-button-primary" type="button" v-if="!form.id" @click="create">Create</button>
          <button class="uk-button uk-button-primary" type="button" v-else @click="save">Save</button>
        </div>
      </div>
    </div>
    <div id="delete-target-dialog" uk-modal ref="deleteTarget">
      <div class="uk-modal-dialog uk-margin-auto-vertical">
        <div class="uk-modal-body">
          Are you sure you want to delete <b>{{target.name}}</b> monitoring?
        </div>
        <div class="uk-modal-footer uk-text-right">
          <button class="uk-button uk-button-default uk-modal-close" type="button">Cancel</button>
          <button class="uk-button uk-button-danger" type="button" @click="deleteTarget">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import UIkit from "uikit"
import {mapState} from "vuex"

import list from "@/components/list.vue"
import edit from "@/components/edit.vue"

export default {
  name: "app",
  components: {"c-edit": edit, "c-list": list},
  computed: {
    ...mapState(["theme", "targets", "form", "target"]),
  },
  methods: {
    toggleTheme() {
      this.$store.dispatch("toggleTheme")
    },
    async create() {
      this.$refs.edit.check().then(() => {
        this.$store.dispatch("createTarget", this.$refs.edit.form).then(() => {
          UIkit.modal(this.$refs.editTarget).hide()
          this.resetForm()
        })
      }).catch(() => {})
    },
    async save() {
      this.$refs.edit.check().then(() => {
        this.$store.dispatch("editTarget", this.$refs.edit.form).then(() => {
          UIkit.modal(this.$refs.editTarget).hide()
          this.resetForm()
        })
      }).catch(() => {})
    },
    async deleteTarget() {
      this.$store.dispatch("deleteTarget").then(() => {
        UIkit.modal(this.$refs.deleteTarget).hide()
      })
    },
    resetForm() {
      this.$store.commit("resetForm")
      this.$refs.edit.reset()
    },
    resetTarget() {
      this.$store.commit("setTarget", {})
    }
  },
  beforeMount() {
    this.$store.dispatch("getTargets")
  },
  mounted() {
    document.body.classList.add(this.theme)
    this.$refs.editTarget.addEventListener("beforehide", this.resetForm)
    this.$refs.deleteTarget.addEventListener("beforehide", this.resetTarget)
  },
  beforeUnmount() {
    this.$refs.editTarget.removeEventListener("beforehide", this.resetForm)
    this.$refs.deleteTarget.removeEventListener("beforehide", this.resetTarget)
  }
}
</script>

<style lang="scss">
@import "style.scss";
</style>