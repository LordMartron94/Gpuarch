package main

import "testing"

func TestVulkanSpecIRMemberCTypeGoVkBool32(t *testing.T) {
	got, ok := vulkanSpecIRMemberCTypeGo("VkBool32")
	if !ok {
		t.Fatal("VkBool32: expected ok")
	}
	if got != "VkBool32" {
		t.Fatalf("VkBool32: got %q, want VkBool32", got)
	}
}
