package fuzzy

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AWSClientInterface defines the interface for AWS client operations
type AWSClientInterface interface {
	GetConfig() aws.Config
	GetEC2Client() *ec2.Client
}

// AWSInstanceLoader implements InstanceLoader interface using the AWS client interface
type AWSInstanceLoader struct {
	client        AWSClientInterface
	regions       []string
	currentRegion string
}

// NewAWSInstanceLoader creates a new AWS instance loader
func NewAWSInstanceLoader(client AWSClientInterface) *AWSInstanceLoader {
	return &AWSInstanceLoader{
		client:        client,
		regions:       []string{client.GetConfig().Region}, // Will be expanded in Phase 3
		currentRegion: client.GetConfig().Region,
	}
}

// LoadInstances loads instances from AWS with filtering support
// Returns instances directly for better performance (no channel overhead)
func (l *AWSInstanceLoader) LoadInstances(ctx context.Context, query *SearchQuery) ([]Instance, error) {
	// Build AWS filters from search query
	awsFilters, err := l.buildAWSFilters(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build AWS filters: %w", err)
	}

	// Use describeInstances to get instances
	instances, err := l.describeInstances(ctx, awsFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	return instances, nil
}

// LoadInstance loads a single instance by ID
func (l *AWSInstanceLoader) LoadInstance(ctx context.Context, instanceID string) (*Instance, error) {
	// Build filter for specific instance ID
	awsFilters := []types.Filter{
		{
			Name:   aws.String("instance-id"),
			Values: []string{instanceID},
		},
	}

	instances, err := l.describeInstances(ctx, awsFilters)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}

	// Convert the single instance directly
	fuzzyInst := Instance{
		InstanceID:       instances[0].InstanceID,
		Name:             instances[0].Name,
		State:            instances[0].State,
		PrivateIP:        instances[0].PrivateIP,
		PublicIP:         instances[0].PublicIP,
		PrivateDNS:       instances[0].PrivateDNS,
		PublicDNS:        instances[0].PublicDNS,
		InstanceType:     instances[0].InstanceType,
		AvailabilityZone: instances[0].AvailabilityZone,
		Tags:             instances[0].Tags,
	}
	return &fuzzyInst, nil
}

// describeInstances uses the AWS client to describe instances
func (l *AWSInstanceLoader) describeInstances(ctx context.Context, filters []types.Filter) ([]Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	result, err := l.client.GetEC2Client().DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []Instance
	for _, reservation := range result.Reservations {
		for _, inst := range reservation.Instances {
			instance := Instance{
				InstanceID:       aws.ToString(inst.InstanceId),
				State:            string(inst.State.Name),
				PrivateIP:        aws.ToString(inst.PrivateIpAddress),
				PublicIP:         aws.ToString(inst.PublicIpAddress),
				PrivateDNS:       aws.ToString(inst.PrivateDnsName),
				PublicDNS:        aws.ToString(inst.PublicDnsName),
				InstanceType:     string(inst.InstanceType),
				AvailabilityZone: aws.ToString(inst.Placement.AvailabilityZone),
				LaunchTime:       aws.ToTime(inst.LaunchTime),
				Tags:             make(map[string]string),
			}

			// Extract tags
			for _, tag := range inst.Tags {
				key := aws.ToString(tag.Key)
				value := aws.ToString(tag.Value)
				instance.Tags[key] = value

				// Extract Name tag to Name field
				if key == "Name" {
					instance.Name = value
				}
			}

			// Extract security groups
			for _, sg := range inst.SecurityGroups {
				instance.SecurityGroups = append(instance.SecurityGroups, aws.ToString(sg.GroupId))
			}

			// Extract instance profile
			if inst.IamInstanceProfile != nil {
				instance.InstanceProfile = aws.ToString(inst.IamInstanceProfile.Arn)
			}

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// GetRegions returns available regions
func (l *AWSInstanceLoader) GetRegions() []string {
	return l.regions
}

// GetCurrentRegion returns the current region
func (l *AWSInstanceLoader) GetCurrentRegion() string {
	return l.currentRegion
}

// buildAWSFilters converts search query to AWS EC2 filters
func (l *AWSInstanceLoader) buildAWSFilters(query *SearchQuery) ([]types.Filter, error) {
	var filters []types.Filter

	// Add exact filters
	for key, value := range query.Filters {
		switch key {
		case "name":
			filters = append(filters, types.Filter{
				Name:   aws.String("tag:Name"),
				Values: []string{value},
			})
		case "instance-id":
			filters = append(filters, types.Filter{
				Name:   aws.String("instance-id"),
				Values: []string{value},
			})
		}
	}

	// Add tag filters
	for tagKey, tagValue := range query.TagFilters {
		filters = append(filters, types.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", tagKey)),
			Values: []string{tagValue},
		})
	}

	// Add state filter
	if query.StateFilter != "" {
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []string{query.StateFilter},
		})
	} else {
		// Default to running instances only if no state specified
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		})
	}

	// Add type filter
	if query.TypeFilter != "" {
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-type"),
			Values: []string{query.TypeFilter},
		})
	}

	// Add AZ filter
	if query.AZFilter != "" {
		filters = append(filters, types.Filter{
			Name:   aws.String("availability-zone"),
			Values: []string{query.AZFilter},
		})
	}

	// Add IP filters (private and public)
	for _, ipPattern := range query.IPFilters {
		// Note: AWS doesn't support pattern matching in IP filters
		// This would need post-filtering for wildcards
		if !containsWildcard(ipPattern) {
			filters = append(filters, types.Filter{
				Name:   aws.String("private-ip-address"),
				Values: []string{ipPattern},
			})
		}
	}

	// Add DNS filters
	for _, dnsPattern := range query.DNSFilters {
		if !containsWildcard(dnsPattern) {
			filters = append(filters, types.Filter{
				Name:   aws.String("private-dns-name"),
				Values: []string{dnsPattern},
			})
		}
	}

	return filters, nil
}

// containsWildcard checks if a string contains wildcard characters
func containsWildcard(s string) bool {
	return strings.Contains(s, "*") || strings.Contains(s, "?")
}
