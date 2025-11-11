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
	InterfaceName    string
	SubnetID         string
	CIDR             string
	SecurityGroup    string
	DeviceIndex      int32
	NetworkCardIndex int32
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

// getInstanceName extracts the instance name from tags
func getInstanceName(instance types.Instance) string {
	for _, tag := range instance.Tags {
		if aws.ToString(tag.Key) == "Name" {
			return aws.ToString(tag.Value)
		}
	}
	return "N/A"
}

// sortInterfacesByCardAndDevice sorts network interfaces by network card index first, then device index
func sortInterfacesByCardAndDevice(interfaces []types.InstanceNetworkInterface) []types.InstanceNetworkInterface {
	sortedInterfaces := make([]types.InstanceNetworkInterface, len(interfaces))
	copy(sortedInterfaces, interfaces)
	sort.Slice(sortedInterfaces, func(i, j int) bool {
		cardI, indexI := int32(0), int32(0)
		cardJ, indexJ := int32(0), int32(0)

		if sortedInterfaces[i].Attachment != nil {
			cardI = aws.ToInt32(sortedInterfaces[i].Attachment.NetworkCardIndex)
			indexI = aws.ToInt32(sortedInterfaces[i].Attachment.DeviceIndex)
		}
		if sortedInterfaces[j].Attachment != nil {
			cardJ = aws.ToInt32(sortedInterfaces[j].Attachment.NetworkCardIndex)
			indexJ = aws.ToInt32(sortedInterfaces[j].Attachment.DeviceIndex)
		}

		// Sort by network card first, then by device index
		if cardI != cardJ {
			return cardI < cardJ
		}
		return indexI < indexJ
	})
	return sortedInterfaces
}

// processInterface creates a NetworkInterface from an EC2 interface with ENS naming
func (c *Client) processInterface(ctx context.Context, iface types.InstanceNetworkInterface, ensName string) (NetworkInterface, error) {
	networkCardIndex := int32(0)
	deviceIndex := int32(0)
	if iface.Attachment != nil {
		networkCardIndex = aws.ToInt32(iface.Attachment.NetworkCardIndex)
		deviceIndex = aws.ToInt32(iface.Attachment.DeviceIndex)
	}

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

	return NetworkInterface{
		InterfaceName:    ensName,
		SubnetID:         subnetID,
		CIDR:             cidr,
		SecurityGroup:    securityGroup,
		DeviceIndex:      deviceIndex,
		NetworkCardIndex: networkCardIndex,
	}, nil
}

// GetInstanceInterfaces retrieves network interfaces for a specific instance
func (c *Client) GetInstanceInterfaces(ctx context.Context, instance types.Instance) (*InstanceInterfaces, error) {
	instanceID := aws.ToString(instance.InstanceId)
	dnsName := aws.ToString(instance.PrivateDnsName)
	if dnsName == "" {
		dnsName = "N/A"
	}

	instanceName := getInstanceName(instance)

	// Process network interfaces
	var interfaces []NetworkInterface

	// Sort interfaces by network card index first, then by device index
	sortedInterfaces := sortInterfacesByCardAndDevice(instance.NetworkInterfaces)

	// First pass: count interfaces per network card
	// This is needed to calculate continuous ENS naming across network cards
	interfaceCountPerCard := make(map[int32]int32)
	for _, iface := range sortedInterfaces {
		if iface.Attachment != nil {
			networkCardIndex := aws.ToInt32(iface.Attachment.NetworkCardIndex)
			interfaceCountPerCard[networkCardIndex]++
		}
	}

	// Second pass: process interfaces and calculate ENS names
	// Use a running counter to ensure continuous numbering
	ensCounter := int32(5) // Start at ens5 (first interface)

	for _, iface := range sortedInterfaces {
		ensName := fmt.Sprintf("ens%d", ensCounter)
		ensCounter++ // Increment for next interface

		netInterface, err := c.processInterface(ctx, iface, ensName)
		if err != nil {
			continue // Skip interfaces that fail to process
		}
		interfaces = append(interfaces, netInterface)
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
	// Build and resolve instance query
	input, err := c.buildInstanceQuery(ctx, opts)
	if err != nil {
		return err
	}

	// Describe instances
	result, err := c.EC2Client.DescribeInstances(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to describe instances: %w", err)
	}

	// Process and display instances
	instanceCount := c.processAndDisplayInstances(ctx, result)

	// Display summary
	displaySummary(instanceCount)

	return nil
}

// buildInstanceQuery builds the EC2 query based on options
func (c *Client) buildInstanceQuery(ctx context.Context, opts InterfacesOptions) (*ec2.DescribeInstancesInput, error) {
	var filters []types.Filter

	// Add state filter
	filters = addStateFilter(opts.ShowAll)

	// Add DNS name filter if node names provided
	if len(opts.NodeNames) > 0 {
		dnsFilter := createDNSNameFilter(opts.NodeNames)
		filters = append(filters, dnsFilter)
	}

	// Add tag filters
	filters = addTagFilters(filters, opts.FilterTags)

	// Build input
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	// Handle specific instance IDs
	if len(opts.InstanceIDs) > 0 {
		input.InstanceIds = opts.InstanceIDs
		return input, nil
	}

	// Handle single identifier
	if opts.Identifier != "" {
		return c.resolveSingleIdentifier(ctx, opts.Identifier)
	}

	return input, nil
}

// addStateFilter adds state filter based on showAll option
func addStateFilter(showAll bool) []types.Filter {
	if !showAll {
		// Default: only show running instances
		return []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		}
	}
	// Show all except terminated
	return []types.Filter{
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running", "stopped", "stopping", "pending"},
		},
	}
}

// createDNSNameFilter creates DNS name filter from node names
func createDNSNameFilter(nodeNames []string) types.Filter {
	var dnsPatterns []string
	for _, nodeName := range nodeNames {
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

	return types.Filter{
		Name:   aws.String("private-dns-name"),
		Values: dnsPatterns,
	}
}

// addTagFilters adds tag filters to existing filters
func addTagFilters(filters []types.Filter, tagFilters []string) []types.Filter {
	for _, tagFilter := range tagFilters {
		parts := strings.SplitN(tagFilter, ":", 2)
		if len(parts) == 2 {
			filters = append(filters, types.Filter{
				Name:   aws.String("tag:" + parts[0]),
				Values: []string{parts[1]},
			})
		}
	}
	return filters
}

// resolveSingleIdentifier resolves a single identifier to instance IDs
func (c *Client) resolveSingleIdentifier(ctx context.Context, identifier string) (*ec2.DescribeInstancesInput, error) {
	instances, err := c.FindInstances(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to find instance: %w", err)
	}
	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found matching: %s", identifier)
	}

	// Convert instances to ID list
	instanceIDs := make([]string, len(instances))
	for i, inst := range instances {
		instanceIDs[i] = inst.InstanceID
	}

	return &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
		Filters:     nil, // Clear filters when using specific instance IDs
	}, nil
}

// processAndDisplayInstances processes and displays network interfaces for all instances
func (c *Client) processAndDisplayInstances(ctx context.Context, result *ec2.DescribeInstancesOutput) int {
	instanceCount := 0
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Skip terminated instances
			if instance.State.Name == types.InstanceStateNameTerminated {
				continue
			}

			c.displayInstanceInterfaces(ctx, instance)
			instanceCount++
		}
	}
	return instanceCount
}

// displayInstanceInterfaces displays network interfaces for a single instance
func (c *Client) displayInstanceInterfaces(ctx context.Context, instance types.Instance) {
	instInterfaces, err := c.GetInstanceInterfaces(ctx, instance)
	if err != nil {
		fmt.Printf("Error getting interfaces for instance %s: %v\n",
			aws.ToString(instance.InstanceId), err)
		return
	}

	DisplayInstanceInterfaces(instInterfaces)
}

// displaySummary displays the final summary
func displaySummary(instanceCount int) {
	if instanceCount == 0 {
		fmt.Println("\nNo instances found matching the criteria.")
	} else {
		fmt.Printf("\n\nTotal instances displayed: %d\n", instanceCount)
	}
}
