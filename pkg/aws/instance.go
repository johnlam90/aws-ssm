package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Instance represents an EC2 instance with its metadata
type Instance struct {
	InstanceID       string
	Name             string
	State            string
	PrivateIP        string
	PublicIP         string
	PrivateDNS       string
	PublicDNS        string
	InstanceType     string
	AvailabilityZone string
	Tags             map[string]string
}

// FindInstances queries EC2 instances based on various identifiers
func (c *Client) FindInstances(ctx context.Context, identifier string) ([]Instance, error) {
	var filters []types.Filter

	// Parse the identifier to determine its type
	idInfo := ParseIdentifier(identifier)

	// Build filters based on identifier type
	switch idInfo.Type {
	case IdentifierTypeInstanceID:
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-id"),
			Values: []string{idInfo.Value},
		})
	case IdentifierTypeTag:
		filters = append(filters, types.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", idInfo.TagKey)),
			Values: []string{idInfo.TagValue},
		})
	case IdentifierTypeIPAddress:
		// Try private IP first
		filters = append(filters, types.Filter{
			Name:   aws.String("private-ip-address"),
			Values: []string{idInfo.Value},
		})
		// Also check public IP
		publicFilters := []types.Filter{
			{
				Name:   aws.String("ip-address"),
				Values: []string{idInfo.Value},
			},
		}
		publicInstances, err := c.describeInstances(ctx, publicFilters)
		if err == nil && len(publicInstances) > 0 {
			return publicInstances, nil
		}
	case IdentifierTypeDNSName:
		// Try private DNS first
		filters = append(filters, types.Filter{
			Name:   aws.String("private-dns-name"),
			Values: []string{idInfo.Value},
		})
		// Also check public DNS
		publicFilters := []types.Filter{
			{
				Name:   aws.String("dns-name"),
				Values: []string{idInfo.Value},
			},
		}
		publicInstances, err := c.describeInstances(ctx, publicFilters)
		if err == nil && len(publicInstances) > 0 {
			return publicInstances, nil
		}
	case IdentifierTypeName:
		filters = append(filters, types.Filter{
			Name:   aws.String("tag:Name"),
			Values: []string{idInfo.Value},
		})
	}

	// Filter by running state by default
	filters = append(filters, types.Filter{
		Name:   aws.String("instance-state-name"),
		Values: []string{"running"},
	})

	return c.describeInstances(ctx, filters)
}

// ListInstances lists all EC2 instances with optional tag filters
func (c *Client) ListInstances(ctx context.Context, tagFilters map[string]string) ([]Instance, error) {
	var filters []types.Filter

	// Add tag filters if provided
	for key, value := range tagFilters {
		filters = append(filters, types.Filter{
			Name:   aws.String(fmt.Sprintf("tag:%s", key)),
			Values: []string{value},
		})
	}

	// Only return running instances by default
	filters = append(filters, types.Filter{
		Name:   aws.String("instance-state-name"),
		Values: []string{"running"},
	})

	return c.describeInstances(ctx, filters)
}

func (c *Client) describeInstances(ctx context.Context, filters []types.Filter) ([]Instance, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	result, err := c.EC2Client.DescribeInstances(ctx, input)
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
				Tags:             make(map[string]string),
			}

			// Extract tags
			for _, tag := range inst.Tags {
				key := aws.ToString(tag.Key)
				value := aws.ToString(tag.Value)
				instance.Tags[key] = value
				if key == "Name" {
					instance.Name = value
				}
			}

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// ResolveSingleInstance finds a single instance by identifier and validates it's running
// Returns an error if no instances found, multiple instances found, or instance is not running
func (c *Client) ResolveSingleInstance(ctx context.Context, identifier string) (*Instance, error) {
	instances, err := c.FindInstances(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to find instance: %w", err)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instances found matching: %s", identifier)
	}

	if len(instances) > 1 {
		return nil, &MultipleInstancesError{
			Identifier:       identifier,
			Instances:        instances,
			AllowInteractive: true,
		}
	}

	instance := instances[0]

	// Check if instance is running
	if instance.State != "running" {
		return nil, fmt.Errorf("instance %s is not running (current state: %s)", instance.InstanceID, instance.State)
	}

	return &instance, nil
}

// MultipleInstancesError is returned when multiple instances match an identifier
type MultipleInstancesError struct {
	Identifier       string
	Instances        []Instance
	AllowInteractive bool
}

func (e *MultipleInstancesError) Error() string {
	return fmt.Sprintf("multiple instances found matching '%s'", e.Identifier)
}

// FormatInstanceList returns a formatted string listing all matching instances
func (e *MultipleInstancesError) FormatInstanceList() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Found %d instances matching '%s':\n\n", len(e.Instances), e.Identifier))
	for i, inst := range e.Instances {
		name := inst.Name
		if name == "" { name = "(no name)" }
		b.WriteString(fmt.Sprintf("%d. %s - %s [%s] - %s\n", i+1, inst.InstanceID, name, inst.State, inst.PrivateIP))
	}
	if e.AllowInteractive {
		b.WriteString("\nOpening interactive selector... (Esc to cancel)\n")
	}
	return b.String()
}
