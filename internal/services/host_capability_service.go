package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/capsali/virtumancer/internal/libvirt"
	"github.com/capsali/virtumancer/internal/storage"
	"gorm.io/gorm"
)

// HostCapabilityService handles discovery and management of host capabilities
type HostCapabilityService struct {
	db        *gorm.DB
	connector *libvirt.Connector
}

// NewHostCapabilityService creates a new host capability service
func NewHostCapabilityService(db *gorm.DB, connector *libvirt.Connector) *HostCapabilityService {
	return &HostCapabilityService{
		db:        db,
		connector: connector,
	}
}

// HostCapabilityData represents comprehensive host capability information
type HostCapabilityData struct {
	// Basic host information
	HostInfo *HostInfo `json:"host_info"`

	// CPU capabilities
	CPUInfo *CPUCapabilityInfo `json:"cpu_info"`

	// Memory capabilities
	MemoryInfo *MemoryCapabilityInfo `json:"memory_info"`

	// Security capabilities
	SecurityInfo *SecurityCapabilityInfo `json:"security_info"`

	// Storage capabilities
	StorageInfo *StorageCapabilityInfo `json:"storage_info"`

	// Network capabilities
	NetworkInfo *NetworkCapabilityInfo `json:"network_info"`

	// Virtualization features
	VirtInfo *VirtualizationCapabilityInfo `json:"virt_info"`
}

type HostInfo struct {
	Architecture string `json:"architecture"`
	Model        string `json:"model"`
	Vendor       string `json:"vendor"`
	Version      string `json:"version"`
	Nodes        int    `json:"nodes"`
	Sockets      int    `json:"sockets"`
	Cores        int    `json:"cores"`
	Threads      int    `json:"threads"`
	Memory       uint64 `json:"memory"`
	CPUs         int32  `json:"cpus"`
	MHz          int32  `json:"mhz"`
}

type CPUCapabilityInfo struct {
	Models   []string         `json:"models"`
	Features []string         `json:"features"`
	Topology *CPUTopologyInfo `json:"topology"`
}

type CPUTopologyInfo struct {
	Sockets int `json:"sockets"`
	Cores   int `json:"cores"`
	Threads int `json:"threads"`
}

type MemoryCapabilityInfo struct {
	Total     uint64 `json:"total"`
	Available uint64 `json:"available"`
	Hugepages bool   `json:"hugepages"`
	KSM       bool   `json:"ksm"`
	Balloon   bool   `json:"balloon"`
}

type SecurityCapabilityInfo struct {
	Model      string `json:"model"`
	SELinux    bool   `json:"selinux"`
	AppArmor   bool   `json:"apparmor"`
	SecureBoot bool   `json:"secure_boot"`
	TPM        bool   `json:"tpm"`
	SEV        bool   `json:"sev"`
	IOMMU      bool   `json:"iommu"`
}

type StorageCapabilityInfo struct {
	Pools      []string `json:"pools"`
	Formats    []string `json:"formats"`
	QoS        bool     `json:"qos"`
	Encryption bool     `json:"encryption"`
}

type NetworkCapabilityInfo struct {
	Networks []string `json:"networks"`
	SRIOV    bool     `json:"sriov"`
	VirtIO   bool     `json:"virtio"`
	VLAN     bool     `json:"vlan"`
	QoS      bool     `json:"qos"`
}

type VirtualizationCapabilityInfo struct {
	Hypervisor      string   `json:"hypervisor"`
	Version         string   `json:"version"`
	NestedVirt      bool     `json:"nested_virt"`
	SupportedGuests []string `json:"supported_guests"`
	MaxVCPUs        int      `json:"max_vcpus"`
	MaxMemory       uint64   `json:"max_memory"`
}

// DiscoverHostCapabilities performs comprehensive host capability discovery
func (hcs *HostCapabilityService) DiscoverHostCapabilities(ctx context.Context, hostID string) (*HostCapabilityData, error) {
	log.Printf("Starting host capability discovery for host: %s", hostID)

	capabilities := &HostCapabilityData{}

	// Discover basic host information
	hostInfo, err := hcs.discoverHostInfo(hostID)
	if err != nil {
		log.Printf("Failed to discover host info: %v", err)
		return nil, fmt.Errorf("failed to discover host info: %w", err)
	}
	capabilities.HostInfo = hostInfo

	// Discover CPU capabilities
	cpuInfo, err := hcs.discoverCPUCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover CPU capabilities: %v", err)
	}
	capabilities.CPUInfo = cpuInfo

	// Discover memory capabilities
	memoryInfo, err := hcs.discoverMemoryCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover memory capabilities: %v", err)
	}
	capabilities.MemoryInfo = memoryInfo

	// Discover security capabilities
	securityInfo, err := hcs.discoverSecurityCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover security capabilities: %v", err)
	}
	capabilities.SecurityInfo = securityInfo

	// Discover storage capabilities
	storageInfo, err := hcs.discoverStorageCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover storage capabilities: %v", err)
	}
	capabilities.StorageInfo = storageInfo

	// Discover network capabilities
	networkInfo, err := hcs.discoverNetworkCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover network capabilities: %v", err)
	}
	capabilities.NetworkInfo = networkInfo

	// Discover virtualization capabilities
	virtInfo, err := hcs.discoverVirtualizationCapabilities(hostID)
	if err != nil {
		log.Printf("Failed to discover virtualization capabilities: %v", err)
	}
	capabilities.VirtInfo = virtInfo

	log.Printf("Completed host capability discovery for host: %s", hostID)
	return capabilities, nil
}

// discoverHostInfo uses NodeGetInfo to get basic host information
func (hcs *HostCapabilityService) discoverHostInfo(hostID string) (*HostInfo, error) {
	conn, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	model, memory, cpus, mhz, nodes, sockets, cores, threads, err := conn.NodeGetInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get node info: %w", err)
	}

	// Convert byte array to string for CPU model
	modelBytes := model[:]
	nullIndex := len(modelBytes)
	for i, b := range modelBytes {
		if b == 0 {
			nullIndex = i
			break
		}
	}
	byteModel := make([]byte, nullIndex)
	for i := 0; i < nullIndex; i++ {
		byteModel[i] = byte(modelBytes[i])
	}

	hostInfo := &HostInfo{
		Model:        string(byteModel),
		Nodes:        int(nodes),
		Sockets:      int(sockets),
		Cores:        int(cores),
		Threads:      int(threads),
		Memory:       memory,
		CPUs:         cpus,
		MHz:          mhz,
		Architecture: "x86_64", // Default, could be detected from capabilities
	}

	log.Printf("Host %s: Model=%s, Memory=%dKiB, CPUs=%d, MHz=%d", hostID, hostInfo.Model, memory, cpus, mhz)

	return hostInfo, nil
}

// discoverCPUCapabilities discovers CPU models, features, and topology
func (hcs *HostCapabilityService) discoverCPUCapabilities(hostID string) (*CPUCapabilityInfo, error) {
	conn, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	cpuInfo := &CPUCapabilityInfo{}

	// Get host capabilities XML to parse CPU features and models
	capsXML, err := conn.ConnectGetCapabilities()
	if err != nil {
		log.Printf("Failed to get capabilities XML: %v", err)
	} else {
		// Parse capabilities XML for CPU features and models
		features := hcs.parseCPUFeaturesFromXML(capsXML)
		cpuInfo.Features = features

		models := hcs.parseCPUModelsFromXML(capsXML)
		cpuInfo.Models = models
	}

	// Get topology information from node info
	_, _, _, _, _, sockets, cores, threads, err := conn.NodeGetInfo()
	if err == nil {
		cpuInfo.Topology = &CPUTopologyInfo{
			Sockets: int(sockets),
			Cores:   int(cores),
			Threads: int(threads),
		}
	}

	return cpuInfo, nil
}

// discoverMemoryCapabilities discovers memory features and capabilities
func (hcs *HostCapabilityService) discoverMemoryCapabilities(hostID string) (*MemoryCapabilityInfo, error) {
	conn, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	memoryInfo := &MemoryCapabilityInfo{}

	// Get basic memory info from NodeGetInfo
	_, memory, _, _, _, _, _, _, err := conn.NodeGetInfo()
	if err == nil {
		memoryInfo.Total = memory * 1024 // Convert from KiB to bytes
	}

	// Check for hugepage support via capabilities XML
	capsXML, err := conn.ConnectGetCapabilities()
	if err == nil {
		memoryInfo.Hugepages = strings.Contains(capsXML, "hugepages")
		memoryInfo.KSM = strings.Contains(capsXML, "ksm")
		memoryInfo.Balloon = true // Most KVM installations support memory balloon
	}

	return memoryInfo, nil
}

// discoverSecurityCapabilities discovers security features like SEV, TPM, etc.
func (hcs *HostCapabilityService) discoverSecurityCapabilities(hostID string) (*SecurityCapabilityInfo, error) {
	conn, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	securityInfo := &SecurityCapabilityInfo{}

	// Parse capabilities XML for security features
	capsXML, err := conn.ConnectGetCapabilities()
	if err == nil {
		securityInfo.SecureBoot = strings.Contains(capsXML, "secure-boot") || strings.Contains(capsXML, "efi")
		securityInfo.TPM = strings.Contains(capsXML, "tpm")
		securityInfo.IOMMU = strings.Contains(capsXML, "iommu")
		securityInfo.SEV = strings.Contains(capsXML, "sev")

		// Check for security models
		if strings.Contains(capsXML, "selinux") {
			securityInfo.Model = "selinux"
			securityInfo.SELinux = true
		} else if strings.Contains(capsXML, "apparmor") {
			securityInfo.Model = "apparmor"
			securityInfo.AppArmor = true
		} else {
			securityInfo.Model = "none"
		}
	}

	return securityInfo, nil
}

// discoverStorageCapabilities discovers storage pools and features
func (hcs *HostCapabilityService) discoverStorageCapabilities(hostID string) (*StorageCapabilityInfo, error) {
	_, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	storageInfo := &StorageCapabilityInfo{}

	// Get storage pools (simplified approach)
	storageInfo.Pools = []string{"default"} // Would need proper API to list all pools

	// Common storage formats supported by QEMU/KVM
	storageInfo.Formats = []string{"raw", "qcow2", "vmdk", "vdi", "vhd"}
	storageInfo.QoS = true
	storageInfo.Encryption = true

	return storageInfo, nil
}

// discoverNetworkCapabilities discovers network features and interfaces
func (hcs *HostCapabilityService) discoverNetworkCapabilities(hostID string) (*NetworkCapabilityInfo, error) {
	_, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	networkInfo := &NetworkCapabilityInfo{}

	// Get networks (simplified approach)
	networkInfo.Networks = []string{"default"} // Would need proper API to list all networks

	// Default network capabilities for KVM/QEMU
	networkInfo.VirtIO = true
	networkInfo.VLAN = true
	networkInfo.QoS = true
	networkInfo.SRIOV = false // Would need to check for actual SR-IOV devices

	return networkInfo, nil
}

// discoverVirtualizationCapabilities discovers hypervisor capabilities
func (hcs *HostCapabilityService) discoverVirtualizationCapabilities(hostID string) (*VirtualizationCapabilityInfo, error) {
	conn, err := hcs.connector.GetConnection(hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	virtInfo := &VirtualizationCapabilityInfo{}

	// Get hypervisor information
	hvType, err := conn.ConnectGetType()
	if err == nil {
		virtInfo.Hypervisor = hvType
	}

	version, err := conn.ConnectGetVersion()
	if err == nil {
		virtInfo.Version = fmt.Sprintf("%d", version)
	}

	// Get capabilities XML for detailed features
	capsXML, err := conn.ConnectGetCapabilities()
	if err == nil {
		virtInfo.NestedVirt = strings.Contains(capsXML, "vmx") || strings.Contains(capsXML, "svm")
		virtInfo.SupportedGuests = hcs.parseGuestTypesFromXML(capsXML)
	}

	// Get node info for max resources
	_, memory, _, _, _, sockets, cores, threads, err := conn.NodeGetInfo()
	if err == nil {
		virtInfo.MaxVCPUs = int(sockets * cores * threads)
		virtInfo.MaxMemory = uint64(memory) * 1024 // Convert from KiB to bytes
	}

	return virtInfo, nil
}

// Helper methods to parse XML capabilities
func (hcs *HostCapabilityService) parseCPUFeaturesFromXML(xml string) []string {
	features := []string{}
	// Basic parsing - in production, use proper XML parsing
	if strings.Contains(xml, "pae") {
		features = append(features, "pae")
	}
	if strings.Contains(xml, "apic") {
		features = append(features, "apic")
	}
	if strings.Contains(xml, "acpi") {
		features = append(features, "acpi")
	}
	if strings.Contains(xml, "vmx") {
		features = append(features, "vmx")
	}
	if strings.Contains(xml, "svm") {
		features = append(features, "svm")
	}
	if strings.Contains(xml, "sse") {
		features = append(features, "sse")
	}
	if strings.Contains(xml, "avx") {
		features = append(features, "avx")
	}
	if strings.Contains(xml, "avx2") {
		features = append(features, "avx2")
	}
	return features
}

func (hcs *HostCapabilityService) parseCPUModelsFromXML(xml string) []string {
	models := []string{}
	// Basic parsing - in production, use proper XML parsing
	if strings.Contains(xml, "host-passthrough") {
		models = append(models, "host-passthrough")
	}
	if strings.Contains(xml, "host-model") {
		models = append(models, "host-model")
	}
	// Add common CPU models
	models = append(models, "qemu64", "Nehalem", "Westmere", "SandyBridge", "IvyBridge", "Haswell", "Broadwell", "Skylake-Client")
	return models
}

func (hcs *HostCapabilityService) parseGuestTypesFromXML(xml string) []string {
	guests := []string{}
	if strings.Contains(xml, "hvm") {
		guests = append(guests, "hvm")
	}
	if strings.Contains(xml, "xen") {
		guests = append(guests, "xen")
	}
	if strings.Contains(xml, "kvm") {
		guests = append(guests, "kvm")
	}
	return guests
}

// StoreHostCapabilities saves discovered capabilities to the database
func (hcs *HostCapabilityService) StoreHostCapabilities(hostID string, capabilities *HostCapabilityData) error {
	// Clear existing capabilities for this host
	if err := hcs.db.Where("host_id = ?", hostID).Delete(&storage.HostCapability{}).Error; err != nil {
		return fmt.Errorf("failed to clear existing capabilities: %w", err)
	}

	// Store each capability category
	if capabilities.HostInfo != nil {
		if err := hcs.storeCapability(hostID, "host_info", "1.0", capabilities.HostInfo); err != nil {
			log.Printf("Failed to store host info: %v", err)
		}
	}

	if capabilities.CPUInfo != nil {
		if err := hcs.storeCapability(hostID, "cpu_info", "1.0", capabilities.CPUInfo); err != nil {
			log.Printf("Failed to store CPU info: %v", err)
		}
	}

	if capabilities.MemoryInfo != nil {
		if err := hcs.storeCapability(hostID, "memory_info", "1.0", capabilities.MemoryInfo); err != nil {
			log.Printf("Failed to store memory info: %v", err)
		}
	}

	if capabilities.SecurityInfo != nil {
		if err := hcs.storeCapability(hostID, "security_info", "1.0", capabilities.SecurityInfo); err != nil {
			log.Printf("Failed to store security info: %v", err)
		}
	}

	if capabilities.StorageInfo != nil {
		if err := hcs.storeCapability(hostID, "storage_info", "1.0", capabilities.StorageInfo); err != nil {
			log.Printf("Failed to store storage info: %v", err)
		}
	}

	if capabilities.NetworkInfo != nil {
		if err := hcs.storeCapability(hostID, "network_info", "1.0", capabilities.NetworkInfo); err != nil {
			log.Printf("Failed to store network info: %v", err)
		}
	}

	if capabilities.VirtInfo != nil {
		if err := hcs.storeCapability(hostID, "virt_info", "1.0", capabilities.VirtInfo); err != nil {
			log.Printf("Failed to store virtualization info: %v", err)
		}
	}

	log.Printf("Successfully stored host capabilities for host: %s", hostID)
	return nil
}

// storeCapability stores a single capability in the database
func (hcs *HostCapabilityService) storeCapability(hostID, name, version string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal capability data: %w", err)
	}

	capability := &storage.HostCapability{
		HostID:      hostID,
		Name:        name,
		Version:     version,
		DetailsJSON: string(jsonData),
	}

	return hcs.db.Create(capability).Error
}

// GetHostCapabilities retrieves stored capabilities for a host
func (hcs *HostCapabilityService) GetHostCapabilities(hostID string) (*HostCapabilityData, error) {
	var capabilities []storage.HostCapability
	if err := hcs.db.Where("host_id = ?", hostID).Find(&capabilities).Error; err != nil {
		return nil, fmt.Errorf("failed to get host capabilities: %w", err)
	}

	data := &HostCapabilityData{}

	for _, cap := range capabilities {
		switch cap.Name {
		case "host_info":
			var hostInfo HostInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &hostInfo); err == nil {
				data.HostInfo = &hostInfo
			}
		case "cpu_info":
			var cpuInfo CPUCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &cpuInfo); err == nil {
				data.CPUInfo = &cpuInfo
			}
		case "memory_info":
			var memoryInfo MemoryCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &memoryInfo); err == nil {
				data.MemoryInfo = &memoryInfo
			}
		case "security_info":
			var securityInfo SecurityCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &securityInfo); err == nil {
				data.SecurityInfo = &securityInfo
			}
		case "storage_info":
			var storageInfo StorageCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &storageInfo); err == nil {
				data.StorageInfo = &storageInfo
			}
		case "network_info":
			var networkInfo NetworkCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &networkInfo); err == nil {
				data.NetworkInfo = &networkInfo
			}
		case "virt_info":
			var virtInfo VirtualizationCapabilityInfo
			if err := json.Unmarshal([]byte(cap.DetailsJSON), &virtInfo); err == nil {
				data.VirtInfo = &virtInfo
			}
		}
	}

	return data, nil
}

// RefreshHostCapabilities performs a fresh discovery and updates stored capabilities
func (hcs *HostCapabilityService) RefreshHostCapabilities(ctx context.Context, hostID string) error {
	log.Printf("Refreshing host capabilities for host: %s", hostID)

	capabilities, err := hcs.DiscoverHostCapabilities(ctx, hostID)
	if err != nil {
		return fmt.Errorf("failed to discover capabilities: %w", err)
	}

	return hcs.StoreHostCapabilities(hostID, capabilities)
}

// PeriodicCapabilityRefresh runs periodic capability refresh for all connected hosts
func (hcs *HostCapabilityService) PeriodicCapabilityRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hcs.refreshAllHostCapabilities(ctx)
		}
	}
}

// refreshAllHostCapabilities refreshes capabilities for all connected hosts
func (hcs *HostCapabilityService) refreshAllHostCapabilities(ctx context.Context) {
	var hosts []storage.Host
	if err := hcs.db.Where("state = ?", "CONNECTED").Find(&hosts).Error; err != nil {
		log.Printf("Failed to get connected hosts: %v", err)
		return
	}

	for _, host := range hosts {
		go func(hostID string) {
			if err := hcs.RefreshHostCapabilities(ctx, hostID); err != nil {
				log.Printf("Failed to refresh capabilities for host %s: %v", hostID, err)
			}
		}(host.ID)
	}
}
