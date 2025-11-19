package aws

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// LaunchTemplateVersion represents a launch template version
type LaunchTemplateVersion struct {
	LaunchTemplateID   string
	LaunchTemplateName string
	VersionNumber      int64
	VersionDescription string
	CreateTime         string
	CreatedBy          string
	DefaultVersion     bool
	IsLatest           bool
}

// GetLaunchTemplateID returns the launch template ID
func (v *LaunchTemplateVersion) GetLaunchTemplateID() string {
	return v.LaunchTemplateID
}

// GetLaunchTemplateName returns the launch template name
func (v *LaunchTemplateVersion) GetLaunchTemplateName() string {
	return v.LaunchTemplateName
}

// GetVersionNumber returns the version number
func (v *LaunchTemplateVersion) GetVersionNumber() int64 {
	return v.VersionNumber
}

// GetVersionDescription returns the version description
func (v *LaunchTemplateVersion) GetVersionDescription() string {
	return v.VersionDescription
}

// GetCreateTime returns the creation time
func (v *LaunchTemplateVersion) GetCreateTime() string {
	return v.CreateTime
}

// GetCreatedBy returns the creator
func (v *LaunchTemplateVersion) GetCreatedBy() string {
	return v.CreatedBy
}

// GetDefaultVersion returns whether this is the default version
func (v *LaunchTemplateVersion) GetDefaultVersion() bool {
	return v.DefaultVersion
}

// EC2LaunchTemplateAPI defines the interface for EC2 launch template operations
type EC2LaunchTemplateAPI interface {
	DescribeLaunchTemplateVersions(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
}

// ListLaunchTemplateVersions retrieves all versions for a launch template
func (c *Client) ListLaunchTemplateVersions(ctx context.Context, launchTemplateID string) ([]LaunchTemplateVersion, error) {
	// Use the existing client if available, otherwise create a new one (fallback)
	var api EC2LaunchTemplateAPI
	if c.EC2Client != nil {
		api = c.EC2Client
	} else {
		api = ec2.NewFromConfig(c.Config)
	}
	return listLaunchTemplateVersions(ctx, api, launchTemplateID)
}

func listLaunchTemplateVersions(ctx context.Context, api EC2LaunchTemplateAPI, launchTemplateID string) ([]LaunchTemplateVersion, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: &launchTemplateID,
	}

	var versions []LaunchTemplateVersion
	paginator := ec2.NewDescribeLaunchTemplateVersionsPaginator(api, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list launch template versions: %w", err)
		}

		for _, v := range page.LaunchTemplateVersions {
			version := convertLaunchTemplateVersion(v)
			versions = append(versions, version)
		}
	}

	return versions, nil
}

// GetLaunchTemplateVersion retrieves a specific version of a launch template
func (c *Client) GetLaunchTemplateVersion(ctx context.Context, launchTemplateID, version string) (*LaunchTemplateVersion, error) {
	// Use the existing client if available, otherwise create a new one (fallback)
	var api EC2LaunchTemplateAPI
	if c.EC2Client != nil {
		api = c.EC2Client
	} else {
		api = ec2.NewFromConfig(c.Config)
	}
	return getLaunchTemplateVersion(ctx, api, launchTemplateID, version)
}

func getLaunchTemplateVersion(ctx context.Context, api EC2LaunchTemplateAPI, launchTemplateID, version string) (*LaunchTemplateVersion, error) {
	input := &ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: &launchTemplateID,
		Versions:         []string{version},
	}

	output, err := api.DescribeLaunchTemplateVersions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get launch template version: %w", err)
	}

	if len(output.LaunchTemplateVersions) == 0 {
		return nil, fmt.Errorf("launch template version %s not found", version)
	}

	ltVersion := convertLaunchTemplateVersion(output.LaunchTemplateVersions[0])
	return &ltVersion, nil
}

// convertLaunchTemplateVersion converts AWS SDK launch template version to our type
func convertLaunchTemplateVersion(v ec2types.LaunchTemplateVersion) LaunchTemplateVersion {
	version := LaunchTemplateVersion{}

	if v.LaunchTemplateId != nil {
		version.LaunchTemplateID = *v.LaunchTemplateId
	}
	if v.LaunchTemplateName != nil {
		version.LaunchTemplateName = *v.LaunchTemplateName
	}
	if v.VersionNumber != nil {
		version.VersionNumber = *v.VersionNumber
	}
	if v.VersionDescription != nil {
		version.VersionDescription = *v.VersionDescription
	}
	if v.CreateTime != nil {
		version.CreateTime = v.CreateTime.String()
	}
	if v.CreatedBy != nil {
		version.CreatedBy = *v.CreatedBy
	}
	if v.DefaultVersion != nil {
		version.DefaultVersion = *v.DefaultVersion
	}

	return version
}

// FormatVersionForDisplay formats a version number for display
func FormatVersionForDisplay(versionNumber int64, isDefault bool) string {
	versionStr := strconv.FormatInt(versionNumber, 10)
	if isDefault {
		return fmt.Sprintf("%s (Default)", versionStr)
	}
	return versionStr
}
