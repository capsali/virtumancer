# Discovered VM Bulk Management Documentation

## Overview

The Discovered VM Bulk Management system provides comprehensive tools for handling multiple discovered VMs efficiently. This feature enables administrators to select, import, and manage multiple unmanaged virtual machines through an intuitive interface with batch operations.

## Features

### Frontend Components

#### DiscoveredVMBulkManager
**Location**: `web/src/components/vm/DiscoveredVMBulkManager.vue`

**Key Features**:
- **Multi-selection Interface**: Checkbox-based selection with select-all functionality
- **Bulk Operations**: Import and delete operations on selected VMs
- **Search and Filter**: Real-time search by name or UUID with sorting options
- **Progress Indicators**: Visual feedback during bulk operations
- **Empty State Handling**: User-friendly messages when no VMs are found

**Selection Controls**:
- Individual VM selection via checkboxes
- Select All / Clear All functionality with visual state indicators
- Selected count display with contextual messaging
- Partial selection state indication

**Bulk Actions**:
- **Bulk Import**: Import multiple selected VMs into management
- **Bulk Delete**: Remove selected discovered VMs from the database
- **Clear Selection**: Reset all selections

**Search and Sorting**:
- Text search by VM name or domain UUID
- Sort options: Name (A-Z, Z-A), Date (Latest/Oldest first)
- Real-time filtering without affecting selection state

### Backend Implementation

#### Storage Layer
**Location**: `internal/storage/discovered_vms.go`

**New Functions**:
- `BulkDeleteDiscoveredVMs(db, hostID, domainUUIDs)` - Remove multiple VMs by UUID list
- `BulkMarkDiscoveredVMsImported(db, hostID, domainUUIDs)` - Mark multiple VMs as imported

#### Service Layer
**Location**: `internal/services/host_service.go`

**New Methods**:
- `ImportSelectedVMs(hostID, domainUUIDs)` - Import specific VMs by UUID list
- `DeleteSelectedDiscoveredVMs(hostID, domainUUIDs)` - Remove specific VMs from database

**Features**:
- Concurrent VM import with mutex protection
- Bulk database operations for efficiency
- WebSocket notifications for real-time updates
- Error resilience with partial failure handling

#### API Layer
**Location**: `internal/api/handlers.go`

**New Endpoints**:
- `POST /api/v1/hosts/{hostID}/vms/import-selected` - Import selected VMs
- `DELETE /api/v1/hosts/{hostID}/discovered-vms` - Delete selected discovered VMs

**Request Format**:
```json
{
  "domain_uuids": ["uuid1", "uuid2", "uuid3"]
}
```

### Frontend Store Integration

#### Host Store Updates
**Location**: `web/src/stores/hostStore.ts`

**New Methods**:
- `importSelectedVMs(hostId, domainUUIDs)` - Import selected VMs with error handling
- `deleteSelectedDiscoveredVMs(hostId, domainUUIDs)` - Delete selected VMs with refresh

**Features**:
- Automatic data refresh after operations
- Error handling with user-friendly messages
- Loading state management
- WebSocket event integration

## Usage

### Host Dashboard Integration
The bulk management interface is integrated into the Host Dashboard's discovered VMs tab:

1. **Selection**: Use checkboxes to select individual VMs or select all
2. **Bulk Import**: Click "Import Selected" to add VMs to management
3. **Bulk Delete**: Click "Remove Selected" to delete VMs from discovery list
4. **Search**: Use the search bar to filter VMs by name or UUID
5. **Sort**: Change sorting order using the dropdown menu

### API Usage

#### Import Selected VMs
```bash
curl -X POST http://localhost:8080/api/v1/hosts/kvmsrv/vms/import-selected \
  -H "Content-Type: application/json" \
  -d '{"domain_uuids": ["uuid1", "uuid2"]}'
```

#### Delete Selected Discovered VMs
```bash
curl -X DELETE http://localhost:8080/api/v1/hosts/kvmsrv/discovered-vms \
  -H "Content-Type: application/json" \
  -d '{"domain_uuids": ["uuid1", "uuid2"]}'
```

## Technical Implementation

### Data Flow

1. **Discovery**: VMs are automatically discovered and stored in `discovered_vms` table
2. **Selection**: Frontend provides multi-selection interface
3. **Bulk Operations**: Selected UUIDs sent to backend for processing
4. **Processing**: Backend processes operations with concurrency control
5. **Notification**: WebSocket events notify clients of changes
6. **Refresh**: Frontend automatically refreshes data

### Performance Optimizations

#### Backend
- **Bulk Database Operations**: Single queries for multiple records
- **Concurrent Processing**: Parallel VM import with mutex protection
- **Efficient Selection**: Database queries with UUID IN clauses
- **Minimal Locking**: Individual VM processing to reduce contention

#### Frontend
- **Virtual Scrolling**: Efficient rendering for large VM lists
- **Debounced Search**: Prevent excessive filtering operations
- **State Management**: Optimized selection state with Set data structure
- **Progressive Enhancement**: Graceful degradation for large datasets

### Error Handling

#### Backend Resilience
- Partial failure handling: Continue processing remaining VMs if some fail
- Transaction rollback: Atomic operations where appropriate
- Detailed logging: Comprehensive error reporting for debugging
- Graceful degradation: System remains functional with partial failures

#### Frontend Recovery
- Toast notifications: User-friendly error messages
- State preservation: Maintain selections across operations
- Retry mechanisms: Allow users to retry failed operations
- Loading indicators: Clear feedback during long operations

## Configuration

### Limits and Thresholds
- Maximum selections: No hard limit (UI optimized for reasonable use)
- Batch size: Operations processed in configurable batches
- Timeout settings: Configurable timeouts for bulk operations
- Concurrency: Adjustable concurrent import limits

### Customization Options
- Sort order preferences: User-configurable default sorting
- Selection persistence: Optional selection state preservation
- Notification preferences: Configurable toast message settings
- Refresh intervals: Adjustable auto-refresh rates

## Future Enhancements

### Planned Features
1. **Export Operations**: Export VM lists to CSV/JSON
2. **Saved Selections**: Persist selection sets across sessions
3. **Advanced Filtering**: Filter by VM properties (CPU, memory, state)
4. **Batch Scheduling**: Schedule bulk operations for later execution
5. **Audit Trail**: Track bulk operation history and results

### Performance Improvements
1. **Pagination**: Handle very large VM lists efficiently
2. **Streaming Updates**: Real-time progress for bulk operations
3. **Background Processing**: Queue large operations for background processing
4. **Caching**: Intelligent caching of discovery data

## Security Considerations

### Access Control
- Operation authorization based on host permissions
- Audit logging for bulk operations
- Rate limiting for API endpoints
- Input validation for UUID lists

### Data Protection
- Sanitized error messages to prevent information disclosure
- Secure handling of VM metadata
- Protection against bulk operation abuse
- Transaction integrity for concurrent operations