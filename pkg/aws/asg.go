// Package aws contains AWS service helpers used by the CLI.
package aws

import (
	"time"
)

// AutoScalingGroup represents an AWS Auto Scaling Group
type AutoScalingGroup struct {
	Name                    string
	ARN                     string
	MinSize                 int32
	MaxSize                 int32
	DesiredCapacity         int32
	CurrentSize             int32
	DefaultCooldown         int32
	HealthCheckType         string
	HealthCheckGracePeriod  int32
	CreatedTime             time.Time
	AvailabilityZones       []string
	LoadBalancerNames       []string
	TargetGroupARNs         []string
	VPCZoneIdentifier       string
	Tags                    map[string]string
	LaunchTemplateName      string
	LaunchTemplateVersion   string
	LaunchConfigurationName string
	Instances               []ASGInstance
}

// ASGInstance represents an instance in an Auto Scaling Group
type ASGInstance struct {
	InstanceID              string
	AvailabilityZone        string
	LifecycleState          string
	HealthStatus            string
	LaunchConfigurationName string
	LaunchTemplateName      string
	ProtectedFromScaleIn    bool
}

// GetName returns the ASG name
func (asg *AutoScalingGroup) GetName() string {
	return asg.Name
}

// GetARN returns the ASG ARN
func (asg *AutoScalingGroup) GetARN() string {
	return asg.ARN
}

// GetMinSize returns the minimum size
func (asg *AutoScalingGroup) GetMinSize() int32 {
	return asg.MinSize
}

// GetMaxSize returns the maximum size
func (asg *AutoScalingGroup) GetMaxSize() int32 {
	return asg.MaxSize
}

// GetDesiredCapacity returns the desired capacity
func (asg *AutoScalingGroup) GetDesiredCapacity() int32 {
	return asg.DesiredCapacity
}

// GetCurrentSize returns the current size (number of instances)
func (asg *AutoScalingGroup) GetCurrentSize() int32 {
	return asg.CurrentSize
}

// GetCreatedTime returns the creation time
func (asg *AutoScalingGroup) GetCreatedTime() time.Time {
	return asg.CreatedTime
}

// GetTags returns the ASG tags
func (asg *AutoScalingGroup) GetTags() map[string]string {
	return asg.Tags
}

// GetLaunchTemplateName returns the launch template name
func (asg *AutoScalingGroup) GetLaunchTemplateName() string {
	return asg.LaunchTemplateName
}

// GetLaunchConfigurationName returns the launch configuration name
func (asg *AutoScalingGroup) GetLaunchConfigurationName() string {
	return asg.LaunchConfigurationName
}

// GetAvailabilityZones returns the availability zones
func (asg *AutoScalingGroup) GetAvailabilityZones() []string {
	return asg.AvailabilityZones
}

// GetHealthCheckType returns the health check type
func (asg *AutoScalingGroup) GetHealthCheckType() string {
	return asg.HealthCheckType
}

// GetInstances returns the instances in the ASG
func (asg *AutoScalingGroup) GetInstances() []ASGInstance {
	return asg.Instances
}
