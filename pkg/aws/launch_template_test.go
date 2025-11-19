package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// MockEC2LaunchTemplateAPI is a mock implementation of EC2LaunchTemplateAPI
type MockEC2LaunchTemplateAPI struct {
	DescribeLaunchTemplateVersionsFunc func(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error)
}

func (m *MockEC2LaunchTemplateAPI) DescribeLaunchTemplateVersions(ctx context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
	if m.DescribeLaunchTemplateVersionsFunc != nil {
		return m.DescribeLaunchTemplateVersionsFunc(ctx, params, optFns...)
	}
	return nil, nil
}

func TestLaunchTemplateVersion_Getters(t *testing.T) {
	lt := LaunchTemplateVersion{
		LaunchTemplateID:   "lt-123",
		LaunchTemplateName: "test-template",
		VersionNumber:      1,
		VersionDescription: "v1",
		CreateTime:         "2023-01-01T00:00:00Z",
		CreatedBy:          "user",
		DefaultVersion:     true,
		IsLatest:           true,
	}

	if lt.GetLaunchTemplateID() != "lt-123" {
		t.Error("GetLaunchTemplateID failed")
	}
	if lt.GetLaunchTemplateName() != "test-template" {
		t.Error("GetLaunchTemplateName failed")
	}
	if lt.GetVersionNumber() != 1 {
		t.Error("GetVersionNumber failed")
	}
	if lt.GetVersionDescription() != "v1" {
		t.Error("GetVersionDescription failed")
	}
	if lt.GetCreateTime() != "2023-01-01T00:00:00Z" {
		t.Error("GetCreateTime failed")
	}
	if lt.GetCreatedBy() != "user" {
		t.Error("GetCreatedBy failed")
	}
	if !lt.GetDefaultVersion() {
		t.Error("GetDefaultVersion failed")
	}
}

func TestFormatVersionForDisplay(t *testing.T) {
	if got := FormatVersionForDisplay(1, true); got != "1 (Default)" {
		t.Errorf("expected '1 (Default)', got '%s'", got)
	}
	if got := FormatVersionForDisplay(2, false); got != "2" {
		t.Errorf("expected '2', got '%s'", got)
	}
}

func TestListLaunchTemplateVersions(t *testing.T) {
	mockAPI := &MockEC2LaunchTemplateAPI{
		DescribeLaunchTemplateVersionsFunc: func(_ context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
			if *params.LaunchTemplateId != "lt-123" {
				return nil, errors.New("invalid template id")
			}
			return &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
					{
						LaunchTemplateId:   aws.String("lt-123"),
						VersionNumber:      aws.Int64(1),
						VersionDescription: aws.String("v1"),
						DefaultVersion:     aws.Bool(true),
						CreateTime:         aws.Time(time.Now()),
					},
				},
			}, nil
		},
	}

	versions, err := listLaunchTemplateVersions(context.Background(), mockAPI, "lt-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(versions))
	}
	if versions[0].VersionNumber != 1 {
		t.Errorf("expected version 1, got %d", versions[0].VersionNumber)
	}
}

func TestGetLaunchTemplateVersion(t *testing.T) {
	mockAPI := &MockEC2LaunchTemplateAPI{
		DescribeLaunchTemplateVersionsFunc: func(_ context.Context, params *ec2.DescribeLaunchTemplateVersionsInput, _ ...func(*ec2.Options)) (*ec2.DescribeLaunchTemplateVersionsOutput, error) {
			if *params.LaunchTemplateId != "lt-123" {
				return nil, errors.New("invalid template id")
			}
			if len(params.Versions) > 0 && params.Versions[0] == "1" {
				return &ec2.DescribeLaunchTemplateVersionsOutput{
					LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{
						{
							LaunchTemplateId:   aws.String("lt-123"),
							VersionNumber:      aws.Int64(1),
							VersionDescription: aws.String("v1"),
						},
					},
				}, nil
			}
			return &ec2.DescribeLaunchTemplateVersionsOutput{
				LaunchTemplateVersions: []ec2types.LaunchTemplateVersion{},
			}, nil
		},
	}

	// Test found
	v, err := getLaunchTemplateVersion(context.Background(), mockAPI, "lt-123", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.VersionNumber != 1 {
		t.Errorf("expected version 1, got %d", v.VersionNumber)
	}

	// Test not found
	_, err = getLaunchTemplateVersion(context.Background(), mockAPI, "lt-123", "999")
	if err == nil {
		t.Error("expected error for non-existent version")
	}
}

func TestConvertLaunchTemplateVersion(t *testing.T) {
	now := time.Now()
	sdkVersion := ec2types.LaunchTemplateVersion{
		LaunchTemplateId:   aws.String("lt-123"),
		LaunchTemplateName: aws.String("test"),
		VersionNumber:      aws.Int64(1),
		VersionDescription: aws.String("desc"),
		CreateTime:         &now,
		CreatedBy:          aws.String("user"),
		DefaultVersion:     aws.Bool(true),
	}

	v := convertLaunchTemplateVersion(sdkVersion)

	if v.LaunchTemplateID != "lt-123" {
		t.Error("LaunchTemplateID mismatch")
	}
	if v.LaunchTemplateName != "test" {
		t.Error("LaunchTemplateName mismatch")
	}
	if v.VersionNumber != 1 {
		t.Error("VersionNumber mismatch")
	}
	if v.VersionDescription != "desc" {
		t.Error("VersionDescription mismatch")
	}
	if v.CreateTime != now.String() {
		t.Error("CreateTime mismatch")
	}
	if v.CreatedBy != "user" {
		t.Error("CreatedBy mismatch")
	}
	if !v.DefaultVersion {
		t.Error("DefaultVersion mismatch")
	}
}
