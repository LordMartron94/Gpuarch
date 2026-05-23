package version

/*
VulkanMakeAPIVersion packs a Vulkan API version number per VK_MAKE_API_VERSION.

[Context]
Use for VkApplicationInfo::apiVersion. Variant is usually 0 for Vulkan API versions.

[Parameters]
variant - API variant (0 for Vulkan).
major, minor, patch - Version components.

[Returns]
Packed uint32 suitable for bindings.VkApplicationInfo.ApiVersion.
*/
func VulkanMakeAPIVersion(variant, major, minor, patch uint32) uint32 {
	return (variant << 29) | (major << 22) | (minor << 12) | patch
}

/*
VulkanMakeVersion packs an application or engine version per the Vulkan version number scheme.

[Context]
Equivalent to VK_MAKE_API_VERSION(0, major, minor, patch). Use for VkApplicationInfo::applicationVersion
and VkApplicationInfo::engineVersion.
*/
func VulkanMakeVersion(major, minor, patch uint32) uint32 {
	return VulkanMakeAPIVersion(0, major, minor, patch)
}
