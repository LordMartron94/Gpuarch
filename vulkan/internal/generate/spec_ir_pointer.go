package main

/*
vulkanSpecIRPointerGoType maps a resolved Go base type and C pointer depth to a Go type string.

void* collapses to unsafe.Pointer (not *unsafe.Pointer). Additional indirection prefixes * as usual,
so void** is *unsafe.Pointer and matches cgo/purego out-parameter conventions for void**.
*/
func vulkanSpecIRPointerGoType(goBase string, pointerDepth int) string {
	goType := goBase
	if goBase == "unsafe.Pointer" && pointerDepth >= 1 {
		goType = "unsafe.Pointer"
		for i := 1; i < pointerDepth; i++ {
			goType = "*" + goType
		}
		return goType
	}
	for i := 0; i < pointerDepth; i++ {
		goType = "*" + goType
	}
	return goType
}
