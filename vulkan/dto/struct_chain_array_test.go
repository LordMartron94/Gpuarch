package dto

import (
	"gpuarch/vulkan/bindings"
	"testing"
	"unsafe"
)

type chainTestNode struct {
	SType bindings.VkStructureType
	PNext unsafe.Pointer
	Data  uint32
}

func TestVulkanStructPNextChainLength(t *testing.T) {
	var root, extA, extB chainTestNode
	root.SType = 1
	extA.SType = 2
	extB.SType = 3
	root.PNext = unsafe.Pointer(&extA)
	extA.PNext = unsafe.Pointer(&extB)
	extB.PNext = nil

	if got := VulkanStructPNextChainLength(unsafe.Pointer(&root), false); got != 3 {
		t.Fatalf("full chain length = %d, want 3", got)
	}
	if got := VulkanStructPNextChainLength(unsafe.Pointer(&root), true); got != 2 {
		t.Fatalf("extensions-only length = %d, want 2", got)
	}
}

func TestVulkanStructPNextChainHeadPointsAtFirstExtension(t *testing.T) {
	var head, extA, extB chainTestNode
	extA.SType = 10
	extB.SType = 11
	head.PNext = unsafe.Pointer(&extA)
	extA.PNext = unsafe.Pointer(&extB)
	extB.PNext = nil

	first := (*chainTestNode)(head.PNext)
	if first != &extA || first.SType != 10 {
		t.Fatal("head.PNext should reference the first extension node")
	}
}
