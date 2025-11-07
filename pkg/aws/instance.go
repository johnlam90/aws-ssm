package aws

import (
	"context"
	"fmt"
	"net"
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

	// Determine the type of identifier and build appropriate filters
	switch {
	case strings.HasPrefix(identifier, "i-"):
		// Instance ID
		filters = append(filters, types.Filter{
			Name:   aws.String("instance-id"),
			Values: []string{identifier},
		})
	case strings.Contains(identifier, ":"):
		// Tag format (Key:Value)
		parts := strings.SplitN(identifier, ":", 2)
		if len(parts) == 2 {
			filters = append(filters, types.Filter{
				Name:   aws.String(fmt.Sprintf("tag:%s", parts[0])),
				Values: []string{parts[1]},
			})
		}
	case isIPAddress(identifier):
		// IP address (public or private)
		filters = append(filters, types.Filter{
			Name:   aws.String("private-ip-address"),
			Values: []string{identifier},
		})
		// Also check public IP
		publicFilters := []types.Filter{
			{
				Name:   aws.String("ip-address"),
				Values: []string{identifier},
			},
		}
		// Try public IP search as well
		publicInstances, err := c.describeInstances(ctx, publicFilters)
		if err == nil && len(publicInstances) > 0 {
			return publicInstances, nil
		}
	case isDNSName(identifier):
		// DNS name (public or private)
		filters = append(filters, types.Filter{
			Name:   aws.String("private-dns-name"),
			Values: []string{identifier},
		})
		// Also check public DNS
		publicFilters := []types.Filter{
			{
				Name:   aws.String("dns-name"),
				Values: []string{identifier},
			},
		}
		publicInstances, err := c.describeInstances(ctx, publicFilters)
		if err == nil && len(publicInstances) > 0 {
			return publicInstances, nil
		}
	default:
		// Assume it's a name tag
		filters = append(filters, types.Filter{
			Name:   aws.String("tag:Name"),
			Values: []string{identifier},
		})
	}

	// Only return running instances by default
	filters = append(filters, types.Filter{
		Name:   aws.String("instance-state-name"),
		Values: []string{"running", "stopped", "stopping", "pending"},
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
		Values: []string{"running", "stopped", "stopping", "pending"},
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

// isIPAddress checks if the string is a valid IP address
func isIPAddress(s string) bool {
	return net.ParseIP(s) != nil
}

// isDNSName checks if the string looks like a DNS name
func isDNSName(s string) bool {
	return strings.Contains(s, ".") && (strings.Contains(s, "compute.amazonaws.com") ||
		strings.Contains(s, "compute.internal") ||
		strings.Count(s, ".") >= 2)
}
