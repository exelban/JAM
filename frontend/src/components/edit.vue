<template>
  <form>
    <ul class="uk-flex-center"  uk-tab>
      <li class="uk-active"><a href="#">Basic</a></li>
      <li><a href="#">Intervals</a></li>
      <li><a href="#">Headers</a></li>
      <li><a href="#">History</a></li>
    </ul>

    <ul class="uk-switcher uk-margin">
      <li>
        <div class="mv-2 row between nowrap">
          <select class="uk-select" v-model="form.type" style="width: 120px;" aria-label="Type">
            <option>http</option>
            <option>mongodb</option>
          </select>
          <input class="uk-input" :class="{'uk-form-danger': errors.name, 'uk-form-success': errors.name === false}" v-model="form.name" type="text" placeholder="Name" aria-label="Name">
        </div>
        <div class="mv-2 row between nowrap">
          <select class="uk-select" :class="{'uk-form-danger': errors.method, 'uk-form-success': errors.method === false}" v-model="form.method" style="width: 120px;" aria-label="Method">
            <option>GET</option>
            <option>POST</option>
            <option>PUT</option>
            <option>PATCH</option>
            <option>DELETE</option>
            <option>HEAD</option>
            <option>OPTIONS</option>
          </select>
          <input class="uk-input" :class="{'uk-form-danger': errors.url, 'uk-form-success': errors.url === false}" type="text" v-model="form.url" placeholder="URL" aria-label="URL">
        </div>
      </li>
      <li>
        <div class="mv-2 row between middle nowrap">
          <label class="uk-form-label" style="flex-grow: 1;flex-basis: 0;" for="form-retry">Retry interval (seconds)</label>
          <input class="uk-input" :class="{'uk-form-danger': errors.retry, 'uk-form-success': errors.retry === false}" v-model="form.retry" style="flex-grow: 1;flex-basis: 0;" id="form-retry" type="number">
        </div>
        <div class="mv-2 row between middle nowrap">
          <label class="uk-form-label" style="flex-grow: 1;flex-basis: 0;" for="form-timeout">Timeout (seconds)</label>
          <input class="uk-input" :class="{'uk-form-danger': errors.timeout, 'uk-form-success': errors.timeout === false}" v-model="form.timeout" style="flex-grow: 1;flex-basis: 0;" id="form-timeout" type="number">
        </div>
        <div class="mv-2 row between middle nowrap">
          <label class="uk-form-label" style="flex-grow: 1;flex-basis: 0;" for="form-delay">Initial delay (seconds)</label>
          <input class="uk-input" :class="{'uk-form-danger': errors.delay, 'uk-form-success': errors.delay === false}" v-model="form.delay" style="flex-grow: 1;flex-basis: 0;" id="form-delay" type="number">
        </div>
        <div class="mv-2 row between middle nowrap">
          <label class="uk-form-label" style="flex-grow: 1;flex-basis: 0;" for="form-success">Success threshold</label>
          <input class="uk-input" :class="{'uk-form-danger': errors.success, 'uk-form-success': errors.success === false}" v-model="form.success" style="flex-grow: 1;flex-basis: 0;" id="form-success" type="number">
        </div>
        <div class="mv-2 row between middle nowrap">
          <label class="uk-form-label" style="flex-grow: 1;flex-basis: 0;" for="form-failure">Failure threshold</label>
          <input class="uk-input" :class="{'uk-form-danger': errors.failure, 'uk-form-success': errors.failure === false}" v-model="form.failure" style="flex-grow: 1;flex-basis: 0;" id="form-failure" type="number">
        </div>
      </li>
      <li>Bazinga!</li>
      <li>Bazinga!</li>
    </ul>
  </form>
</template>

<script>
import {mapState} from "vuex"

export default {
  name: "c-edit-target",
  data: () => ({
    errors: {
      name: undefined,
      url: undefined,
      retry: undefined,
      timeout: undefined,
      delay: undefined,
      success: undefined,
      failure: undefined
    }
  }),
  computed: {
    ...mapState(["form"]),
  },
  methods: {
    async check() {
      return new Promise((resolve, reject) => {
        this.errors.name = this.form.name === ""
        this.errors.url = this.form.url === ""
        this.errors.retry = this.form.retry < 0
        this.errors.timeout = this.form.timeout < 0
        this.errors.delay = this.form.delay < 0
        this.errors.success = this.form.success < 0
        this.errors.failure = this.form.failure < 0

        if (this.errors.name || this.errors.url || this.errors.retry || this.errors.timeout || this.errors.delay || this.errors.success || this.errors.failure) {
          reject()
        } else {
          resolve()
        }
      })
    },
    reset() {
      this.errors.name = undefined
      this.errors.url = undefined
      this.errors.retry = undefined
      this.errors.timeout = undefined
      this.errors.delay = undefined
      this.errors.success = undefined
      this.errors.failure = undefined
    }
  },
}
</script>

<style lang="scss">

</style>