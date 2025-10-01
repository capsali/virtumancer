import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue')
  },
  {
    path: '/hosts/:hostId/dashboard',
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
    path: '/vnc/:hostId/:vmName',
    name: 'vnc-console',
    component: () => import('@/views/VNCView.vue'),
    props: true
  },
  // Network routes
  {
    path: '/network',
    name: 'network-overview',
    component: () => import('@/views/network/NetworkOverview.vue')
  },
  {
    path: '/network/networks',
    name: 'networks',
    component: () => import('@/views/network/NetworksView.vue')
  },
  {
    path: '/network/ports',
    name: 'network-ports',
    component: () => import('@/views/network/PortsView.vue')
  },
  {
    path: '/network/topology',
    name: 'network-topology',
    component: () => import('@/views/network/TopologyView.vue')
  },
  
  // Host routes
  {
    path: '/hosts',
    name: 'hosts-overview',
    component: () => import('@/views/hosts/HostsOverview.vue')
  },
  {
    path: '/hosts/:hostId',
    name: 'host-detail',
    component: () => import('@/views/hosts/HostDetailView.vue'),
    props: true
  },
  
  // VM routes
  {
    path: '/vms',
    name: 'vms-overview',
    component: () => import('@/views/vms/VMOverview.vue')
  },
  {
    path: '/vms/managed',
    name: 'managed-vms',
    component: () => import('@/views/vms/ManagedVMsView.vue')
  },
  {
    path: '/vms/discovered',
    name: 'discovered-vms',
    component: () => import('@/views/vms/DiscoveredVMsView.vue')
  },
  
  // Storage routes
  {
    path: '/storage',
    name: 'storage-overview',
    component: () => import('@/views/storage/StorageOverview.vue')
  },
  {
    path: '/storage/pools',
    name: 'storage-pools',
    component: () => import('@/views/storage/StoragePoolsView.vue')
  },
  {
    path: '/storage/volumes',
    name: 'storage-volumes',
    component: () => import('@/views/storage/StorageVolumesView.vue')
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/SettingsView.vue')
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