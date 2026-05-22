package dto

import (
	"errors"
	"fmt"
	"memcore"
	"memstruct"
)

/*
VulkanArrayBindU32 writes a Vulkan uint32 element count and pointer-to-first-element from memstruct.Array.

[Context]
Vulkan list parameters are (count, pointer to first). memstruct.Array stores elements contiguously; the
driver reads count elements starting at pointer. Capacity becomes the count; index zero supplies the pointer.

[Returns]
When capacity is zero, sets count to zero and first to nil. When capacity exceeds uint32 maximum, returns an error.
*/
func VulkanArrayBindU32[T any](count *uint32, first **T, array memcore.MarkRaw) error {
	if count == nil || first == nil {
		return errors.New("vulkan dto: count and first out-parameters must be non-nil")
	}

	capacity := memstruct.ArrayCapacityGet[T](array)
	if capacity == 0 {
		*count = 0
		*first = nil
		return nil
	}
	if capacity > uint64(^uint32(0)) {
		return fmt.Errorf("vulkan dto: array capacity %d exceeds uint32 maximum", capacity)
	}

	*count = uint32(capacity)
	*first = memstruct.ArrayItemPtrGetAtUnsafe[T](array, 0)
	return nil
}

/*
VulkanArrayBindU64 writes a Vulkan uint64 element count and pointer-to-first-element from memstruct.Array.

[Context]
Same as VulkanArrayBindU32 for 64-bit Vulkan counts (for example timeline semaphore wait values).
*/
func VulkanArrayBindU64[T any](count *uint64, first **T, array memcore.MarkRaw) error {
	if count == nil || first == nil {
		return errors.New("vulkan dto: count and first out-parameters must be non-nil")
	}

	capacity := memstruct.ArrayCapacityGet[T](array)
	if capacity == 0 {
		*count = 0
		*first = nil
		return nil
	}

	*count = capacity
	*first = memstruct.ArrayItemPtrGetAtUnsafe[T](array, 0)
	return nil
}
