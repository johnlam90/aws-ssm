package aws

import (
	"testing"
	"time"
)

func TestClusterStructure(t *testing.T) {
	tests := []struct {
		name     string
		cluster  Cluster
		validate func(*Cluster) bool
	}{
		{
			name: "basic cluster",
			cluster: Cluster{
				Name:    "test-cluster",
				Status:  "ACTIVE",
				Version: "1.27",
				ARN:     "arn:aws:eks:us-east-1:123456789012:cluster/test-cluster",
			},
			validate: func(c *Cluster) bool {
				return c.Name == "test-cluster" && c.Status == "ACTIVE" && c.Version == "1.27"
			},
		},
		{
			name: "cluster with VPC info",
			cluster: Cluster{
				Name:    "vpc-cluster",
				Status:  "ACTIVE",
				Version: "1.28",
				VPC: VPCInfo{
					VpcID:              "vpc-12345678",
					SubnetIDs:          []string{"subnet-1", "subnet-2"},
					SecurityGroupIDs:   []string{"sg-1", "sg-2"},
					EndpointPrivateAccess: true,
					EndpointPublicAccess:  true,
				},
			},
			validate: func(c *Cluster) bool {
				return c.VPC.VpcID == "vpc-12345678" && len(c.VPC.SubnetIDs) == 2
			},
		},
		{
			name: "cluster with node groups",
			cluster: Cluster{
				Name:    "ng-cluster",
				Status:  "ACTIVE",
				Version: "1.27",
				NodeGroups: []NodeGroup{
					{
						Name:        "ng-1",
						Status:      "ACTIVE",
						DesiredSize: 3,
						MinSize:     1,
						MaxSize:     5,
					},
				},
			},
			validate: func(c *Cluster) bool {
				return len(c.NodeGroups) == 1 && c.NodeGroups[0].Name == "ng-1"
			},
		},
		{
			name: "cluster with Fargate profiles",
			cluster: Cluster{
				Name:    "fargate-cluster",
				Status:  "ACTIVE",
				Version: "1.27",
				FargateProfiles: []FargateProfile{
					{
						Name:   "fp-1",
						Status: "ACTIVE",
						Selectors: []FargateSelector{
							{
								Namespace: "default",
								Labels:    map[string]string{"workload": "fargate"},
							},
						},
					},
				},
			},
			validate: func(c *Cluster) bool {
				return len(c.FargateProfiles) == 1 && c.FargateProfiles[0].Name == "fp-1"
			},
		},
		{
			name: "cluster with tags",
			cluster: Cluster{
				Name:    "tagged-cluster",
				Status:  "ACTIVE",
				Version: "1.27",
				Tags: map[string]string{
					"Environment": "production",
					"Team":        "platform",
				},
			},
			validate: func(c *Cluster) bool {
				return c.Tags["Environment"] == "production" && c.Tags["Team"] == "platform"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(&tt.cluster) {
				t.Errorf("Cluster validation failed for %s", tt.name)
			}
		})
	}
}

func TestNodeGroupStructure(t *testing.T) {
	tests := []struct {
		name      string
		nodeGroup NodeGroup
		validate  func(*NodeGroup) bool
	}{
		{
			name: "basic node group",
			nodeGroup: NodeGroup{
				Name:        "ng-1",
				Status:      "ACTIVE",
				Version:     "1.27",
				DesiredSize: 3,
				MinSize:     1,
				MaxSize:     5,
			},
			validate: func(ng *NodeGroup) bool {
				return ng.Name == "ng-1" && ng.DesiredSize == 3
			},
		},
		{
			name: "node group with instance types",
			nodeGroup: NodeGroup{
				Name:          "ng-2",
				Status:        "ACTIVE",
				InstanceTypes: []string{"t3.medium", "t3.large"},
				DiskSize:      20,
			},
			validate: func(ng *NodeGroup) bool {
				return len(ng.InstanceTypes) == 2 && ng.DiskSize == 20
			},
		},
		{
			name: "node group with taints",
			nodeGroup: NodeGroup{
				Name:   "ng-3",
				Status: "ACTIVE",
				Taints: []Taint{
					{
						Key:    "workload",
						Value:  "gpu",
						Effect: "NoSchedule",
					},
				},
			},
			validate: func(ng *NodeGroup) bool {
				return len(ng.Taints) == 1 && ng.Taints[0].Key == "workload"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(&tt.nodeGroup) {
				t.Errorf("NodeGroup validation failed for %s", tt.name)
			}
		})
	}
}

func TestFargateProfileStructure(t *testing.T) {
	tests := []struct {
		name             string
		fargateProfile   FargateProfile
		validate         func(*FargateProfile) bool
	}{
		{
			name: "basic fargate profile",
			fargateProfile: FargateProfile{
				Name:   "fp-1",
				Status: "ACTIVE",
			},
			validate: func(fp *FargateProfile) bool {
				return fp.Name == "fp-1" && fp.Status == "ACTIVE"
			},
		},
		{
			name: "fargate profile with selectors",
			fargateProfile: FargateProfile{
				Name:   "fp-2",
				Status: "ACTIVE",
				Selectors: []FargateSelector{
					{
						Namespace: "default",
						Labels:    map[string]string{"workload": "fargate"},
					},
					{
						Namespace: "kube-system",
						Labels:    map[string]string{},
					},
				},
			},
			validate: func(fp *FargateProfile) bool {
				return len(fp.Selectors) == 2 && fp.Selectors[0].Namespace == "default"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.validate(&tt.fargateProfile) {
				t.Errorf("FargateProfile validation failed for %s", tt.name)
			}
		})
	}
}

func TestVPCInfoStructure(t *testing.T) {
	vpc := VPCInfo{
		VpcID:                 "vpc-12345678",
		SubnetIDs:             []string{"subnet-1", "subnet-2", "subnet-3"},
		SecurityGroupIDs:      []string{"sg-1", "sg-2"},
		PublicAccessCIDRs:     []string{"0.0.0.0/0"},
		EndpointPrivateAccess: true,
		EndpointPublicAccess:  true,
	}

	if vpc.VpcID != "vpc-12345678" {
		t.Errorf("Expected VPC ID vpc-12345678, got %s", vpc.VpcID)
	}

	if len(vpc.SubnetIDs) != 3 {
		t.Errorf("Expected 3 subnets, got %d", len(vpc.SubnetIDs))
	}

	if len(vpc.SecurityGroupIDs) != 2 {
		t.Errorf("Expected 2 security groups, got %d", len(vpc.SecurityGroupIDs))
	}

	if !vpc.EndpointPrivateAccess || !vpc.EndpointPublicAccess {
		t.Errorf("Expected both endpoint access to be true")
	}
}

func TestLoggingInfoStructure(t *testing.T) {
	logging := LoggingInfo{
		ClusterLogging: []LoggingType{
			{Type: "api", Enabled: true},
			{Type: "audit", Enabled: true},
			{Type: "authenticator", Enabled: false},
		},
	}

	if len(logging.ClusterLogging) != 3 {
		t.Errorf("Expected 3 logging types, got %d", len(logging.ClusterLogging))
	}

	if !logging.ClusterLogging[0].Enabled {
		t.Errorf("Expected API logging to be enabled")
	}

	if logging.ClusterLogging[2].Enabled {
		t.Errorf("Expected authenticator logging to be disabled")
	}
}

func TestEncryptionConfigStructure(t *testing.T) {
	encConfig := EncryptionConfig{
		Resources: []string{"secrets"},
		Provider: EncryptionProvider{
			KeyARN: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		},
	}

	if len(encConfig.Resources) != 1 || encConfig.Resources[0] != "secrets" {
		t.Errorf("Expected resources to contain 'secrets'")
	}

	if encConfig.Provider.KeyARN == "" {
		t.Errorf("Expected KeyARN to be set")
	}
}

func TestIdentityInfoStructure(t *testing.T) {
	identity := IdentityInfo{
		OIDC: OIDCInfo{
			Issuer: "https://oidc.eks.us-east-1.amazonaws.com/id/EXAMPLEID",
		},
	}

	if identity.OIDC.Issuer == "" {
		t.Errorf("Expected OIDC issuer to be set")
	}
}

func TestClusterWithTimestamps(t *testing.T) {
	now := time.Now()
	cluster := Cluster{
		Name:      "time-cluster",
		CreatedAt: now,
	}

	if cluster.CreatedAt != now {
		t.Errorf("Expected CreatedAt to be %v, got %v", now, cluster.CreatedAt)
	}

	// Verify it's a valid time
	if cluster.CreatedAt.IsZero() {
		t.Errorf("Expected CreatedAt to be non-zero")
	}
}

