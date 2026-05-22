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
	fmt.Printf("Vulkan spec IR: %d basetypes, %d API constants, %d enum types, %d flag types, %d funcpointers, %d handles, %d structs/unions, %d api versions, %d extensions, %d features\n",
		len(ir.Basetypes), len(ir.Constants.Values), len(ir.Enums), len(ir.Flags), len(ir.Funcpointers), len(ir.Handles), len(ir.Structs),
		len(ir.ApiVersions), len(ir.Extensions), len(ir.Features))
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
	for _, handle := range ir.Handles {
		kind := "non-dispatchable"
		if handle.Dispatchable {
			kind = "dispatchable"
		}
		if handle.AliasOf != "" {
			fmt.Printf("  handle %s = %s (%s)\n", handle.Name, handle.AliasOf, kind)
			continue
		}
		fmt.Printf("  handle %s %s (%s)\n", handle.Name, handle.Underlying, kind)
	}
}
