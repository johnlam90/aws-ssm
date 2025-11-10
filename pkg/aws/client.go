package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	appconfig "github.com/johnlam90/aws-ssm/pkg/config"
)

// Ensure Client implements the fuzzy.AWSClientInterface interface
var _ fuzzyClientInterface = (*Client)(nil)

// Client is an AWS client that provides access to EC2 and SSM services
type Client struct {
	EC2Client      *ec2.Client
	SSMClient      *ssm.Client
	Config         aws.Config
	AppConfig      *appconfig.Config // Cached application config for performance
	CircuitBreaker *CircuitBreaker   // Circuit breaker for AWS API calls

	// Test hook: if set, overrides instance description logic used by FindInstances
	describeInstancesHook func(ctx context.Context, filters []types.Filter) ([]Instance, error)

	// Interactive UI flags
	InteractiveMode bool
	InteractiveCols []string
	NoColor         bool
	Width           int
	Favorites       bool
}

// fuzzyClientInterface is a private interface to avoid import cycles
type fuzzyClientInterface interface {
	GetConfig() aws.Config
	GetEC2Client() *ec2.Client
}

// GetConfig returns the AWS configuration
func (c *Client) GetConfig() aws.Config {
	return c.Config
}

// GetEC2Client returns the EC2 client
func (c *Client) GetEC2Client() *ec2.Client {
	return c.EC2Client
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

	// Load application config once for performance (cached in client)
	appCfg, err := appconfig.LoadConfig("")
	if err != nil {
		return nil, fmt.Errorf("failed to load application config: %w", err)
	}

	return &Client{
		EC2Client:             ec2.NewFromConfig(cfg),
		SSMClient:             ssm.NewFromConfig(cfg),
		Config:                cfg,
		AppConfig:             appCfg,
		CircuitBreaker:        NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		describeInstancesHook: nil,
		// Interactive UI flags
		InteractiveMode: false,
		InteractiveCols: []string{},
		NoColor:         false,
		Width:           0,
		Favorites:       false,
	}, nil
}

// NewClientWithFlags creates a new AWS client with interactive UI flags
func NewClientWithFlags(ctx context.Context, region, profile string, interactiveMode bool, interactiveCols []string, noColor bool, width int, favorites bool) (*Client, error) {
	client, err := NewClient(ctx, region, profile)
	if err != nil {
		return nil, err
	}

	client.InteractiveMode = interactiveMode
	client.InteractiveCols = interactiveCols
	client.NoColor = noColor
	client.Width = width
	client.Favorites = favorites

	return client, nil
}
