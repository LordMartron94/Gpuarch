package loader

import (
	"fmt"

	"gpuarch/vulkan/bindings"
	"syscore"
)

/*
VulkanModule holds the loaded Vulkan ICD library and core proc addr resolvers.
*/
type VulkanModule struct {
	Library             syscore.SYSCORE_Pure_DynamicLibrary
	GetInstanceProcAddr func(instance bindings.VkInstance, name *byte) uintptr
	GetDeviceProcAddr   func(device bindings.VkDevice, name *byte) uintptr
}

/*
VulkanModuleLoad opens the platform Vulkan library and binds global entry points.

[Returns]
A module with vkGetInstanceProcAddr and vkGetDeviceProcAddr ready for command loading.
*/
func VulkanModuleLoad() (*VulkanModule, error) {
	library, err := vulkanModuleLibraryLoad()
	if err != nil {
		return nil, fmt.Errorf("vulkan loader: open ICD library: %w", err)
	}

	module := &VulkanModule{Library: library}

	if err := syscore.SYSCORE_Pure_LibraryFunctionBind(library, &module.GetInstanceProcAddr, "vkGetInstanceProcAddr"); err != nil {
		return nil, fmt.Errorf("vulkan loader: bind vkGetInstanceProcAddr: %w", err)
	}
	if err := syscore.SYSCORE_Pure_LibraryFunctionBind(library, &module.GetDeviceProcAddr, "vkGetDeviceProcAddr"); err != nil {
		return nil, fmt.Errorf("vulkan loader: bind vkGetDeviceProcAddr: %w", err)
	}

	return module, nil
}
