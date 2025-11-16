package aws

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestSortInterfacesByCardAndDevice(t *testing.T) {
	in := []types.InstanceNetworkInterface{
		{Attachment: &types.InstanceNetworkInterfaceAttachment{NetworkCardIndex: awssdk.Int32(1), DeviceIndex: awssdk.Int32(0)}},
		{Attachment: &types.InstanceNetworkInterfaceAttachment{NetworkCardIndex: awssdk.Int32(0), DeviceIndex: awssdk.Int32(2)}},
		{Attachment: &types.InstanceNetworkInterfaceAttachment{NetworkCardIndex: awssdk.Int32(0), DeviceIndex: awssdk.Int32(1)}},
		{Attachment: &types.InstanceNetworkInterfaceAttachment{NetworkCardIndex: awssdk.Int32(1), DeviceIndex: awssdk.Int32(1)}},
	}
	out := sortInterfacesByCardAndDevice(in)
	got := [][2]int32{}
	for _, ni := range out {
		got = append(got, [2]int32{awssdk.ToInt32(ni.Attachment.NetworkCardIndex), awssdk.ToInt32(ni.Attachment.DeviceIndex)})
	}
	want := [][2]int32{{0, 1}, {0, 2}, {1, 0}, {1, 1}}
	if len(got) != len(want) {
		t.Fatalf("len")
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("order")
		}
	}
}
