import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/hosts/:hostId/vms/:vmName/console',
      name: 'console',
      // Route level code-splitting for VNC console
      component: () => import('../views/ConsoleView.vue')
    },
    {
      path: '/hosts/:hostId/vms/:vmName/spice',
      name: 'spice',
      // Route level code-splitting for SPICE console
      component: () => import('../views/SpiceView.vue')
    }
  ]
})

export default router


