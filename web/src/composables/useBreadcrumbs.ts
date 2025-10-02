import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

export interface BreadcrumbItem {
  label: string
  path?: string
  icon?: string
  isActive?: boolean
}

export function useBreadcrumbs() {
  const route = useRoute()
  const router = useRouter()

  const breadcrumbs = computed<BreadcrumbItem[]>(() => {
    const items: BreadcrumbItem[] = []
    const path = route.path
    const pathSegments = path.split('/').filter(segment => segment)

    // Always start with Home
    items.push({
      label: 'Home',
      path: '/',
      icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6'
    })

    // Build breadcrumb trail based on route segments
    if (pathSegments.length > 0) {
      const firstSegment = pathSegments[0]
      
      if (firstSegment === 'vms') {
        items.push({
          label: 'Virtual Machines',
          path: '/vms',
          icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
        })
        
        // If we have more segments, it's a specific VM
        if (pathSegments.length > 1) {
          const vmName = (route.params.vmName as string) || pathSegments[pathSegments.length - 1] || 'Unknown VM'
          items.push({
            label: vmName,
            isActive: true
          })
        }
      }
      
      if (firstSegment === 'hosts') {
        items.push({
          label: 'Hosts',
          path: '/hosts',
          icon: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01'
        })
        
        // Host-specific pages
        if (pathSegments.length > 1) {
          const hostId = pathSegments[1]
          
          // If it's a VM under a host - show VM breadcrumb instead
          if (pathSegments.length > 3 && pathSegments[2] === 'vms') {
            // Replace with Virtual Machines breadcrumb for better UX
            items[items.length - 1] = {
              label: 'Virtual Machines',
              path: '/vms',
              icon: 'M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
            }
            
            // Add Managed VMs breadcrumb
            items.push({
              label: 'Managed VMs',
              path: '/vms/managed',
              icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z'
            })
            
            const vmName = pathSegments[3] || 'VM Details'
            items.push({
              label: vmName,
              isActive: true
            })
          } else {
            // Just host dashboard
            const hostName = route.params.hostId as string || 'Host Dashboard'
            items.push({
              label: hostName,
              isActive: true
            })
          }
        }
      }
      
      if (firstSegment === 'network') {
        items.push({
          label: 'Networks',
          path: '/network',
          icon: 'M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0',
          isActive: true
        })
      }
      
      if (firstSegment === 'settings') {
        items.push({
          label: 'Settings',
          path: '/settings',
          icon: 'M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z',
          isActive: true
        })
      }
    }

    return items
  })

  const navigateTo = (path: string) => {
    router.push(path)
  }

  const navigateBack = () => {
    // Smart back navigation - go to parent in breadcrumb if available
    const parentItem = breadcrumbs.value[breadcrumbs.value.length - 2]
    if (parentItem?.path) {
      router.push(parentItem.path)
    } else {
      router.back()
    }
  }

  return {
    breadcrumbs,
    navigateTo,
    navigateBack
  }
}