import './style.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { useMainStore } from './stores/mainStore'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())

const mainStore = useMainStore()

app.use(router)

router.isReady().then(() => {
    const currentRoute = router.currentRoute.value;
    if (currentRoute.name === 'host-dashboard' && currentRoute.params.hostId) {
        mainStore.selectHost(currentRoute.params.hostId);
    }
});

app.mount('#app')
