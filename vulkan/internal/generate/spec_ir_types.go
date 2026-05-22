package main

/*
VulkanSpecIR is the intermediate representation of a parsed Vulkan registry document.

Additional registry constructs will be added here as binding generation grows.
*/
type VulkanSpecIR struct {
	Basetypes []VulkanSpecIRBasetype
	Constants VulkanSpecIRConstants
	Enums     []VulkanSpecIREnum
	Flags     []VulkanSpecIRFlags
	Handles   []VulkanSpecIRHandle
}

/*
VulkanSpecIRConstants is the API Constants block from <enums type="constants">.
*/
type VulkanSpecIRConstants struct {
	Name   string
	Doc    string
	Values []VulkanSpecIRConstantValue
}

/*
VulkanSpecIRConstantValue is one hardcoded constant from the registry API Constants block.

Key is the Vulkan token name. Value is a Go constant literal expression. GoType is set when the
constant must be typed in Go output (for example float32); empty means an untyped const.
*/
type VulkanSpecIRConstantValue struct {
	Key    string
	Value  string
	GoType string
	Doc    string
}

/*
VulkanSpecIRHandle is one Vulkan handle from <type category="handle">.

Dispatchable handles map to uintptr; non-dispatchable handles map to uint64. AliasOf is set for
registry type aliases that share the same underlying handle representation.
*/
type VulkanSpecIRHandle struct {
	Name           string
	Underlying     string
	Dispatchable   bool
	Parent         string
	ObjectTypeEnum string
	AliasOf        string
	Doc            string
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
