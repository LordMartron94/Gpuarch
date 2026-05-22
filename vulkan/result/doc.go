/*
Package result provides VkResult classification, assertion helpers, and conversion to Go errors.

Vulkan commands return VkResult. Negative values are failure codes; zero is VK_SUCCESS; other
non-negative values are non-failure status codes (for example VK_NOT_READY or VK_INCOMPLETE).
Use VulkanResultAssertSuccess when only VK_SUCCESS is acceptable, VulkanResultAssertSuccessOr
when additional status codes are valid, and VulkanResultAssertNoFailure when any non-negative
result is acceptable.

VulkanResultToError maps any result other than VK_SUCCESS to an error. VulkanResultToErrorIfFailure
maps only negative failure codes to an error and leaves other non-negative codes as nil.
*/
package result
