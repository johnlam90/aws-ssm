package tui

import (
	"testing"
)

func TestSearchFunctions(t *testing.T) {
	// Test basic search functionality
	inst := EC2Instance{
		Name:         "web-server",
		InstanceID:   "i-1234567890",
		State:        "running",
		InstanceType: "t3.micro",
	}

	// Test EC2 matching
	if !ec2MatchesQuery(inst, "web") {
		t.Error("Should match 'web' in instance name")
	}
	if !ec2MatchesQuery(inst, "running") {
		t.Error("Should match 'running' in state")
	}
	if !ec2MatchesQuery(inst, "t3") {
		t.Error("Should match 't3' in instance type")
	}
	if ec2MatchesQuery(inst, "stopped") {
		t.Error("Should not match 'stopped' when state is running")
	}
}

func TestEC2MatchesQuery(t *testing.T) {
	inst := EC2Instance{
		Name:         "web-server",
		InstanceID:   "i-1234567890abcdef0",
		PrivateIP:    "10.0.1.100",
		PublicIP:     "54.123.45.67",
		InstanceType: "t3.medium",
		State:        "running",
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "backend",
		},
	}

	tests := []struct {
		query string
		match bool
	}{
		{"web", true},
		{"i-1234567890", true},
		{"10.0.1.100", true},
		{"t3.medium", true},
		{"running", true},
		{"name:web", true},
		{"name:server", true},
		{"name:db", false},
		{"id:i-1234567890", true},
		{"id:i-999", false},
		{"privateip:10.0.1.100", true},
		{"pip:10.0.1.100", true},
		{"ip:10.0.1.100", true},
		{"type:t3", true},
		{"type:t2", false},
		{"state:running", true},
		{"state:stopped", false},
		{"tag:Environment=production", true},
		{"tag:Team=backend", true},
		{"tag:Environment=staging", false},
		{"tag:Environment", true},
		{"tag:Missing", false},
		{"name:web state:running", true},
		{"name:web state:stopped", false},
		{"type:t3 state:running", true},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := ec2MatchesQuery(inst, tt.query)
			if result != tt.match {
				t.Errorf("Query '%s' expected %v, got %v", tt.query, tt.match, result)
			}
		})
	}
}

func TestNodeGroupMatchesQuery(t *testing.T) {
	ng := NodeGroup{
		ClusterName:           "production-cluster",
		Name:                  "worker-nodes",
		Status:                "active",
		Version:               "1.21",
		InstanceTypes:         []string{"t3.medium", "t3.large"},
		LaunchTemplateName:    "worker-template",
		LaunchTemplateVersion: "1",
		LaunchTemplateID:      "lt-1234567890abcdef0",
	}

	tests := []struct {
		query string
		match bool
	}{
		{"worker", true},
		{"production", true},
		{"cluster:production", true},
		{"cluster:staging", false},
		{"name:worker", true},
		{"name:master", false},
		{"status:active", true},
		{"status:inactive", false},
		{"version:1.21", true},
		{"version:1.20", false},
		{"ltname:worker-template", true},
		{"ltname:other-template", false},
		{"ltid:lt-1234567890", true},
		{"ltid:lt-999", false},
		{"ltversion:1", true},
		{"ltversion:2", false},
		{"cluster:production name:worker", true},
		{"cluster:staging name:worker", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := nodeGroupMatchesQuery(ng, tt.query)
			if result != tt.match {
				t.Errorf("Query '%s' expected %v, got %v", tt.query, tt.match, result)
			}
		})
	}
}

func TestASGMatchesQuery(t *testing.T) {
	asg := ASG{
		Name:            "web-asg",
		Status:          "Healthy",
		DesiredCapacity: 3,
		MinSize:         1,
		MaxSize:         5,
		CurrentSize:     3,
	}

	tests := []struct {
		query string
		match bool
	}{
		{"web", true},
		{"asg", true},
		{"web-asg", true},
		{"healthy", true},
		{"3", true},
		{"1", true},
		{"5", true},
		{"999", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := asgMatchesQuery(asg, tt.query)
			if result != tt.match {
				t.Errorf("Query '%s' expected %v, got %v", tt.query, tt.match, result)
			}
		})
	}
}

func TestEKSClusterMatchesQuery(t *testing.T) {
	cluster := EKSCluster{
		Name:    "production-cluster",
		Status:  "ACTIVE",
		Version: "1.21",
		Arn:     "arn:aws:eks:us-west-2:123456789012:cluster/production-cluster",
	}

	tests := []struct {
		query string
		match bool
	}{
		{"production", true},
		{"cluster", true},
		{"production-cluster", true},
		{"active", true},
		{"1.21", true},
		{"arn:aws:eks", true},
		{"staging", false},
		{"inactive", false},
		{"1.20", false},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := eksMatchesQuery(cluster, tt.query)
			if result != tt.match {
				t.Errorf("Query '%s' expected %v, got %v", tt.query, tt.match, result)
			}
		})
	}
}
