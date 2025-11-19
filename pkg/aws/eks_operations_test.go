package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
)

// MockEKSAPI is a mock implementation of EKSAPI
type MockEKSAPI struct {
	ListClustersFunc           func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	DescribeClusterFunc        func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
	ListNodegroupsFunc         func(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error)
	DescribeNodegroupFunc      func(ctx context.Context, params *eks.DescribeNodegroupInput, optFns ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error)
	ListFargateProfilesFunc    func(ctx context.Context, params *eks.ListFargateProfilesInput, optFns ...func(*eks.Options)) (*eks.ListFargateProfilesOutput, error)
	DescribeFargateProfileFunc func(ctx context.Context, params *eks.DescribeFargateProfileInput, optFns ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error)
	UpdateNodegroupConfigFunc  func(ctx context.Context, params *eks.UpdateNodegroupConfigInput, optFns ...func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error)
	UpdateNodegroupVersionFunc func(ctx context.Context, params *eks.UpdateNodegroupVersionInput, optFns ...func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error)
}

func (m *MockEKSAPI) ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	if m.ListClustersFunc != nil {
		return m.ListClustersFunc(ctx, params, optFns...)
	}
	return &eks.ListClustersOutput{}, nil
}

func (m *MockEKSAPI) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	if m.DescribeClusterFunc != nil {
		return m.DescribeClusterFunc(ctx, params, optFns...)
	}
	return &eks.DescribeClusterOutput{}, nil
}

func (m *MockEKSAPI) ListNodegroups(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
	if m.ListNodegroupsFunc != nil {
		return m.ListNodegroupsFunc(ctx, params, optFns...)
	}
	return &eks.ListNodegroupsOutput{}, nil
}

func (m *MockEKSAPI) DescribeNodegroup(ctx context.Context, params *eks.DescribeNodegroupInput, optFns ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error) {
	if m.DescribeNodegroupFunc != nil {
		return m.DescribeNodegroupFunc(ctx, params, optFns...)
	}
	return &eks.DescribeNodegroupOutput{}, nil
}

func (m *MockEKSAPI) ListFargateProfiles(ctx context.Context, params *eks.ListFargateProfilesInput, optFns ...func(*eks.Options)) (*eks.ListFargateProfilesOutput, error) {
	if m.ListFargateProfilesFunc != nil {
		return m.ListFargateProfilesFunc(ctx, params, optFns...)
	}
	return &eks.ListFargateProfilesOutput{}, nil
}

func (m *MockEKSAPI) DescribeFargateProfile(ctx context.Context, params *eks.DescribeFargateProfileInput, optFns ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error) {
	if m.DescribeFargateProfileFunc != nil {
		return m.DescribeFargateProfileFunc(ctx, params, optFns...)
	}
	return &eks.DescribeFargateProfileOutput{}, nil
}

func (m *MockEKSAPI) UpdateNodegroupConfig(ctx context.Context, params *eks.UpdateNodegroupConfigInput, optFns ...func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error) {
	if m.UpdateNodegroupConfigFunc != nil {
		return m.UpdateNodegroupConfigFunc(ctx, params, optFns...)
	}
	return &eks.UpdateNodegroupConfigOutput{}, nil
}

func (m *MockEKSAPI) UpdateNodegroupVersion(ctx context.Context, params *eks.UpdateNodegroupVersionInput, optFns ...func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error) {
	if m.UpdateNodegroupVersionFunc != nil {
		return m.UpdateNodegroupVersionFunc(ctx, params, optFns...)
	}
	return &eks.UpdateNodegroupVersionOutput{}, nil
}

func TestListClusters(t *testing.T) {
	mockAPI := &MockEKSAPI{
		ListClustersFunc: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-1", "cluster-2"},
			}, nil
		},
	}

	clusters, err := listClusters(context.Background(), mockAPI)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(clusters))
	}
}

func TestDescribeClusterBasic(t *testing.T) {
	mockAPI := &MockEKSAPI{
		DescribeClusterFunc: func(_ context.Context, params *eks.DescribeClusterInput, _ ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
			if *params.Name == "cluster-1" {
				return &eks.DescribeClusterOutput{
					Cluster: &ekstypes.Cluster{
						Name:            aws.String("cluster-1"),
						Status:          ekstypes.ClusterStatusActive,
						Version:         aws.String("1.21"),
						Endpoint:        aws.String("https://endpoint"),
						RoleArn:         aws.String("arn:role"),
						CreatedAt:       aws.Time(time.Now()),
						PlatformVersion: aws.String("eks.1"),
					},
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	t.Run("Success", func(t *testing.T) {
		cluster, err := describeClusterBasic(context.Background(), mockAPI, "cluster-1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cluster.Name != "cluster-1" {
			t.Errorf("expected cluster name cluster-1, got %s", cluster.Name)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := describeClusterBasic(context.Background(), mockAPI, "cluster-2")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestDescribeCluster(t *testing.T) {
	mockAPI := &MockEKSAPI{
		DescribeClusterFunc: func(_ context.Context, _ *eks.DescribeClusterInput, _ ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
			return &eks.DescribeClusterOutput{
				Cluster: &ekstypes.Cluster{
					Name:            aws.String("cluster-1"),
					Status:          ekstypes.ClusterStatusActive,
					Version:         aws.String("1.21"),
					Endpoint:        aws.String("https://endpoint"),
					RoleArn:         aws.String("arn:role"),
					CreatedAt:       aws.Time(time.Now()),
					PlatformVersion: aws.String("eks.1"),
				},
			}, nil
		},
		ListNodegroupsFunc: func(_ context.Context, _ *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			return &eks.ListNodegroupsOutput{
				Nodegroups: []string{"ng-1"},
			}, nil
		},
		DescribeNodegroupFunc: func(_ context.Context, _ *eks.DescribeNodegroupInput, _ ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error) {
			return &eks.DescribeNodegroupOutput{
				Nodegroup: &ekstypes.Nodegroup{
					NodegroupName: aws.String("ng-1"),
					Status:        ekstypes.NodegroupStatusActive,
				},
			}, nil
		},
		ListFargateProfilesFunc: func(_ context.Context, _ *eks.ListFargateProfilesInput, _ ...func(*eks.Options)) (*eks.ListFargateProfilesOutput, error) {
			return &eks.ListFargateProfilesOutput{
				FargateProfileNames: []string{"fp-1"},
			}, nil
		},
		DescribeFargateProfileFunc: func(_ context.Context, _ *eks.DescribeFargateProfileInput, _ ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error) {
			return &eks.DescribeFargateProfileOutput{
				FargateProfile: &ekstypes.FargateProfile{
					FargateProfileName: aws.String("fp-1"),
					Status:             ekstypes.FargateProfileStatusActive,
				},
			}, nil
		},
	}

	cluster, err := describeCluster(context.Background(), mockAPI, "cluster-1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cluster.NodeGroups) != 1 {
		t.Errorf("expected 1 node group, got %d", len(cluster.NodeGroups))
	}
	if len(cluster.FargateProfiles) != 1 {
		t.Errorf("expected 1 fargate profile, got %d", len(cluster.FargateProfiles))
	}
}

func TestUpdateNodeGroupScaling(t *testing.T) {
	mockAPI := &MockEKSAPI{
		UpdateNodegroupConfigFunc: func(_ context.Context, _ *eks.UpdateNodegroupConfigInput, _ ...func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error) {
			return &eks.UpdateNodegroupConfigOutput{}, nil
		},
	}

	t.Run("Success", func(t *testing.T) {
		err := updateNodeGroupScaling(context.Background(), mockAPI, "cluster-1", "ng-1", 1, 3, 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("InvalidParams", func(t *testing.T) {
		err := updateNodeGroupScaling(context.Background(), mockAPI, "cluster-1", "ng-1", -1, 3, 2)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUpdateNodeGroupLaunchTemplate(t *testing.T) {
	mockAPI := &MockEKSAPI{
		UpdateNodegroupVersionFunc: func(_ context.Context, _ *eks.UpdateNodegroupVersionInput, _ ...func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error) {
			return &eks.UpdateNodegroupVersionOutput{}, nil
		},
	}

	t.Run("Success", func(t *testing.T) {
		err := updateNodeGroupLaunchTemplate(context.Background(), mockAPI, "cluster-1", "ng-1", "lt-123", "1")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("EmptyVersion", func(t *testing.T) {
		err := updateNodeGroupLaunchTemplate(context.Background(), mockAPI, "cluster-1", "ng-1", "lt-123", "")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
