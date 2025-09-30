import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router/index'
import { vClickAway } from './directives/clickAway'
import './style.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.directive('click-away', vClickAway)
app.mount('#app')