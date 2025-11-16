package aws

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestInstanceStreamProcessReservations(t *testing.T) {
	is := &InstanceStream{maxInstances: 100}
	total := 0
	res := []types.Reservation{{Instances: []types.Instance{
		{InstanceId: awssdk.String("i-1"), State: &types.InstanceState{Name: types.InstanceStateNameRunning}, PrivateIpAddress: awssdk.String("10.0.0.1"), PublicIpAddress: awssdk.String("54.1.1.1"), PrivateDnsName: awssdk.String("ip-10-0-0-1.ec2.internal"), PublicDnsName: awssdk.String("ec2-54-1-1-1.compute.amazonaws.com"), InstanceType: types.InstanceTypeT3Micro, Placement: &types.Placement{AvailabilityZone: awssdk.String("us-east-1a")}, Tags: []types.Tag{{Key: awssdk.String("Name"), Value: awssdk.String("web")}}},
		{InstanceId: awssdk.String("i-2"), State: &types.InstanceState{Name: types.InstanceStateNameRunning}, PrivateIpAddress: awssdk.String("10.0.0.2"), InstanceType: types.InstanceTypeT3Micro, Placement: &types.Placement{AvailabilityZone: awssdk.String("us-east-1b")}},
	}}}
	got := is.processReservations(res, &total)
	if len(got) != 2 {
		t.Fatalf("len")
	}
	if total != 2 {
		t.Fatalf("total")
	}
	if got[0].InstanceID != "i-1" || got[0].Name != "web" || got[0].AvailabilityZone != "us-east-1a" {
		t.Fatalf("map")
	}
}
