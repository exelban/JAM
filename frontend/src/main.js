import { createApp } from "vue"

import App from "./App.vue"

import UIkit from "uikit"

import { library } from "@fortawesome/fontawesome-svg-core"
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome"
import { faFilter, faPlus, faCircleCheck, faCircleXmark, faCheck, faXmark, faCircleHalfStroke } from "@fortawesome/free-solid-svg-icons"

library.add(faFilter, faPlus, faCircleCheck, faCircleXmark, faCheck, faXmark, faCircleHalfStroke)

createApp(App)
  .component("fa", FontAwesomeIcon)
  .mount("#app")