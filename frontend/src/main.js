import { createApp } from "vue"

import App from "./App.vue"

import store from "@/store"

import UIkit from "uikit"

import { library } from "@fortawesome/fontawesome-svg-core"
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome"
import { faFilter, faCircleCheck, faCircleXmark, faCheck, faXmark, faCircleHalfStroke, faCircleQuestion } from "@fortawesome/free-solid-svg-icons"

library.add(faFilter, faCircleCheck, faCircleXmark, faCheck, faXmark, faCircleHalfStroke, faCircleQuestion)

createApp(App)
  .use(store)
  .component("fa", FontAwesomeIcon)
  .mount("#app")