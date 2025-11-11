package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	asgtypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
)

// ListAutoScalingGroups retrieves all Auto Scaling Groups in the current region
func (c *Client) ListAutoScalingGroups(ctx context.Context) ([]string, error) {
	asgClient := autoscaling.NewFromConfig(c.Config)

	var asgNames []string
	paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(asgClient, &autoscaling.DescribeAutoScalingGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Auto Scaling Groups: %w", err)
		}

		for _, asg := range page.AutoScalingGroups {
			if asg.AutoScalingGroupName != nil {
				asgNames = append(asgNames, *asg.AutoScalingGroupName)
			}
		}
	}

	return asgNames, nil
}

// DescribeAutoScalingGroup retrieves details about a specific Auto Scaling Group
func (c *Client) DescribeAutoScalingGroup(ctx context.Context, asgName string) (*AutoScalingGroup, error) {
	asgClient := autoscaling.NewFromConfig(c.Config)

	output, err := asgClient.DescribeAutoScalingGroups(ctx, &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe Auto Scaling Group %s: %w", asgName, err)
	}

	if len(output.AutoScalingGroups) == 0 {
		return nil, fmt.Errorf("auto scaling group %s not found", asgName)
	}

	return convertAutoScalingGroup(&output.AutoScalingGroups[0]), nil
}

// UpdateAutoScalingGroupCapacity updates the capacity of an Auto Scaling Group
func (c *Client) UpdateAutoScalingGroupCapacity(ctx context.Context, asgName string, minSize, maxSize, desiredCapacity int32) error {
	asgClient := autoscaling.NewFromConfig(c.Config)

	// Validate scaling parameters
	if minSize < 0 {
		return fmt.Errorf("min size cannot be negative")
	}
	if maxSize < minSize {
		return fmt.Errorf("max size (%d) cannot be less than min size (%d)", maxSize, minSize)
	}
	if desiredCapacity < minSize || desiredCapacity > maxSize {
		return fmt.Errorf("desired capacity (%d) must be between min size (%d) and max size (%d)", desiredCapacity, minSize, maxSize)
	}

	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: &asgName,
		MinSize:              &minSize,
		MaxSize:              &maxSize,
		DesiredCapacity:      &desiredCapacity,
	}

	_, err := asgClient.UpdateAutoScalingGroup(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update Auto Scaling Group capacity: %w", err)
	}

	return nil
}

// convertAutoScalingGroup converts AWS SDK ASG type to our internal type
func convertAutoScalingGroup(asg *asgtypes.AutoScalingGroup) *AutoScalingGroup {
	if asg == nil {
		return nil
	}

	result := &AutoScalingGroup{
		Tags: make(map[string]string),
	}

	// Set basic string fields
	setStringIfNotNil(asg.AutoScalingGroupName, &result.Name)
	setStringIfNotNil(asg.AutoScalingGroupARN, &result.ARN)
	setStringIfNotNil(asg.HealthCheckType, &result.HealthCheckType)
	setStringIfNotNil(asg.VPCZoneIdentifier, &result.VPCZoneIdentifier)
	setStringIfNotNil(asg.LaunchConfigurationName, &result.LaunchConfigurationName)

	// Set integer fields
	setInt32IfNotNil(asg.MinSize, &result.MinSize)
	setInt32IfNotNil(asg.MaxSize, &result.MaxSize)
	setInt32IfNotNil(asg.DesiredCapacity, &result.DesiredCapacity)
	setInt32IfNotNil(asg.DefaultCooldown, &result.DefaultCooldown)
	setInt32IfNotNil(asg.HealthCheckGracePeriod, &result.HealthCheckGracePeriod)

	// Set time field
	if asg.CreatedTime != nil {
		result.CreatedTime = *asg.CreatedTime
	}

	// Set current size based on instance count
	result.CurrentSize = int32(len(asg.Instances))

	// Copy slice fields
	result.AvailabilityZones = asg.AvailabilityZones
	result.LoadBalancerNames = asg.LoadBalancerNames
	result.TargetGroupARNs = asg.TargetGroupARNs

	// Convert tags
	convertTagDescriptions(asg.Tags, result.Tags)

	// Convert instances
	result.Instances = convertInstances(asg.Instances)

	return result
}

// setStringIfNotNil safely sets a string field if the pointer is not nil
func setStringIfNotNil(ptr *string, target *string) {
	if ptr != nil {
		*target = *ptr
	}
}

// setInt32IfNotNil safely sets an int32 field if the pointer is not nil
func setInt32IfNotNil(ptr *int32, target *int32) {
	if ptr != nil {
		*target = *ptr
	}
}

// convertTagDescriptions converts AWS tag description slices to map
func convertTagDescriptions(tags []asgtypes.TagDescription, tagMap map[string]string) {
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			tagMap[*tag.Key] = *tag.Value
		}
	}
}

// convertInstances converts ASG instances from AWS types to internal types
func convertInstances(instances []asgtypes.Instance) []ASGInstance {
	var converted []ASGInstance
	for _, instance := range instances {
		converted = append(converted, convertInstance(instance))
	}
	return converted
}

// convertInstance converts a single ASG instance
func convertInstance(instance asgtypes.Instance) ASGInstance {
	asgInstance := ASGInstance{}

	setStringIfNotNil(instance.InstanceId, &asgInstance.InstanceID)
	setStringIfNotNil(instance.AvailabilityZone, &asgInstance.AvailabilityZone)
	setStringIfNotNil(instance.HealthStatus, &asgInstance.HealthStatus)
	setStringIfNotNil(instance.LaunchConfigurationName, &asgInstance.LaunchConfigurationName)

	if instance.LifecycleState != "" {
		asgInstance.LifecycleState = string(instance.LifecycleState)
	}

	if instance.LaunchTemplate != nil && instance.LaunchTemplate.LaunchTemplateName != nil {
		asgInstance.LaunchTemplateName = *instance.LaunchTemplate.LaunchTemplateName
	}

	setBoolIfNotNil(instance.ProtectedFromScaleIn, &asgInstance.ProtectedFromScaleIn)

	return asgInstance
}

// setBoolIfNotNil safely sets a bool field if the pointer is not nil
func setBoolIfNotNil(ptr *bool, target *bool) {
	if ptr != nil {
		*target = *ptr
	}
}
