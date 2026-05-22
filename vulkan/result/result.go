package result

import (
	"fmt"

	"gpuarch/vulkan/bindings"
)

/*
VulkanResultIsFailure reports whether result is a Vulkan failure code (negative value).

[Context]
Failure codes are all VkResult values less than zero per the Khronos registry.
*/
func VulkanResultIsFailure(result bindings.VkResult) bool {
	return int32(result) < 0
}

/*
VulkanResultIsStrictSuccess reports whether result is VK_SUCCESS.

[Context]
Many commands require exactly VK_SUCCESS; non-zero non-negative codes such as VK_INCOMPLETE are not success for those calls.
*/
func VulkanResultIsStrictSuccess(result bindings.VkResult) bool {
	return result == bindings.VK_SUCCESS
}

/*
VulkanResultIsNonFailure reports whether result is not a failure code (int32 value is non-negative).

[Context]
Includes VK_SUCCESS, status codes like VK_NOT_READY, and positive codes such as VK_SUBOPTIMAL_KHR.
*/
func VulkanResultIsNonFailure(result bindings.VkResult) bool {
	return int32(result) >= 0
}

/*
VulkanResultName returns the registry constant name for a known VkResult value.

[Context]
Unknown numeric values format as VkResult(<n>).
*/
func VulkanResultName(result bindings.VkResult) string {
	if name, ok := vulkanResultNameLookup(result); ok {
		return name
	}
	return fmt.Sprintf("VkResult(%d)", int32(result))
}

/*
VulkanResultToError returns nil when result is VK_SUCCESS.

[Context]
Any other value, including non-failure status codes such as VK_NOT_READY, becomes a VulkanResultError.
Use VulkanResultToErrorIfFailure when only negative failure codes should produce an error.
*/
func VulkanResultToError(result bindings.VkResult) error {
	if VulkanResultIsStrictSuccess(result) {
		return nil
	}
	return vulkanResultErrorNew(result)
}

/*
VulkanResultToErrorIfFailure returns nil when result is not a failure code.

[Context]
Non-negative status codes (VK_NOT_READY, VK_INCOMPLETE, VK_SUBOPTIMAL_KHR, and similar) return nil.
*/
func VulkanResultToErrorIfFailure(result bindings.VkResult) error {
	if !VulkanResultIsFailure(result) {
		return nil
	}
	return vulkanResultErrorNew(result)
}

/*
VulkanResultAssertSuccess returns nil when result is VK_SUCCESS.

[Context]
Use after commands that must complete with no status code other than success.
*/
func VulkanResultAssertSuccess(result bindings.VkResult) error {
	if VulkanResultIsStrictSuccess(result) {
		return nil
	}
	return vulkanResultErrorNew(result)
}

/*
VulkanResultAssertSuccessOr returns nil when result is VK_SUCCESS or matches one of allowed.

[Context]
Use when a command may legitimately return additional status codes (for example VK_INCOMPLETE from enumeration).
*/
func VulkanResultAssertSuccessOr(result bindings.VkResult, allowed ...bindings.VkResult) error {
	if VulkanResultIsStrictSuccess(result) {
		return nil
	}
	for _, code := range allowed {
		if result == code {
			return nil
		}
	}
	return vulkanResultErrorNew(result)
}

/*
VulkanResultAssertNoFailure returns nil when result is not a failure code.

[Context]
Accepts any non-negative VkResult. Use for flows that treat status codes as normal control flow.
*/
func VulkanResultAssertNoFailure(result bindings.VkResult) error {
	if VulkanResultIsNonFailure(result) {
		return nil
	}
	return vulkanResultErrorNew(result)
}

/*
VulkanResultAssertExpected returns nil when result equals expected.

[Context]
Use when a single specific VkResult is required (for example VK_EVENT_SET from vkGetEventStatus).
*/
func VulkanResultAssertExpected(result, expected bindings.VkResult) error {
	if result == expected {
		return nil
	}
	return vulkanResultErrorNew(result)
}
