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
	fmt.Printf("Vulkan spec IR: %d basetypes, %d API constants, %d enum types, %d flag types\n",
		len(ir.Basetypes), len(ir.Constants.Values), len(ir.Enums), len(ir.Flags))
	for _, basetype := range ir.Basetypes {
		fmt.Printf("  basetype %s -> %s\n", basetype.Name, basetype.Underlying)
	}
	if len(ir.Constants.Values) > 0 {
		fmt.Printf("  constants %s (%d values)\n", ir.Constants.Name, len(ir.Constants.Values))
	}
	for _, enumType := range ir.Enums {
		fmt.Printf("  enum %s (%d values)\n", enumType.Name, len(enumType.Values))
	}
	for _, flagType := range ir.Flags {
		fmt.Printf("  flags %s -> %s (%d bits)\n", flagType.Name, flagType.AggregateName, len(flagType.Values))
	}
}
