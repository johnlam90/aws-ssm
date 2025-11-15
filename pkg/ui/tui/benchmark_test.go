package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/johnlam90/aws-ssm/pkg/aws"
)

// BenchmarkSearchPerformance benchmarks search performance with cached vs non-cached fields
func BenchmarkSearchPerformance(b *testing.B) {
	// Create test data
	ec2Instances := generateTestEC2Instances(1000) // Large dataset
	nodeGroups := generateTestNodeGroups(500)      // Medium dataset
	
	b.Run("EC2SearchWithCache", func(b *testing.B) {
		// Precompute search fields (simulating our optimization)
		for i := range ec2Instances {
			ec2Instances[i].PrecomputeSearchFields()
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := "web"
			_ = filterEC2InstancesWithCache(ec2Instances, query)
		}
	})
	
	b.Run("EC2SearchWithoutCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := "web"
			_ = filterEC2InstancesWithoutCache(ec2Instances, query)
		}
	})
	
	b.Run("NodeGroupSearchWithCache", func(b *testing.B) {
		// Precompute search fields
		for i := range nodeGroups {
			nodeGroups[i].PrecomputeSearchFields()
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := "worker"
			_ = filterNodeGroupsWithCache(nodeGroups, query)
		}
	})
	
	b.Run("NodeGroupSearchWithoutCache", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			query := "worker"
			_ = filterNodeGroupsWithoutCache(nodeGroups, query)
		}
	})
}

// BenchmarkRenderingPerformance benchmarks rendering performance
func BenchmarkRenderingPerformance(b *testing.B) {
	ctx := context.Background()
	client := &aws.Client{} // Mock client
	config := Config{NoColor: false}
	
	model := NewModel(ctx, client, config)
	model.width = 120
	model.height = 40
	model.ready = true
	
	// Generate test data
	model.ec2Instances = generateTestEC2Instances(100)
	model.filteredEC2 = model.ec2Instances
	model.eksClusters = generateTestEKSClusters(50)
	model.filteredEKS = model.eksClusters
	model.asgs = generateTestASGs(75)
	model.filteredASGs = model.asgs
	model.nodeGroups = generateTestNodeGroups(60)
	model.filteredNodeGroups = model.nodeGroups
	
	b.Run("RenderEC2View", func(b *testing.B) {
		model.currentView = ViewEC2Instances
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.renderEC2Instances()
		}
	})
	
	b.Run("RenderEKSView", func(b *testing.B) {
		model.currentView = ViewEKSClusters
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.renderEKSClusters()
		}
	})
	
	b.Run("RenderASGView", func(b *testing.B) {
		model.currentView = ViewASGs
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.renderASGs()
		}
	})
	
	b.Run("RenderNodeGroupView", func(b *testing.B) {
		model.currentView = ViewNodeGroups
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = model.renderNodeGroups()
		}
	})
}

// BenchmarkStylePerformance benchmarks style application performance
func BenchmarkStylePerformance(b *testing.B) {
	b.Run("ModernThemeWithColors", func(b *testing.B) {
		SetTheme(NewModernTheme(true))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = TitleStyle().Render("Test Title")
			_ = StatusBarStyle().Render("Status: Active")
			_ = ErrorStyle().Render("Error Message")
			_ = SuccessStyle().Render("Success Message")
		}
	})
	
	b.Run("ModernThemeNoColors", func(b *testing.B) {
		SetTheme(NewModernTheme(false))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = TitleStyle().Render("Test Title")
			_ = StatusBarStyle().Render("Status: Active")
			_ = ErrorStyle().Render("Error Message")
			_ = SuccessStyle().Render("Success Message")
		}
	})
}

// BenchmarkSearchDebounce benchmarks search debouncing performance
func BenchmarkSearchDebounce(b *testing.B) {
	model := &Model{
		searchActive:  true,
		searchBuffer:  "",
		currentView:   ViewEC2Instances,
		searchQueries: make(map[ViewMode]string),
	}
	
	b.Run("SearchInputProcessing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate typing "web" character by character
			chars := []string{"w", "e", "b"}
			for _, char := range chars {
				// This would normally be called with a tea.KeyMsg
				model.searchBuffer = model.searchBuffer + char
			}
			model.searchBuffer = "" // Reset for next iteration
		}
	})
}

// Helper functions to generate test data

func generateTestEC2Instances(count int) []EC2Instance {
	instances := make([]EC2Instance, count)
	for i := 0; i < count; i++ {
		instances[i] = EC2Instance{
			InstanceID:       fmt.Sprintf("i-%010d", i),
			Name:             fmt.Sprintf("web-server-%d", i),
			State:            "running",
			PrivateIP:        fmt.Sprintf("10.0.%d.%d", i/256, i%256),
			PublicIP:         fmt.Sprintf("54.123.%d.%d", i/256, i%256),
			InstanceType:     "t3.medium",
			AvailabilityZone: "us-east-1a",
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "backend",
				"Application": fmt.Sprintf("app-%d", i%10),
			},
			LaunchTime:      time.Now().Add(-time.Duration(i) * time.Hour),
			InstanceProfile: "ec2-instance-profile",
			SecurityGroups:  []string{"sg-12345678", "sg-87654321"},
		}
	}
	return instances
}

func generateTestEKSClusters(count int) []EKSCluster {
	clusters := make([]EKSCluster, count)
	for i := 0; i < count; i++ {
		clusters[i] = EKSCluster{
			Name:    fmt.Sprintf("cluster-%d", i),
			Status:  "ACTIVE",
			Version: "1.21",
			Arn:     fmt.Sprintf("arn:aws:eks:us-east-1:123456789012:cluster/cluster-%d", i),
		}
	}
	return clusters
}

func generateTestASGs(count int) []ASG {
	asgs := make([]ASG, count)
	for i := 0; i < count; i++ {
		asgs[i] = ASG{
			Name:              fmt.Sprintf("asg-web-%d", i),
			DesiredCapacity:   int32(3 + i%5),
			MinSize:           int32(1 + i%3),
			MaxSize:           int32(10 + i%5),
			CurrentSize:       int32(3 + i%5),
			Status:            "Healthy",
			HealthCheckType:   "ELB",
			CreatedAt:         time.Now().Add(-time.Duration(i) * time.Hour * 24),
			AvailabilityZones: []string{"us-east-1a", "us-east-1b", "us-east-1c"},
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "infrastructure",
			},
			LaunchTemplateName:    fmt.Sprintf("lt-web-%d", i),
			LaunchTemplateVersion: "$Latest",
			LoadBalancerNames:     []string{},
			TargetGroupARNs:       []string{},
		}
	}
	return asgs
}

func generateTestNodeGroups(count int) []NodeGroup {
	nodeGroups := make([]NodeGroup, count)
	for i := 0; i < count; i++ {
		nodeGroups[i] = NodeGroup{
			ClusterName:           fmt.Sprintf("cluster-%d", i%10),
			Name:                    fmt.Sprintf("worker-group-%d", i),
			Status:                  "ACTIVE",
			Version:                 "1.21",
			InstanceTypes:           []string{"t3.medium", "t3.large"},
			DesiredSize:             int32(3 + i%5),
			MinSize:                 int32(1 + i%3),
			MaxSize:                 int32(10 + i%5),
			CurrentSize:             int32(3 + i%5),
			CreatedAt:               time.Now().Add(-time.Duration(i) * time.Hour * 24).Format("2006-01-02 15:04:05"),
			LaunchTemplateID:        fmt.Sprintf("lt-%010d", i),
			LaunchTemplateName:      fmt.Sprintf("worker-template-%d", i),
			LaunchTemplateVersion:   "1",
			Tags: map[string]string{
				"Environment": "production",
				"Team":        "kubernetes",
			},
		}
	}
	return nodeGroups
}

// Filter functions for benchmarking

func filterEC2InstancesWithCache(instances []EC2Instance, query string) []EC2Instance {
	if query == "" {
		return instances
	}
	
	query = strings.ToLower(query)
	var filtered []EC2Instance
	
	for _, inst := range instances {
		if strings.Contains(inst.cachedNameLower, query) ||
			strings.Contains(inst.cachedIDLower, query) ||
			strings.Contains(inst.cachedStateLower, query) ||
			strings.Contains(inst.cachedTypeLower, query) ||
			strings.Contains(inst.cachedPrivateIPLower, query) ||
			strings.Contains(inst.cachedPublicIPLower, query) ||
			strings.Contains(inst.cachedTagsString, query) {
			filtered = append(filtered, inst)
		}
	}
	
	return filtered
}

func filterEC2InstancesWithoutCache(instances []EC2Instance, query string) []EC2Instance {
	if query == "" {
		return instances
	}
	
	query = strings.ToLower(query)
	var filtered []EC2Instance
	
	for _, inst := range instances {
		if strings.Contains(strings.ToLower(inst.Name), query) ||
			strings.Contains(strings.ToLower(inst.InstanceID), query) ||
			strings.Contains(strings.ToLower(inst.State), query) ||
			strings.Contains(strings.ToLower(inst.InstanceType), query) ||
			strings.Contains(strings.ToLower(inst.PrivateIP), query) ||
			strings.Contains(strings.ToLower(inst.PublicIP), query) {
			filtered = append(filtered, inst)
		}
		
		// Check tags
		for k, v := range inst.Tags {
			tag := strings.ToLower(fmt.Sprintf("%s:%s", k, v))
			if strings.Contains(tag, query) {
				filtered = append(filtered, inst)
				break
			}
		}
	}
	
	return filtered
}

func filterNodeGroupsWithCache(groups []NodeGroup, query string) []NodeGroup {
	if query == "" {
		return groups
	}
	
	query = strings.ToLower(query)
	var filtered []NodeGroup
	
	for _, ng := range groups {
		if strings.Contains(ng.cachedClusterLower, query) ||
			strings.Contains(ng.cachedNameLower, query) ||
			strings.Contains(ng.cachedStatusLower, query) ||
			strings.Contains(ng.cachedVersionLower, query) ||
			strings.Contains(ng.cachedInstanceTypesLower, query) ||
			strings.Contains(ng.cachedLaunchTemplateNameLower, query) ||
			strings.Contains(ng.cachedLaunchTemplateVersionLower, query) ||
			strings.Contains(ng.cachedLaunchTemplateIDLower, query) {
			filtered = append(filtered, ng)
		}
	}
	
	return filtered
}

func filterNodeGroupsWithoutCache(groups []NodeGroup, query string) []NodeGroup {
	if query == "" {
		return groups
	}
	
	query = strings.ToLower(query)
	var filtered []NodeGroup
	
	for _, ng := range groups {
		if strings.Contains(strings.ToLower(ng.ClusterName), query) ||
			strings.Contains(strings.ToLower(ng.Name), query) ||
			strings.Contains(strings.ToLower(ng.Status), query) ||
			strings.Contains(strings.ToLower(ng.Version), query) ||
			strings.Contains(strings.ToLower(strings.Join(ng.InstanceTypes, ",")), query) ||
			strings.Contains(strings.ToLower(ng.LaunchTemplateName), query) ||
			strings.Contains(strings.ToLower(ng.LaunchTemplateVersion), query) ||
			strings.Contains(strings.ToLower(ng.LaunchTemplateID), query) {
			filtered = append(filtered, ng)
		}
	}
	
	return filtered
}