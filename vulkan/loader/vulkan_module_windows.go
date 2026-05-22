//go:build windows

package loader

import "syscore"

const vulkanModuleLibraryName = "vulkan-1.dll"

func vulkanModuleLibraryLoad() (syscore.SYSCORE_Pure_DynamicLibrary, error) {
	return syscore.SYSCORE_Pure_LibraryLoad(vulkanModuleLibraryName)
}
