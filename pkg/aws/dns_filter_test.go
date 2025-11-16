package aws

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

func TestCreateDNSNameFilterEC2(t *testing.T) {
	f := createDNSNameFilter([]string{"ip-10-0-0-5.ec2.internal"})
	if awssdk.ToString(f.Name) != "private-dns-name" {
		t.Fatalf("name")
	}
	have := map[string]bool{}
	for _, v := range f.Values {
		have[v] = true
	}
	if !have["ip-10-0-0-5.ec2.internal"] || !have["ip-10-0-0-5.*.compute.internal"] {
		t.Fatalf("values")
	}
}

func TestCreateDNSNameFilterCompute(t *testing.T) {
	f := createDNSNameFilter([]string{"ip-10-0-0-5.compute.internal"})
	if awssdk.ToString(f.Name) != "private-dns-name" {
		t.Fatalf("name")
	}
	have := map[string]bool{}
	for _, v := range f.Values {
		have[v] = true
	}
	if !have["ip-10-0-0-5.compute.internal"] || !have["ip-10-0-0-5.ec2.internal"] {
		t.Fatalf("values")
	}
}
