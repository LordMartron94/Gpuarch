package main

/*
VulkanSpecIR is the intermediate representation of a parsed Vulkan registry document.

Additional registry constructs will be added here as binding generation grows.
*/
type VulkanSpecIR struct {
	Enums []VulkanSpecIREnum
}

/*
VulkanSpecIREnum is one Vulkan enumerated type from <enums type="enum" name="...">.
*/
type VulkanSpecIREnum struct {
	Name   string
	Doc    string
	Values []VulkanSpecIREnumValue
}

/*
VulkanSpecIREnumValue is one enumerator from <enum name="..." value="...">.

Key is the Vulkan token name. Value is the numeric token value as written in the registry XML.
*/
type VulkanSpecIREnumValue struct {
	Key   string
	Value string
	Doc   string
}
