import Vuex from "vuex"

const Cookies = {
  get(key) {
    const nameEQ = key + "="
    const ca = document.cookie.split(';')
    for (let i=0; i < ca.length; i++) {
      let c = ca[i]
      while (c.charAt(0) === ' ') c = c.substring(1,c.length)
      if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length,c.length)
    }
    return null
  },
  set(key, value, hours) {
    let expires = ""
    if (hours) {
      const date = new Date()
      date.setTime(date.getTime() + (hours*60*60*1000))
      expires = "; expires=" + date.toUTCString()
    }
    document.cookie = key + "=" + (value || "")  + expires + "; path=/"
  },
  erase(key) {
    document.cookie = key +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;'
  },
}

export default new Vuex.Store({
  state: {
    targets: undefined,
    theme: Cookies.get("theme") || "light",
    form: {
      type: "http",
      method: "GET",
    },
    target: {}
  },
  mutations: {
    setTarget(state, target) {
      state.target = target
    },
    setTargets(state, targets) {
      state.targets = targets
    },
    setForm(state, form) {
      state.form = form
    },
    resetForm(state) {
      state.form = {
        type: "http",
        method: "GET",
      }
    },
    setTheme(state, theme) {
      state.theme = theme
      Cookies.set("theme", theme, 24)
    }
  },
  actions: {
    toggleTheme({ state, commit }) {
      const newTheme = state.theme === "light" ? "dark" : "light"
      document.body.classList.remove(state.theme)
      commit("setTheme", newTheme)
      document.body.classList.add(state.theme)
    },

    async getTargets({ commit }) {
      const response = await fetch("http://localhost:8080/target").catch((error) => {
        commit("setTargets", null)
        return null
      })

      if (!response) return
      const targets = await response.json()
      commit("setTargets", targets)
    },
    async createTarget({ commit, dispatch }, target) {
      await fetch("http://localhost:8080/target", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(target),
      })

      commit("resetForm")

      setTimeout(() => {
        dispatch("getTargets")
      }, 1000)
    },
    async deleteTarget({ state, dispatch }) {
      await fetch(`http://localhost:8080/target/${state.target.id}`, {
        method: "DELETE",
      })
      dispatch("getTargets")
    },
    async editTarget({ dispatch }, target) {
      await fetch(`http://localhost:8080/target/${target.id}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(target),
      })
      dispatch("getTargets")
    },
  }
})