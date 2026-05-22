package result

import (
	"fmt"

	"gpuarch/vulkan/bindings"
)

/*
VulkanResultError carries a VkResult failure or non-success status converted for Go error handling.

[Context]
Returned by VulkanResultToError and VulkanResultToErrorIfFailure when the corresponding predicate fails.
*/
type VulkanResultError struct {
	Code bindings.VkResult
}

func (e *VulkanResultError) Error() string {
	if e == nil {
		return "vulkan: nil result error"
	}
	return fmt.Sprintf("vulkan: %s (%d)", VulkanResultName(e.Code), int32(e.Code))
}

func vulkanResultErrorNew(code bindings.VkResult) error {
	return &VulkanResultError{Code: code}
}
