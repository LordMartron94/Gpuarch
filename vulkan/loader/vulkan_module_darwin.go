//go:build darwin

package loader

import "syscore"

const vulkanModuleLibraryName = "libvulkan.1.dylib"

func vulkanModuleLibraryLoad() (syscore.SYSCORE_Pure_DynamicLibrary, error) {
	return syscore.SYSCORE_Pure_LibraryLoad(vulkanModuleLibraryName)
}
