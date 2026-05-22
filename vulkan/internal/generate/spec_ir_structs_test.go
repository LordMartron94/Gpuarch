package main

import "testing"

func TestVulkanSpecIRMemberGoTypeString(t *testing.T) {
	member := &xmlSpecNode{
		Name: "member",
		Text: "const struct ",
		Children: []*xmlSpecNode{
			{Name: "type", Text: "VkInstance"},
			{Text: "*"},
			{Name: "name", Text: "instance"},
		},
	}
	goType, ok := vulkanSpecIRMemberGoTypeString(member, nil)
	if !ok || goType != "*VkInstance" {
		t.Fatalf("got %q ok=%v", goType, ok)
	}
}

func TestVulkanSpecIRMemberGoTypeStringArray(t *testing.T) {
	member := &xmlSpecNode{
		Name: "member",
		Children: []*xmlSpecNode{
			{Name: "type", Text: "float"},
			{Name: "name", Text: "float32"},
		},
		Text: "[4]",
	}
	goType, ok := vulkanSpecIRMemberGoTypeString(member, nil)
	if !ok || goType != "[4]float32" {
		t.Fatalf("got %q ok=%v", goType, ok)
	}
}

func TestVulkanSpecIRMemberArrayLengthIgnoresCommentBrackets(t *testing.T) {
	member := &xmlSpecNode{
		Name: "member",
		Children: []*xmlSpecNode{
			{Name: "type", Text: "VkDescriptorBufferInfo"},
			{Name: "name", Text: "pBufferInfo"},
			{Name: "comment", Text: "for {UNIFORM,STORAGE}_BUFFER[_DYNAMIC] types"},
		},
	}
	if length := vulkanSpecIRMemberArrayLength(member); length != "" {
		t.Fatalf("expected no array length from comment, got %q", length)
	}
}

func TestVulkanSpecIRMemberGoTypeStringCharArray(t *testing.T) {
	member := &xmlSpecNode{
		Name: "member",
		Children: []*xmlSpecNode{
			{Name: "type", Text: "char"},
			{Name: "name", Text: "deviceName"},
			{Name: "enum", Text: "VK_MAX_PHYSICAL_DEVICE_NAME_SIZE"},
		},
	}
	goType, ok := vulkanSpecIRMemberGoTypeString(member, nil)
	if !ok || goType != "[VK_MAX_PHYSICAL_DEVICE_NAME_SIZE]byte" {
		t.Fatalf("got %q ok=%v", goType, ok)
	}
}
