package aws

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// InterfacesOptions contains options for listing network interfaces
type InterfacesOptions struct {
	Identifier  string   // Single instance identifier (ID, name, DNS, IP, tag)
	NodeNames   []string // Kubernetes node DNS names
	InstanceIDs []string // Specific instance IDs
	FilterTags  []string // Tag filters in Key:Value format
	ShowAll     bool     // Show all instances including stopped
}

// NetworkInterface represents a network interface with its details
type NetworkInterface struct {
	InterfaceName string
	SubnetID      string
	CIDR          string
	SecurityGroup string
	DeviceIndex   int32
}

// InstanceInterfaces represents an instance with its network interfaces
type InstanceInterfaces struct {
	InstanceID   string
	DNSName      string
	InstanceName string
	Interfaces   []NetworkInterface
}

// GetSubnetCIDR retrieves the CIDR block for a given subnet ID
func (c *Client) GetSubnetCIDR(ctx context.Context, subnetID string) (string, error) {
	if subnetID == "" {
		return "N/A", nil
	}

	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{subnetID},
	}

	result, err := c.EC2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return "N/A", fmt.Errorf("failed to describe subnet: %w", err)
	}

	if len(result.Subnets) > 0 {
		return aws.ToString(result.Subnets[0].CidrBlock), nil
	}

	return "N/A", nil
}

// GetInstanceInterfaces retrieves network interfaces for a specific instance
func (c *Client) GetInstanceInterfaces(ctx context.Context, instance types.Instance) (*InstanceInterfaces, error) {
	instanceID := aws.ToString(instance.InstanceId)
	dnsName := aws.ToString(instance.PrivateDnsName)
	if dnsName == "" {
		dnsName = "N/A"
	}

	// Get instance name from tags
	instanceName := "N/A"
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			instanceName = aws.ToString(tag.Value)
			break
		}
	}

	// Process network interfaces
	var interfaces []NetworkInterface

	// Sort interfaces by device index
	sortedInterfaces := make([]types.InstanceNetworkInterface, len(instance.NetworkInterfaces))
	copy(sortedInterfaces, instance.NetworkInterfaces)
	sort.Slice(sortedInterfaces, func(i, j int) bool {
		indexI := int32(0)
		indexJ := int32(0)
		if sortedInterfaces[i].Attachment != nil {
			indexI = aws.ToInt32(sortedInterfaces[i].Attachment.DeviceIndex)
		}
		if sortedInterfaces[j].Attachment != nil {
			indexJ = aws.ToInt32(sortedInterfaces[j].Attachment.DeviceIndex)
		}
		return indexI < indexJ
	})

	for _, iface := range sortedInterfaces {
		deviceIndex := int32(0)
		if iface.Attachment != nil {
			deviceIndex = aws.ToInt32(iface.Attachment.DeviceIndex)
		}

		// For Amazon Linux 2023, interfaces start at ens5
		// ens5 = device index 0, ens6 = device index 1, etc.
		interfaceName := fmt.Sprintf("ens%d", deviceIndex+5)

		subnetID := aws.ToString(iface.SubnetId)
		if subnetID == "" {
			subnetID = "N/A"
		}

		// Get CIDR block for the subnet
		cidr := "N/A"
		if subnetID != "N/A" {
			cidrBlock, err := c.GetSubnetCIDR(ctx, subnetID)
			if err == nil {
				cidr = cidrBlock
			}
		}

		// Get first security group ID
		securityGroup := "N/A"
		if len(iface.Groups) > 0 {
			securityGroup = aws.ToString(iface.Groups[0].GroupId)
		}

		interfaces = append(interfaces, NetworkInterface{
			InterfaceName: interfaceName,
			SubnetID:      subnetID,
			CIDR:          cidr,
			SecurityGroup: securityGroup,
			DeviceIndex:   deviceIndex,
		})
	}

	return &InstanceInterfaces{
		InstanceID:   instanceID,
		DNSName:      dnsName,
		InstanceName: instanceName,
		Interfaces:   interfaces,
	}, nil
}

// DisplayInstanceInterfaces prints the network interfaces for an instance
func DisplayInstanceInterfaces(instInterfaces *InstanceInterfaces) {
	fmt.Printf("\nInstance: %s | DNS Name: %s | Instance Name: %s\n",
		instInterfaces.InstanceID,
		instInterfaces.DNSName,
		instInterfaces.InstanceName)
	fmt.Printf("%-10s| %-26s| %-19s| SG ID\n", "Interface", "Subnet ID", "CIDR")
	fmt.Println(strings.Repeat("-", 86))

	if len(instInterfaces.Interfaces) == 0 {
		fmt.Println("No network interfaces found")
		return
	}

	for _, iface := range instInterfaces.Interfaces {
		fmt.Printf("%-10s| %-26s| %-19s| %s\n",
			iface.InterfaceName,
			iface.SubnetID,
			iface.CIDR,
			iface.SecurityGroup)
	}
}

// ListNetworkInterfaces lists network interfaces for EC2 instances based on options
func (c *Client) ListNetworkInterfaces(ctx context.Context, opts InterfacesOptions) error {
	var filters []types.Filter

	// Build filters based on options
	if !opts.ShowAll {
		// Default: only show running instances
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		})
	} else {
		// Show all except terminated
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running", "stopped", "stopping", "pending"},
		})
	}

	// Handle node names (DNS names)
	if len(opts.NodeNames) > 0 {
		var dnsPatterns []string
		for _, nodeName := range opts.NodeNames {
			// Add exact match
			dnsPatterns = append(dnsPatterns, nodeName)

			// Support both short and long DNS formats
			if strings.Contains(nodeName, ".ec2.internal") {
				baseName := strings.Replace(nodeName, ".ec2.internal", "", 1)
				dnsPatterns = append(dnsPatterns, baseName+".*.compute.internal")
			} else if strings.Contains(nodeName, ".compute.internal") {
				baseName := strings.Split(nodeName, ".")[0]
				dnsPatterns = append(dnsPatterns, baseName+".ec2.internal")
			}
		}
		filters = append(filters, types.Filter{
			Name:   aws.String("private-dns-name"),
			Values: dnsPatterns,
		})
	}

	// Handle tag filters
	if len(opts.FilterTags) > 0 {
		for _, tagFilter := range opts.FilterTags {
			parts := strings.SplitN(tagFilter, ":", 2)
			if len(parts) == 2 {
				filters = append(filters, types.Filter{
					Name:   aws.String("tag:" + parts[0]),
					Values: []string{parts[1]},
				})
			}
		}
	}

	// Build input
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	// If specific instance IDs are provided, use them
	if len(opts.InstanceIDs) > 0 {
		input.InstanceIds = opts.InstanceIDs
	}

	// If a single identifier is provided, find the instance first
	var instances []Instance
	var err error

	if opts.Identifier != "" {
		instances, err = c.FindInstances(ctx, opts.Identifier)
		if err != nil {
			return fmt.Errorf("failed to find instance: %w", err)
		}
		if len(instances) == 0 {
			return fmt.Errorf("no instances found matching: %s", opts.Identifier)
		}
		// Use the found instance IDs
		input.InstanceIds = make([]string, len(instances))
		for i, inst := range instances {
			input.InstanceIds[i] = inst.InstanceID
		}
		input.Filters = nil // Clear filters when using specific instance IDs
	}

	// Describe instances
	result, err := c.EC2Client.DescribeInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	// Process and display instances
	instanceCount := 0
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Skip terminated instances
			if instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}

			instInterfaces, err := c.GetInstanceInterfaces(ctx, instance)
			if err != nil {
				fmt.Printf("Error getting interfaces for instance %s: %v\n",
					aws.ToString(instance.InstanceId), err)
				continue
			}

			DisplayInstanceInterfaces(instInterfaces)
			instanceCount++
		}
	}

	if instanceCount == 0 {
		fmt.Println("\nNo instances found matching the criteria.")
	} else {
		fmt.Printf("\n\nTotal instances displayed: %d\n", instanceCount)
	}

	return nil
}
