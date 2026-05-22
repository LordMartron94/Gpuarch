package dto

import (
	"gpuarch/vulkan/bindings"
	"unsafe"
)

/*
vulkanChainLink is the common prefix of Vulkan extensible structures (sType, pNext).

[Context]
Generated bindings place SType and PNext first on every chain-capable struct. Helpers cast through
this layout; callers must not reorder those fields.
*/
type vulkanChainLink struct {
	SType bindings.VkStructureType
	PNext unsafe.Pointer
}

/*
VulkanStructPNextPrepend inserts extension as the first node of head's pNext chain.

[Context]
Matches the usual Vulkan pattern: extension.PNext receives the previous head.PNext, then head.PNext
points at extension. extension.SType is set to extensionSType; head.SType is not modified.

[Invariants]
head and extension must remain valid until the consuming Vulkan call returns.
*/
func VulkanStructPNextPrepend(head unsafe.Pointer, extension unsafe.Pointer, extensionSType bindings.VkStructureType) {
	if head == nil || extension == nil {
		return
	}
	headLink := (*vulkanChainLink)(head)
	extLink := (*vulkanChainLink)(extension)
	extLink.SType = extensionSType
	extLink.PNext = headLink.PNext
	headLink.PNext = extension
}

/*
VulkanStructPNextAppend links extension at the tail of head's pNext chain.

[Context]
Walks the existing chain and sets the last node's PNext to extension. extension.PNext is cleared.
extension.SType is set to extensionSType.
*/
func VulkanStructPNextAppend(head unsafe.Pointer, extension unsafe.Pointer, extensionSType bindings.VkStructureType) {
	if head == nil || extension == nil {
		return
	}
	headLink := (*vulkanChainLink)(head)
	extLink := (*vulkanChainLink)(extension)
	extLink.SType = extensionSType
	extLink.PNext = nil

	if headLink.PNext == nil {
		headLink.PNext = extension
		return
	}

	tail := headLink.PNext
	for {
		tailLink := (*vulkanChainLink)(tail)
		if tailLink.PNext == nil {
			tailLink.PNext = extension
			return
		}
		tail = tailLink.PNext
	}
}

/*
VulkanStructPNextSet replaces head's entire pNext chain with a single extension node.

[Context]
Sets head.PNext to extension and extension.PNext to nil. Use when the chain has at most one extension.
*/
func VulkanStructPNextSet(head unsafe.Pointer, extension unsafe.Pointer, extensionSType bindings.VkStructureType) {
	if head == nil {
		return
	}
	headLink := (*vulkanChainLink)(head)
	if extension == nil {
		headLink.PNext = nil
		return
	}
	extLink := (*vulkanChainLink)(extension)
	extLink.SType = extensionSType
	extLink.PNext = nil
	headLink.PNext = extension
}

/*
VulkanStructPNextClear sets head.PNext to nil.
*/
func VulkanStructPNextClear(head unsafe.Pointer) {
	if head == nil {
		return
	}
	(*vulkanChainLink)(head).PNext = nil
}

/*
VulkanStructPNextFind returns the first node in head or its pNext chain whose sType matches, or nil.
*/
func VulkanStructPNextFind(head unsafe.Pointer, sType bindings.VkStructureType) unsafe.Pointer {
	var found unsafe.Pointer
	VulkanStructPNextEach(head, func(nodeSType bindings.VkStructureType, node unsafe.Pointer) bool {
		if nodeSType == sType {
			found = node
			return false
		}
		return true
	})
	return found
}

/*
VulkanStructPNextFindInChain returns the first extension node (head.PNext onward) whose sType matches, or nil.

[Context]
Use when head is a root create-info struct and extensions live only in the pNext chain.
*/
func VulkanStructPNextFindInChain(head unsafe.Pointer, sType bindings.VkStructureType) unsafe.Pointer {
	if head == nil {
		return nil
	}
	return VulkanStructPNextFind((*vulkanChainLink)(head).PNext, sType)
}

/*
VulkanStructPNextEach walks head and its pNext chain.

[Context]
fn receives each node's sType and pointer. Return false from fn to stop iteration. The root head is
included (unlike VulkanStructPNextFind, which only inspects the head when head is the chain root).
*/
func VulkanStructPNextEach(head unsafe.Pointer, fn func(sType bindings.VkStructureType, node unsafe.Pointer) bool) {
	for node := head; node != nil; {
		link := (*vulkanChainLink)(node)
		if !fn(link.SType, node) {
			return
		}
		if link.PNext == nil {
			return
		}
		node = link.PNext
	}
}

/*
VulkanStructPNextPrependTyped prepends extension onto head's pNext chain.

[Context]
Typed wrapper around VulkanStructPNextPrepend for call sites that already hold *Head and *Ext.
*/
func VulkanStructPNextPrependTyped[Head any, Ext any](head *Head, extension *Ext, extensionSType bindings.VkStructureType) {
	VulkanStructPNextPrepend(unsafe.Pointer(head), unsafe.Pointer(extension), extensionSType)
}

/*
VulkanStructPNextAppendTyped appends extension onto head's pNext chain.
*/
func VulkanStructPNextAppendTyped[Head any, Ext any](head *Head, extension *Ext, extensionSType bindings.VkStructureType) {
	VulkanStructPNextAppend(unsafe.Pointer(head), unsafe.Pointer(extension), extensionSType)
}

/*
VulkanStructPNextSetTyped replaces head's pNext chain with extension.
*/
func VulkanStructPNextSetTyped[Head any, Ext any](head *Head, extension *Ext, extensionSType bindings.VkStructureType) {
	VulkanStructPNextSet(unsafe.Pointer(head), unsafe.Pointer(extension), extensionSType)
}

/*
VulkanStructPNextClearTyped clears head.PNext.
*/
func VulkanStructPNextClearTyped[Head any](head *Head) {
	VulkanStructPNextClear(unsafe.Pointer(head))
}
