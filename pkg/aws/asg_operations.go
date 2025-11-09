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
		return nil, fmt.Errorf("Auto Scaling Group %s not found", asgName)
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

	// Safe pointer dereferences
	if asg.AutoScalingGroupName != nil {
		result.Name = *asg.AutoScalingGroupName
	}
	if asg.AutoScalingGroupARN != nil {
		result.ARN = *asg.AutoScalingGroupARN
	}
	if asg.MinSize != nil {
		result.MinSize = *asg.MinSize
	}
	if asg.MaxSize != nil {
		result.MaxSize = *asg.MaxSize
	}
	if asg.DesiredCapacity != nil {
		result.DesiredCapacity = *asg.DesiredCapacity
	}
	if asg.DefaultCooldown != nil {
		result.DefaultCooldown = *asg.DefaultCooldown
	}
	if asg.HealthCheckType != nil {
		result.HealthCheckType = *asg.HealthCheckType
	}
	if asg.HealthCheckGracePeriod != nil {
		result.HealthCheckGracePeriod = *asg.HealthCheckGracePeriod
	}
	if asg.CreatedTime != nil {
		result.CreatedTime = *asg.CreatedTime
	}
	if asg.VPCZoneIdentifier != nil {
		result.VPCZoneIdentifier = *asg.VPCZoneIdentifier
	}

	// Current size is the number of instances
	result.CurrentSize = int32(len(asg.Instances))

	// Copy availability zones
	result.AvailabilityZones = asg.AvailabilityZones

	// Copy load balancer names
	result.LoadBalancerNames = asg.LoadBalancerNames

	// Copy target group ARNs
	result.TargetGroupARNs = asg.TargetGroupARNs

	// Handle launch template
	if asg.LaunchTemplate != nil {
		if asg.LaunchTemplate.LaunchTemplateName != nil {
			result.LaunchTemplateName = *asg.LaunchTemplate.LaunchTemplateName
		}
		if asg.LaunchTemplate.Version != nil {
			result.LaunchTemplateVersion = *asg.LaunchTemplate.Version
		}
	}

	// Handle mixed instances policy launch template
	if asg.MixedInstancesPolicy != nil && asg.MixedInstancesPolicy.LaunchTemplate != nil {
		if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification != nil {
			if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName != nil {
				result.LaunchTemplateName = *asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName
			}
			if asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version != nil {
				result.LaunchTemplateVersion = *asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version
			}
		}
	}

	// Handle launch configuration
	if asg.LaunchConfigurationName != nil {
		result.LaunchConfigurationName = *asg.LaunchConfigurationName
	}

	// Convert tags
	for _, tag := range asg.Tags {
		if tag.Key != nil && tag.Value != nil {
			result.Tags[*tag.Key] = *tag.Value
		}
	}

	// Convert instances
	for _, instance := range asg.Instances {
		asgInstance := ASGInstance{}
		if instance.InstanceId != nil {
			asgInstance.InstanceID = *instance.InstanceId
		}
		if instance.AvailabilityZone != nil {
			asgInstance.AvailabilityZone = *instance.AvailabilityZone
		}
		if instance.LifecycleState != "" {
			asgInstance.LifecycleState = string(instance.LifecycleState)
		}
		if instance.HealthStatus != nil {
			asgInstance.HealthStatus = *instance.HealthStatus
		}
		if instance.LaunchConfigurationName != nil {
			asgInstance.LaunchConfigurationName = *instance.LaunchConfigurationName
		}
		if instance.LaunchTemplate != nil && instance.LaunchTemplate.LaunchTemplateName != nil {
			asgInstance.LaunchTemplateName = *instance.LaunchTemplate.LaunchTemplateName
		}
		if instance.ProtectedFromScaleIn != nil {
			asgInstance.ProtectedFromScaleIn = *instance.ProtectedFromScaleIn
		}

		result.Instances = append(result.Instances, asgInstance)
	}

	return result
}
