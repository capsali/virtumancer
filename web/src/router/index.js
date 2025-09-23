import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import HostDashboard from '@/components/views/HostDashboard.vue'
import VmView from '@/components/views/VmView.vue'
import Datacenter from '@/components/views/Datacenter.vue'
import NetworkView from '@/components/views/NetworkView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: HomeView,
      children: [
        {
          path: '',
          name: 'datacenter',
          component: Datacenter, // Default view is now the datacenter
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
        {
          path: 'network',
          name: 'network',
          component: NetworkView,
        },
      ],
    },
  ]
})

import { useMainStore } from '@/stores/mainStore'; // Import the main store

// Global navigation guard
router.beforeEach((to, from, next) => {
  const mainStore = useMainStore(); // Get the store instance inside the guard

  // If navigating away from a host-dashboard or vm-view to a non-subscribed route,
  // clear all active subscriptions.
  const isLeavingSubscribedRoute = (from.name === 'host-dashboard' || from.name === 'vm-view');
  const isGoingToUnsubscribedRoute = (to.name !== 'host-dashboard' && to.name !== 'vm-view');

  if (isLeavingSubscribedRoute && isGoingToUnsubscribedRoute) {
    mainStore.clearAllSubscriptions();
  }

  next();
});

export default router

