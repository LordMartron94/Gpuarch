//go:build !linux && !windows && !darwin

package loader

import (
	"fmt"
	"runtime"

	"syscore"
)

func vulkanModuleLibraryLoad() (syscore.SYSCORE_Pure_DynamicLibrary, error) {
	return 0, fmt.Errorf("vulkan loader: unsupported GOOS %q", runtime.GOOS)
}
