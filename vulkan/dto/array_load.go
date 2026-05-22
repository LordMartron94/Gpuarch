package dto

import (
	"errors"
	"fmt"
	"memcore"
	"memstruct"
	"unsafe"
)

/*
VulkanArrayLoadU32 copies a Vulkan contiguous list into memstruct.Array.

[Context]
Use after a driver call wrote count and pointer (for example the second pass of vkEnumeratePhysicalDevices).
The array must already have capacity at least equal to count.

[Returns]
When count is zero, clears the array. When count is non-zero and first is nil, returns an error.
*/
func VulkanArrayLoadU32[T any](count uint32, first *T, array memcore.MarkRaw) error {
	if count == 0 {
		memstruct.ArrayClear[T](array)
		return nil
	}
	if first == nil {
		return errors.New("vulkan dto: first element pointer must be non-nil when count is non-zero")
	}

	capacity := memstruct.ArrayCapacityGet[T](array)
	if uint64(count) > capacity {
		return fmt.Errorf("vulkan dto: array capacity %d is less than element count %d", capacity, count)
	}

	return memstruct.ArraySetFromSlice(array, unsafe.Slice(first, count))
}

/*
VulkanArrayLoadU64 copies a Vulkan contiguous list into memstruct.Array.

[Context]
Same as VulkanArrayLoadU32 for 64-bit Vulkan counts.
*/
func VulkanArrayLoadU64[T any](count uint64, first *T, array memcore.MarkRaw) error {
	if count == 0 {
		memstruct.ArrayClear[T](array)
		return nil
	}
	if first == nil {
		return errors.New("vulkan dto: first element pointer must be non-nil when count is non-zero")
	}

	capacity := memstruct.ArrayCapacityGet[T](array)
	if count > capacity {
		return fmt.Errorf("vulkan dto: array capacity %d is less than element count %d", capacity, count)
	}

	return memstruct.ArraySetFromSlice(array, unsafe.Slice(first, int(count)))
}
