# **TypeScript Type System Documentation**

## **Overview**

Virtumancer's frontend employs a comprehensive TypeScript type system that ensures type safety, improved developer experience, and runtime reliability. The type definitions are centralized in `web/src/types/index.ts` and provide complete coverage for all API responses, frontend data structures, and WebSocket communications.

## **Core Type Definitions**

### **Virtual Machine Types**

```typescript
interface VirtualMachine {
  db_id: number;
  host_id: string;
  uuid: string;
  name: string;
  description: string;
  vcpu_count: number;
  memory_bytes: number;
  state: number;
  cpu_model?: string;
  cpu_topology_json?: string;
  created_at?: string;
  updated_at?: string;
  graphics?: {
    vnc?: boolean;
    spice?: boolean;
  };
}
```

### **Host Management Types**

```typescript
interface Host {
  id: string;
  uri: string;
  state: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  task_state?: string;
  auto_reconnect_disabled: boolean;
  created_at?: string;
  updated_at?: string;
  stats?: HostStats;
}
```

### **Hardware Configuration Types**

```typescript
interface VMHardware {
  disks: VMDisk[];
  networks: VMNetwork[];
  graphics?: VMGraphics[];
}

interface VMDisk {
  type: string;
  device: string;
  driver: {
    driver_name: string;
    type: string;
  };
  path: string;
  target: {
    dev: string;
    bus: string;
  };
  size_bytes?: number;
}

interface VMNetwork {
  type: string;
  mac: {
    address: string;
  };
  source: {
    bridge?: string;
    network?: string;
    portgroup?: string;  // Added for OpenVSwitch support
  };
  model: {
    model_type: string;
  };
}
```

## **Real-Time Communication Types**

### **WebSocket Message Types**

```typescript
interface WebSocketMessage {
  type: 'vm_stats' | 'host_stats' | 'vm_state_change' | 'host_state_change';
  data: any;
}

interface VMStatsMessage extends WebSocketMessage {
  type: 'vm_stats';
  data: {
    host_id: string;
    vm_uuid: string;
    stats: VMStats;
  };
}
```

### **Statistics Types**

```typescript
interface VMStats {
  cpu_percent: number;
  memory_used: number;
  memory_total: number;
  disk_stats: VMDiskStats[];
  network_stats: VMNetworkStats[];
  timestamp: string;
}

interface HostStats {
  vm_count: number;
  cpu_percent: number;
  memory_total: number;
  memory_available: number;
  uptime?: number;
  host_info?: {
    hostname: string;
    cpu: string;
    memory: number;
    version: string;
  };
}
```

## **Form and UI Types**

### **User Interface State**

```typescript
interface UIState {
  loading: {
    hosts?: boolean;
    vms?: boolean;
    connectHost?: Record<string, boolean>;
    hostStats?: Record<string, boolean>;
  };
  errors: {
    message: string;
    timestamp: string;
  }[];
  modals: {
    createVM: boolean;
    vmDetail: boolean;
    vmHardware: boolean;
  };
}
```

### **Form Validation Types**

```typescript
interface CreateVMForm {
  name: string;
  description?: string;
  vcpu_count: number;
  memory_mb: number;
  disk_size_gb: number;
  iso_path?: string;
  network_source: string;
}

interface VMActionForm {
  action: 'start' | 'shutdown' | 'reboot' | 'destroy' | 'reset';
}
```

## **Type System Benefits**

### **Compile-Time Safety**
- **API Contract Enforcement**: Ensures frontend correctly handles all API response fields
- **Null Safety**: Proper optional field handling prevents runtime errors
- **Type Checking**: Catches type mismatches during development
- **Refactoring Support**: Safe code changes with IDE assistance

### **Developer Experience**
- **IntelliSense Support**: Full autocomplete for all data structures
- **Documentation Integration**: Types serve as living documentation
- **Error Prevention**: Catches common mistakes before runtime
- **Code Navigation**: Jump-to-definition for type properties

### **Runtime Reliability**
- **Data Structure Validation**: Ensures consistent data handling
- **WebSocket Message Safety**: Proper typing for real-time communications
- **Form Validation**: Type-safe form handling and submission
- **State Management**: Strongly typed Pinia store interfaces

## **Type System Architecture**

### **Centralized Definition Strategy**
All types are defined in a single `types/index.ts` file to:
- Maintain consistency across the application
- Enable easy updates and refactoring
- Provide a single source of truth
- Simplify import statements

### **API Response Mapping**
Types directly mirror backend API responses to ensure:
- Accurate data representation
- Seamless API integration
- Consistent field naming
- Proper optional field handling

### **Component Integration**
Vue components leverage types for:
- Props validation
- Emit event typing
- Computed property types
- Template type checking

## **Maintenance and Evolution**

### **Adding New Types**
When adding new features:
1. Define types in `types/index.ts`
2. Update API response interfaces
3. Add WebSocket message types if needed
4. Update form validation types
5. Test type coverage with TypeScript compiler

### **Breaking Changes**
When modifying existing types:
1. Check all usage locations
2. Update component props and emits
3. Verify API response compatibility
4. Test WebSocket message handling
5. Update documentation

### **Type Safety Validation**
Regular validation ensures:
- No `any` types in production code
- Proper optional field handling
- Complete API response coverage
- WebSocket message type safety
- Form validation completeness

## **Critical System Components**

### **Emergency Type Recovery**
The type system is critical for application functionality. If types are lost or corrupted:
1. Application will fail to compile
2. Components will not render properly
3. API integration will break
4. WebSocket communication will fail

### **System Recovery Process**
In case of type system failure:
1. Restore from `types/index.ts` backup
2. Verify all interface definitions
3. Check API response compatibility
4. Test WebSocket message types
5. Validate form handling types
6. Run full TypeScript compilation
7. Test application functionality

This comprehensive type system ensures Virtumancer's frontend maintains high code quality, developer productivity, and runtime reliability across all features and integrations.