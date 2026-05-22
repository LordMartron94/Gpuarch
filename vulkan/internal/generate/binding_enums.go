package main

import "fmt"

/*
vulkanBindingSpecIRPrint writes a short summary of the spec IR to stdout for binding generation testing.

[Parameters]
ir is the registry intermediate representation.

[Side Effects]
Writes to stdout.
*/
func vulkanBindingSpecIRPrint(ir VulkanSpecIR) {
	fmt.Printf("Vulkan spec IR: %d enum types\n", len(ir.Enums))
	for _, enumType := range ir.Enums {
		fmt.Printf("  %s (%d values", enumType.Name, len(enumType.Values))
		if enumType.Doc != "" {
			fmt.Printf(", documented")
		}
		fmt.Println(")")
	}
}
