package dto

import (
	"errors"
	"fmt"
	"gpuarch/vulkan/bindings"
	"memcore"
	"memstruct"
	"unsafe"
)

/*
VulkanStructPNextChainBindFromArray wires head.PNext to the first element of extensions and links
each array element to the next.

[Context]
Use when several extension structs of the same type live in one memstruct.Array (for example a pool
of VkDeviceQueueCreateInfo-sized blocks is not this case; use VulkanArrayBindU32 for counted lists).
Each element must remain valid until the Vulkan call completes. Set SType on every element before
calling; this helper only wires PNext links.

[Returns]
Clears head.PNext when the array has zero capacity.
*/
func VulkanStructPNextChainBindFromArray[Ext any](head unsafe.Pointer, extensions memcore.MarkRaw) error {
	if head == nil {
		return errors.New("vulkan dto: head must be non-nil")
	}

	capacity := memstruct.ArrayCapacityGet[Ext](extensions)
	headLink := (*vulkanChainLink)(head)
	if capacity == 0 {
		headLink.PNext = nil
		return nil
	}

	for index := uint64(0); index < capacity; index++ {
		node := unsafe.Pointer(memstruct.ArrayItemPtrGetAtUnsafe[Ext](extensions, index))
		link := (*vulkanChainLink)(node)
		if index+1 < capacity {
			next := unsafe.Pointer(memstruct.ArrayItemPtrGetAtUnsafe[Ext](extensions, index+1))
			link.PNext = next
		} else {
			link.PNext = nil
		}
	}

	headLink.PNext = unsafe.Pointer(memstruct.ArrayItemPtrGetAtUnsafe[Ext](extensions, 0))
	return nil
}

/*
VulkanStructPNextChainBindFromPointerArray wires head.PNext from memstruct.Array[unsafe.Pointer].

[Context]
Each array slot holds the address of one extension struct (often different types and sTypes).
Nodes are chained in index order; existing SType values on each node are not modified.
*/
func VulkanStructPNextChainBindFromPointerArray(head unsafe.Pointer, nodes memcore.MarkRaw) error {
	if head == nil {
		return errors.New("vulkan dto: head must be non-nil")
	}

	capacity := memstruct.ArrayCapacityGet[unsafe.Pointer](nodes)
	headLink := (*vulkanChainLink)(head)
	if capacity == 0 {
		headLink.PNext = nil
		return nil
	}

	for index := uint64(0); index < capacity; index++ {
		node := memstruct.ArrayItemGetAtUnsafe[unsafe.Pointer](nodes, index)
		if node == nil {
			return fmt.Errorf("vulkan dto: nil extension pointer at array index %d", index)
		}
		link := (*vulkanChainLink)(node)
		if index+1 < capacity {
			next := memstruct.ArrayItemGetAtUnsafe[unsafe.Pointer](nodes, index+1)
			if next == nil {
				return fmt.Errorf("vulkan dto: nil extension pointer at array index %d", index+1)
			}
			link.PNext = next
		} else {
			link.PNext = nil
		}
	}

	headLink.PNext = memstruct.ArrayItemGetAtUnsafe[unsafe.Pointer](nodes, 0)
	return nil
}

/*
VulkanStructPNextChainLength returns how many nodes appear in a pNext walk.

[Context]
When extensionsOnly is true, only nodes reachable via head.PNext are counted (typical for root
create-info structs). When false, head itself is included in the count.
*/
func VulkanStructPNextChainLength(head unsafe.Pointer, extensionsOnly bool) uint64 {
	if head == nil {
		return 0
	}
	start := head
	if extensionsOnly {
		start = (*vulkanChainLink)(head).PNext
		if start == nil {
			return 0
		}
	}
	var length uint64
	VulkanStructPNextEach(start, func(_ bindings.VkStructureType, _ unsafe.Pointer) bool {
		length++
		return true
	})
	return length
}

/*
VulkanStructPNextChainLoadIntoPointerArray copies each node pointer from a pNext walk into nodes.

[Context]
Use after a driver call to capture the chain for later VulkanStructPNextFindInChain lookups.
Set extensionsOnly to true when head is a root struct and slots should hold only extensions.
Unused array capacity is filled with nil. Returns the number of pointers written.

[Returns]
An error when the chain has more nodes than array capacity.
*/
func VulkanStructPNextChainLoadIntoPointerArray(
	head unsafe.Pointer,
	nodes memcore.MarkRaw,
	extensionsOnly bool,
) (uint64, error) {
	capacity := memstruct.ArrayCapacityGet[unsafe.Pointer](nodes)
	if head == nil {
		memstruct.ArrayClear[unsafe.Pointer](nodes)
		return 0, nil
	}

	start := head
	if extensionsOnly {
		start = (*vulkanChainLink)(head).PNext
	}

	var index uint64
	var chainTooLong bool
	VulkanStructPNextEach(start, func(_ bindings.VkStructureType, node unsafe.Pointer) bool {
		if index >= capacity {
			chainTooLong = true
			return false
		}
		memstruct.ArraySetAtUnsafe(nodes, index, node)
		index++
		return true
	})
	if chainTooLong {
		return index, fmt.Errorf("vulkan dto: pNext chain has more than %d nodes", capacity)
	}

	for clearIndex := index; clearIndex < capacity; clearIndex++ {
		memstruct.ArraySetAtUnsafe[unsafe.Pointer](nodes, clearIndex, nil)
	}
	return index, nil
}

/*
VulkanStructPNextChainBindFromArrayTyped is the typed entry point for VulkanStructPNextChainBindFromArray.
*/
func VulkanStructPNextChainBindFromArrayTyped[Head any, Ext any](head *Head, extensions memcore.MarkRaw) error {
	return VulkanStructPNextChainBindFromArray[Ext](unsafe.Pointer(head), extensions)
}

/*
VulkanStructPNextChainBindFromPointerArrayTyped is the typed entry point for VulkanStructPNextChainBindFromPointerArray.
*/
func VulkanStructPNextChainBindFromPointerArrayTyped[Head any](head *Head, nodes memcore.MarkRaw) error {
	return VulkanStructPNextChainBindFromPointerArray(unsafe.Pointer(head), nodes)
}

/*
VulkanStructPNextChainLoadIntoPointerArrayTyped loads a chain into memstruct.Array[unsafe.Pointer].
*/
func VulkanStructPNextChainLoadIntoPointerArrayTyped[Head any](
	head *Head,
	nodes memcore.MarkRaw,
	extensionsOnly bool,
) (uint64, error) {
	return VulkanStructPNextChainLoadIntoPointerArray(unsafe.Pointer(head), nodes, extensionsOnly)
}
