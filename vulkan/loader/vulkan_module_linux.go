//go:build linux

package loader

import "syscore"

const vulkanModuleLibraryName = "libvulkan.so.1"

func vulkanModuleLibraryLoad() (syscore.SYSCORE_Pure_DynamicLibrary, error) {
	return syscore.SYSCORE_Pure_LibraryLoad(vulkanModuleLibraryName)
}
