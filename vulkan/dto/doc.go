/*
Package dto bridges gpuarch Vulkan bindings with memstruct.Array and pNext chains.

Vulkan list parameters use a uint32/uint64 count plus a pointer to the first element of a
contiguous run. Use VulkanArrayBindU32 and VulkanArrayBindU64 to write those fields from a
memstruct.Array before a driver call. After a call fills count and pointer, use VulkanArrayLoadU32
and VulkanArrayLoadU64 to copy the run into a memstruct.Array the application owns.

Structures that participate in extensibility chains expose SType and PNext as their first fields.
Use VulkanStructPNextPrepend, VulkanStructPNextAppend, and VulkanStructPNextClear to build or
reset chains without manual unsafe.Pointer wiring.

When extensions are staged in memstruct.Array, use VulkanStructPNextChainBindFromArray for a run of
identical extension structs, or VulkanStructPNextChainBindFromPointerArray when each slot holds an
unsafe.Pointer to a different struct. After a driver fills a chain, VulkanStructPNextChainLength and
VulkanStructPNextChainLoadIntoPointerArray capture node addresses for later lookup with
VulkanStructPNextFindInChain.
*/
package dto
