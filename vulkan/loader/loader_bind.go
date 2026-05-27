package loader

import (
	"fmt"

	"gpuarch/vulkan/bindings"
)

/*
vulkanLoaderEntryPointResolveError reports why a Vulkan entry point could not be loaded.

Call when vkGetInstanceProcAddr or vkGetDeviceProcAddr returns NULL, or before resolve when the
required VkInstance or VkDevice handle is still zero for that loader tier.
*/
func vulkanLoaderEntryPointResolveError(
	vkName string,
	tier int,
	instance bindings.VkInstance,
	device bindings.VkDevice,
) error {
	switch tier {
	case vulkanCommandLoaderTierGlobal:
		return fmt.Errorf(
			"vulkan loader: missing entry point %q (global tier via vkGetInstanceProcAddr(NULL, ...); verify the ICD is installed and exports the symbol",
			vkName,
		)
	case vulkanCommandLoaderTierInstance:
		if instance == 0 {
			return fmt.Errorf(
				"vulkan loader: cannot load %q: instance-tier commands require a valid VkInstance — pass the handle from vkCreateInstance to VulkanCommandsLoadManifest or VulkanCommandsLoadInstance (do not load instance-tier commands in init before the instance exists)",
				vkName,
			)
		}
		return fmt.Errorf(
			"vulkan loader: missing entry point %q (instance tier via vkGetInstanceProcAddr(instance, ...); required extension may be disabled, or the ICD does not export this symbol for the given instance",
			vkName,
		)
	case vulkanCommandLoaderTierDevice:
		if device == 0 {
			return fmt.Errorf(
				"vulkan loader: cannot load %q: device-tier commands require a valid VkDevice — pass the handle from vkCreateDevice to VulkanCommandsLoadManifest or VulkanCommandsLoadDevice (do not load device-tier commands before the device exists)",
				vkName,
			)
		}
		return fmt.Errorf(
			"vulkan loader: missing entry point %q (device tier via vkGetDeviceProcAddr(device, ...); required extension may be disabled, or the ICD does not export this symbol for the given device",
			vkName,
		)
	default:
		return fmt.Errorf("vulkan loader: missing entry point %q (unknown loader tier %d)", vkName, tier)
	}
}
