package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/johnlam90/aws-ssm/pkg/logging"
)

// InstanceStream provides streaming/chunked access to large instance lists
type InstanceStream struct {
	client       *Client
	filters      []types.Filter
	pageSize     int
	maxInstances int
	logger       logging.Logger
	memoryLimit  int64 // Memory limit in bytes
}

// InstanceStreamConfig holds configuration for instance streaming
type InstanceStreamConfig struct {
	PageSize     int   // Number of instances per page
	MaxInstances int   // Maximum total instances to fetch
	MemoryLimit  int64 // Memory limit in bytes (0 = no limit)
}

// DefaultInstanceStreamConfig returns default streaming configuration
func DefaultInstanceStreamConfig() *InstanceStreamConfig {
	return &InstanceStreamConfig{
		PageSize:     100,              // Process 100 instances at a time
		MaxInstances: 10000,            // Maximum 10,000 instances
		MemoryLimit:  50 * 1024 * 1024, // 50 MB memory limit
	}
}

// NewInstanceStream creates a new instance stream
func (c *Client) NewInstanceStream(filters []types.Filter, cfg *InstanceStreamConfig) *InstanceStream {
	if cfg == nil {
		cfg = DefaultInstanceStreamConfig()
	}

	return &InstanceStream{
		client:       c,
		filters:      filters,
		pageSize:     cfg.PageSize,
		maxInstances: cfg.MaxInstances,
		memoryLimit:  cfg.MemoryLimit,
		logger:       logging.With(logging.String("component", "instance_stream")),
	}
}

// ForEach iterates over instances in chunks, calling the provided function for each chunk
// This prevents loading all instances into memory at once
func (is *InstanceStream) ForEach(ctx context.Context, fn func(instances []Instance) error) error {
	var nextToken *string
	totalFetched := 0

	is.logger.Info("Starting instance stream",
		logging.Int("page_size", is.pageSize),
		logging.Int("max_instances", is.maxInstances))

	for {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("stream cancelled: %w", ctx.Err())
		default:
		}

		// Check if we've reached the maximum number of instances
		if totalFetched >= is.maxInstances {
			is.logger.Warn("Reached maximum instance limit",
				logging.Int("total_fetched", totalFetched),
				logging.Int("limit", is.maxInstances))
			break
		}

		// Fetch and process page
		pageInstances, err := is.fetchPage(ctx, &nextToken, &totalFetched)
		if err != nil {
			return err
		}

		// Call the provided function with this chunk
		if len(pageInstances) > 0 {
			is.logger.Debug("Processing instance chunk",
				logging.Int("chunk_size", len(pageInstances)),
				logging.Int("total_fetched", totalFetched))

			if err := fn(pageInstances); err != nil {
				return fmt.Errorf("chunk processing failed: %w", err)
			}
		}

		// Check if there are more pages
		if nextToken == nil || totalFetched >= is.maxInstances {
			break
		}
	}

	is.logger.Info("Instance stream completed",
		logging.Int("total_processed", totalFetched))

	return nil
}

// fetchPage fetches a single page of instances
func (is *InstanceStream) fetchPage(ctx context.Context, nextToken **string, totalFetched *int) ([]Instance, error) {
	// Check circuit breaker before making API call
	if err := is.client.CircuitBreaker.Allow(); err != nil {
		return nil, fmt.Errorf("circuit breaker open: %w", err)
	}

	// Fetch next page
	input := &ec2.DescribeInstancesInput{
		Filters:    is.filters,
		NextToken:  *nextToken,
		MaxResults: aws.Int32(int32(is.pageSize)),
	}

	result, err := is.client.EC2Client.DescribeInstances(ctx, input)
	if err != nil {
		is.client.CircuitBreaker.RecordFailure()
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	is.client.CircuitBreaker.RecordSuccess()

	// Process instances from this page
	pageInstances := is.processReservations(result.Reservations, totalFetched)

	// Update next token
	*nextToken = result.NextToken

	return pageInstances, nil
}

// processReservations converts EC2 reservations to Instance structs
func (is *InstanceStream) processReservations(reservations []types.Reservation, totalFetched *int) []Instance {
	var instances []Instance

	for _, reservation := range reservations {
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
			*totalFetched++

			// Check if we've reached the maximum
			if *totalFetched >= is.maxInstances {
				return instances
			}
		}

		if *totalFetched >= is.maxInstances {
			break
		}
	}

	return instances
}

// Collect fetches all instances up to the maximum limit
// Use this when you need all instances in memory, but be aware of memory constraints
func (is *InstanceStream) Collect(ctx context.Context) ([]Instance, error) {
	var allInstances []Instance
	estimatedMemory := int64(0)

	err := is.ForEach(ctx, func(instances []Instance) error {
		// Estimate memory usage (rough approximation)
		// Each instance is roughly 1KB on average
		chunkMemory := int64(len(instances) * 1024)

		if is.memoryLimit > 0 && estimatedMemory+chunkMemory > is.memoryLimit {
			return fmt.Errorf("memory limit exceeded: would use %d bytes, limit is %d bytes",
				estimatedMemory+chunkMemory, is.memoryLimit)
		}

		allInstances = append(allInstances, instances...)
		estimatedMemory += chunkMemory
		return nil
	})

	if err != nil {
		return nil, err
	}

	return allInstances, nil
}

// Count returns the total number of instances matching the filters without loading them all into memory
func (is *InstanceStream) Count(ctx context.Context) (int, error) {
	count := 0

	err := is.ForEach(ctx, func(instances []Instance) error {
		count += len(instances)
		return nil
	})

	return count, err
}

// Filter applies a client-side filter to the stream
// This is useful for complex filtering that can't be done server-side
func (is *InstanceStream) Filter(ctx context.Context, predicate func(Instance) bool) ([]Instance, error) {
	var filtered []Instance
	estimatedMemory := int64(0)

	err := is.ForEach(ctx, func(instances []Instance) error {
		for _, inst := range instances {
			if predicate(inst) {
				// Estimate memory usage
				if is.memoryLimit > 0 {
					estimatedMemory += 1024 // Rough estimate: 1KB per instance
					if estimatedMemory > is.memoryLimit {
						return fmt.Errorf("memory limit exceeded during filtering: %d bytes used, limit is %d bytes",
							estimatedMemory, is.memoryLimit)
					}
				}
				filtered = append(filtered, inst)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return filtered, nil
}

// ListInstancesStreaming lists instances using streaming to handle large datasets efficiently
func (c *Client) ListInstancesStreaming(ctx context.Context, tagFilters map[string]string, cfg *InstanceStreamConfig) ([]Instance, error) {
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

	stream := c.NewInstanceStream(filters, cfg)
	return stream.Collect(ctx)
}

// FindInstancesStreaming finds instances using streaming for memory efficiency
func (c *Client) FindInstancesStreaming(ctx context.Context, identifier string, cfg *InstanceStreamConfig) ([]Instance, error) {
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
	case IdentifierTypeDNSName:
		// Try private DNS first
		filters = append(filters, types.Filter{
			Name:   aws.String("private-dns-name"),
			Values: []string{idInfo.Value},
		})
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

	stream := c.NewInstanceStream(filters, cfg)
	instances, err := stream.Collect(ctx)
	if err != nil {
		return nil, err
	}

	// For IP and DNS, try public if private didn't work
	if len(instances) == 0 && (idInfo.Type == IdentifierTypeIPAddress || idInfo.Type == IdentifierTypeDNSName) {
		var publicFilters []types.Filter

		if idInfo.Type == IdentifierTypeIPAddress {
			publicFilters = []types.Filter{
				{
					Name:   aws.String("ip-address"),
					Values: []string{idInfo.Value},
				},
			}
		} else {
			publicFilters = []types.Filter{
				{
					Name:   aws.String("dns-name"),
					Values: []string{idInfo.Value},
				},
			}
		}

		publicFilters = append(publicFilters, types.Filter{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		})

		stream = c.NewInstanceStream(publicFilters, cfg)
		return stream.Collect(ctx)
	}

	return instances, nil
}
