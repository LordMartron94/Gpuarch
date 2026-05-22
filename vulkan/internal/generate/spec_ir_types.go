package main

/*
VulkanSpecIR is the intermediate representation of a parsed Vulkan registry document.

Additional registry constructs will be added here as binding generation grows.
*/
type VulkanSpecIR struct {
	Basetypes []VulkanSpecIRBasetype
	Enums     []VulkanSpecIREnum
	Flags     []VulkanSpecIRFlags
}

/*
VulkanSpecIRBasetype is one Vulkan typedef from <type category="basetype"> with a Vk name.
*/
type VulkanSpecIRBasetype struct {
	Name       string
	Underlying string
	Doc        string
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

/*
VulkanSpecIRFlags is one Vulkan bitmask from <enums type="bitmask" name="...">.

Name is the registry FlagBits type (for example VkQueueFlagBits). AggregateName is the companion
Flags type used in Go bindings (for example VkQueueFlags).
*/
type VulkanSpecIRFlags struct {
	Name          string
	AggregateName string
	Doc           string
	Bitwidth      int
	BaseTypeName  string
	Values        []VulkanSpecIRFlagsValue
}

/*
VulkanSpecIRFlagsValue is one bit flag from <enum bitpos="..." name="..."> or <enum value="..." name="...">.

Key is the Vulkan token name. Value is the Go literal expression (for example 1 << 2 or 0x10).
*/
type VulkanSpecIRFlagsValue struct {
	Key   string
	Value string
	Doc   string
}
