# Network Topology View Documentation

## Overview

The NetworkTopologyView component provides both grid and interactive network diagram views of the virtualization infrastructure, showing hosts, virtual machines, and their relationships. This component delivers an intuitive visual representation of the datacenter topology with real-time status updates.

## Features

### Dual View Modes

1. **Grid View**: Displays hosts as cards with detailed VM information
2. **Network View**: Interactive SVG-based network diagram with circular topology

### Interactive Elements

- **Host Cards**: Click to navigate to host dashboard
- **VM Cards**: Click to navigate to VM details
- **Network Nodes**: Interactive host and VM nodes with status indicators
- **Real-time Updates**: Automatic refresh of topology data

### Status Visualization

- **Host States**: Connected (green), Disconnected (red), Error (red)
- **VM States**: Active (green), Stopped (red), Paused (yellow), Error (red)
- **Connection Lines**: Color-coded based on VM state in network view

## Implementation Details

### Component Structure

Located at: `web/src/views/NetworkTopologyView.vue`

### Key Features

#### Grid View
- Responsive card layout with glass morphism effects
- Real-time statistics display (total hosts, VMs, connected hosts, active VMs)
- Host information cards showing:
  - Connection status and display name
  - VM count and status breakdown
  - Individual VM cards with state indicators

#### Network Diagram
- SVG-based circular topology layout
- Host nodes positioned in circle with radius of 250px
- VM nodes positioned around their host in 80px radius
- Interactive elements with hover effects
- Legend showing status color mappings
- Limits display to 8 VMs per host (with overflow indicator)

### Data Management

#### Store Integration
- **hostStore**: Manages host data and state
- **vmStore**: Manages VM data with host-based filtering

#### Real-time Updates
- Fetches all hosts on component mount
- Loads VMs for each connected host
- Automatic error handling and loading states

### Styling & Theme

#### Visual Effects
- Glass morphism cards with backdrop blur
- Neon glow effects based on component state
- Smooth transitions and hover animations
- Color-coded status indicators

#### Responsive Design
- Adaptive layout for different screen sizes
- Collapsible elements for mobile optimization
- Touch-friendly interactive elements

## Usage

### Navigation
Access via the main navigation menu as "Network" view.

### View Toggle
Use the toggle buttons in the header to switch between:
- Grid view: Detailed card-based layout
- Network view: Interactive topology diagram

### Interactions
- **Host Navigation**: Click any host card or node to view host dashboard
- **VM Navigation**: Click any VM card or node to view VM details
- **Refresh**: Use the refresh button to manually update topology data

## API Dependencies

### REST Endpoints
- `GET /api/v1/hosts` - Fetch all configured hosts
- `GET /api/v1/hosts/:id/vms` - Fetch VMs for specific host

### WebSocket Events
- `hosts-changed` - Triggers host data refresh
- `vms-changed` - Triggers VM data refresh for affected host

## Configuration

### Display Limits
- Maximum VMs shown per host in network view: 8
- Network diagram dimensions: 1200x800px
- Host circle radius: 250px
- VM circle radius: 80px

### Status Colors
- **Connected/Active**: Green (#10b981)
- **Disconnected/Stopped**: Red (#ef4444)
- **Paused/Warning**: Yellow (#f59e0b)
- **Error**: Dark Red (#dc2626)

## Performance Considerations

### Data Loading
- Parallel VM fetching for multiple hosts
- Error boundary prevents cascade failures
- Loading states prevent UI blocking

### SVG Optimization
- Efficient positioning calculations
- Minimal DOM updates for status changes
- Event delegation for interactive elements

## Future Enhancements

### Planned Features
1. **Zoom Controls**: Pan and zoom functionality for large topologies
2. **Filtering**: Hide/show hosts or VMs based on status
3. **Search**: Quick search for specific hosts or VMs
4. **Layout Options**: Alternative layouts (tree, force-directed)
5. **Export**: Export topology as image or PDF

### Accessibility
- Keyboard navigation support
- Screen reader compatibility
- High contrast mode support

## Code Examples

### Basic Usage
```vue
<template>
  <NetworkTopologyView />
</template>

<script setup>
import NetworkTopologyView from '@/views/NetworkTopologyView.vue'
</script>
```

### Store Integration
```typescript
// Access topology data
const hosts = computed(() => hostStore.hosts)
const vms = computed(() => vmStore.vms)

// Get VMs for specific host
const hostVMs = vmStore.vmsByHost(hostId)
```

### Navigation Integration
```typescript
// Navigate to host dashboard
router.push(`/hosts/${host.id}`)

// Navigate to VM details
router.push(`/hosts/${host.id}/vms/${vm.name}`)
```

## Error Handling

### Connection Failures
- Graceful degradation when hosts are unreachable
- Error messages for failed VM data loading
- Retry mechanisms for transient failures

### UI States
- Loading indicators during data fetching
- Error state display with user-friendly messages
- Empty state handling for no hosts/VMs

## Testing

### Unit Tests
- Component rendering tests
- Store integration tests
- Event handling verification

### Integration Tests
- API endpoint integration
- WebSocket event handling
- Navigation flow testing

### Visual Tests
- SVG rendering accuracy
- Responsive layout verification
- Theme and color consistency