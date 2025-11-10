package aws

import "testing"

func TestIdentifierHelpers(t *testing.T) {
	cases := []struct {
		in     string
		expect IdentifierType
	}{
		{"i-1234567890abcdef0", IdentifierTypeInstanceID},
		{"Name:web", IdentifierTypeTag},
		{"10.0.0.5", IdentifierTypeIPAddress},
		{"ip-10-0-0-5.ec2.internal", IdentifierTypeDNSName},
		{"web-server", IdentifierTypeName},
	}
	for _, c := range cases {
		info := ParseIdentifier(c.in)
		if info.Type != c.expect {
			t.Fatalf("%s expected %v got %v", c.in, c.expect, info.Type)
		}
	}
	if !IsTag("Env:prod") || IsTag("NoColon") {
		t.Fatalf("tag detection failed")
	}
	if !IsInstanceID("i-abc123456") || IsInstanceID("x-zzz") {
		t.Fatalf("instance id detection failed")
	}
	if !IsIPAddress("192.168.1.1") || IsIPAddress("999.999.1.1") {
		t.Fatalf("ip logic")
	}
	if !IsDNSName("ip-10-0-0-5.ec2.internal") || IsDNSName("notadns") {
		t.Fatalf("dns logic")
	}
}
