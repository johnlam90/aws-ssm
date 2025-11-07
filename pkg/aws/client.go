package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type Client struct {
	EC2Client *ec2.Client
	SSMClient *ssm.Client
	Config    aws.Config
}

// NewClient creates a new AWS client with EC2 and SSM services
func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	var opts []func(*config.LoadOptions) error

	// Set region if provided
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	} else if r := os.Getenv("AWS_REGION"); r != "" {
		opts = append(opts, config.WithRegion(r))
	}

	// Set profile if provided
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	} else if p := os.Getenv("AWS_PROFILE"); p != "" {
		opts = append(opts, config.WithSharedConfigProfile(p))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &Client{
		EC2Client: ec2.NewFromConfig(cfg),
		SSMClient: ssm.NewFromConfig(cfg),
		Config:    cfg,
	}, nil
}

