package main

import "testing"

func TestVulkanSpecIRPointerGoTypeVoidStarStar(t *testing.T) {
	got := vulkanSpecIRPointerGoType("unsafe.Pointer", 2)
	if got != "*unsafe.Pointer" {
		t.Fatalf("void**: got %q, want *unsafe.Pointer", got)
	}
}

func TestVulkanSpecIRPointerGoTypeVoidStar(t *testing.T) {
	got := vulkanSpecIRPointerGoType("unsafe.Pointer", 1)
	if got != "unsafe.Pointer" {
		t.Fatalf("void*: got %q, want unsafe.Pointer", got)
	}
}

func TestVulkanSpecIRPointerGoTypeUint32Star(t *testing.T) {
	got := vulkanSpecIRPointerGoType("uint32", 1)
	if got != "*uint32" {
		t.Fatalf("uint32*: got %q, want *uint32", got)
	}
}
