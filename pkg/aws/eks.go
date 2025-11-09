package aws

import (
	"time"
)

// Cluster represents an EKS cluster with its metadata
type Cluster struct {
	Name                 string
	ARN                  string
	Status               string
	Version              string
	Endpoint             string
	RoleARN              string
	CreatedAt            time.Time
	Tags                 map[string]string
	VPC                  VPCInfo
	Logging              LoggingInfo
	NodeGroups           []NodeGroup
	FargateProfiles      []FargateProfile
	EncryptionConfig     []EncryptionConfig
	Identity             IdentityInfo
	PlatformVersion      string
	CertificateAuthority CertificateAuthority
}

// VPCInfo contains VPC and networking configuration
type VPCInfo struct {
	VpcID                 string
	SubnetIDs             []string
	SecurityGroupIDs      []string
	PublicAccessCIDRs     []string
	EndpointPrivateAccess bool
	EndpointPublicAccess  bool
}

// LoggingInfo contains cluster logging configuration
type LoggingInfo struct {
	ClusterLogging []LoggingType
}

// LoggingType represents a logging type configuration
type LoggingType struct {
	Type    string // api, audit, authenticator, controllerManager, scheduler
	Enabled bool
}

// NodeGroup represents an EKS managed node group
type NodeGroup struct {
	Name           string
	NodeGroupARN   string
	Status         string
	Version        string
	InstanceTypes  []string
	DiskSize       int32
	DesiredSize    int32
	MinSize        int32
	MaxSize        int32
	CurrentSize    int32
	CreatedAt      time.Time
	Tags           map[string]string
	LaunchTemplate LaunchTemplateInfo
	SubnetIDs      []string
	RemoteAccess   RemoteAccessConfig
	Labels         map[string]string
	Taints         []Taint
}

// LaunchTemplateInfo contains launch template information
type LaunchTemplateInfo struct {
	ID      string
	Name    string
	Version string
}

// RemoteAccessConfig contains remote access configuration
type RemoteAccessConfig struct {
	EC2SshKeyName          string
	SourceSecurityGroupIDs []string
}

// Taint represents a Kubernetes taint
type Taint struct {
	Key    string
	Value  string
	Effect string
}

// GetName returns the node group name
func (ng *NodeGroup) GetName() string {
	return ng.Name
}

// GetStatus returns the node group status
func (ng *NodeGroup) GetStatus() string {
	return ng.Status
}

// GetVersion returns the node group version
func (ng *NodeGroup) GetVersion() string {
	return ng.Version
}

// GetInstanceTypes returns the node group instance types
func (ng *NodeGroup) GetInstanceTypes() []string {
	return ng.InstanceTypes
}

// GetDesiredSize returns the node group desired size
func (ng *NodeGroup) GetDesiredSize() int32 {
	return ng.DesiredSize
}

// GetMinSize returns the node group min size
func (ng *NodeGroup) GetMinSize() int32 {
	return ng.MinSize
}

// GetMaxSize returns the node group max size
func (ng *NodeGroup) GetMaxSize() int32 {
	return ng.MaxSize
}

// GetCurrentSize returns the node group current size
func (ng *NodeGroup) GetCurrentSize() int32 {
	return ng.CurrentSize
}

// GetCreatedAt returns the node group creation time
func (ng *NodeGroup) GetCreatedAt() time.Time {
	return ng.CreatedAt
}

// GetTags returns the node group tags
func (ng *NodeGroup) GetTags() map[string]string {
	return ng.Tags
}

// GetLaunchTemplateName returns the node group launch template name
func (ng *NodeGroup) GetLaunchTemplateName() string {
	return ng.LaunchTemplate.Name
}

// FargateProfile represents an EKS Fargate profile
type FargateProfile struct {
	Name                string
	FargateProfileARN   string
	Status              string
	CreatedAt           time.Time
	Selectors           []FargateSelector
	Tags                map[string]string
	SubnetIDs           []string
	PodExecutionRoleARN string
}

// FargateSelector represents a Fargate profile selector
type FargateSelector struct {
	Namespace string
	Labels    map[string]string
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Resources []string
	Provider  EncryptionProvider
}

// EncryptionProvider represents an encryption provider
type EncryptionProvider struct {
	KeyARN string
}

// IdentityInfo contains identity provider information
type IdentityInfo struct {
	OIDC OIDCInfo
}

// OIDCInfo contains OIDC provider information
type OIDCInfo struct {
	Issuer string
}

// CertificateAuthority contains certificate authority information
type CertificateAuthority struct {
	Data string
}
