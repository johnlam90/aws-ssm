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

// Client is an AWS client that provides access to EC2 and SSM services
type Client struct {
	EC2Client *ec2.Client
	SSMClient *ssm.Client
	Config    aws.Config
}

// NewClient creates a new AWS client with EC2 and SSM services
func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	var opts []func(*config.LoadOptions) error

	// Set region if provided
	switch {
	case region != "":
		opts = append(opts, config.WithRegion(region))
	case os.Getenv("AWS_REGION") != "":
		opts = append(opts, config.WithRegion(os.Getenv("AWS_REGION")))
	}

	// Set profile if provided
	switch {
	case profile != "":
		opts = append(opts, config.WithSharedConfigProfile(profile))
	case os.Getenv("AWS_PROFILE") != "":
		opts = append(opts, config.WithSharedConfigProfile(os.Getenv("AWS_PROFILE")))
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
