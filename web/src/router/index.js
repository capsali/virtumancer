import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import HostDashboard from '@/components/views/HostDashboard.vue'
import VmView from '@/components/views/VmView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: HomeView,
      children: [
        {
          path: '',
          name: 'home',
          component: HostDashboard, // Default view
        },
        {
          path: 'hosts/:hostId',
          name: 'host-dashboard',
          component: HostDashboard,
          props: true,
        },
        {
          path: 'vms/:vmName',
          name: 'vm-view',
          component: VmView,
          props: true,
        },
      ],
    },
    // The old full-page console routes are no longer needed.
    // They are replaced by the embedded console components in VmView.
  ]
})

export default router


