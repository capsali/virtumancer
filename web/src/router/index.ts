import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue')
  },
  {
    path: '/hosts/:hostId',
    name: 'host-dashboard',
    component: () => import('@/views/HostDashboard.vue'),
    props: true
  },
  {
    path: '/hosts/:hostId/vms/:vmName',
    name: 'vm-detail',
    component: () => import('@/views/VMDetailView.vue'),
    props: true
  },
  {
    path: '/spice/:hostId/:vmName',
    name: 'spice-console',
    component: () => import('@/views/SpiceView.vue'),
    props: true
  },
  {
    path: '/network',
    name: 'network-topology',
    component: () => import('@/views/NetworkTopologyView.vue')
  },
  {
    path: '/error-demo',
    name: 'error-demo',
    component: () => import('@/components/ui/ErrorHandlingDemo.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router