package main

import (
	"strings"
)

/*
VulkanSpecIRBitmaskRegistry maps Vulkan FlagBits type names to their companion Flags typedef metadata from
<type category="bitmask"> entries in the registry.
*/
type VulkanSpecIRBitmaskRegistry map[string]VulkanSpecIRBitmaskTypeInfo

/*
VulkanSpecIRBitmaskTypeInfo is one typedef VkFlags VkFooFlags requires="VkFooFlagBits" from the types section.
*/
type VulkanSpecIRBitmaskTypeInfo struct {
	FlagBitsName  string
	AggregateName string
	BaseTypeName  string
	Doc           string
}

func vulkanSpecIRBitmaskRegistryCollect(root *xmlSpecNode) VulkanSpecIRBitmaskRegistry {
	registry := make(VulkanSpecIRBitmaskRegistry)
	if root == nil {
		return registry
	}
	vulkanSpecIRBitmaskRegistryCollectWalk(root, registry)
	return registry
}

func vulkanSpecIRBitmaskRegistryCollectWalk(node *xmlSpecNode, registry VulkanSpecIRBitmaskRegistry) {
	if node.Name == "type" && xmlSpecNodeAttrValue(node, "category") == "bitmask" {
		if info := vulkanSpecIRBitmaskTypeInfoParse(node); info.FlagBitsName != "" && info.AggregateName != "" {
			registry[info.FlagBitsName] = info
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRBitmaskRegistryCollectWalk(child, registry)
	}
}

func vulkanSpecIRBitmaskTypeInfoParse(typeNode *xmlSpecNode) VulkanSpecIRBitmaskTypeInfo {
	flagBitsName := xmlSpecNodeAttrValue(typeNode, "requires")
	if flagBitsName == "" {
		flagBitsName = xmlSpecNodeAttrValue(typeNode, "bitvalues")
	}

	var baseTypeName string
	var aggregateName string
	for _, child := range typeNode.Children {
		switch child.Name {
		case "type":
			if text := strings.TrimSpace(child.Text); text != "" {
				baseTypeName = text
			}
		case "name":
			if text := strings.TrimSpace(child.Text); text != "" {
				aggregateName = text
			}
		}
	}

	if flagBitsName == "" && aggregateName != "" {
		flagBitsName = vulkanSpecIRFlagsFlagBitsNameFromAggregate(aggregateName)
	}
	if baseTypeName == "" {
		baseTypeName = vulkanSpecIRFlagsBase32
	}

	return VulkanSpecIRBitmaskTypeInfo{
		FlagBitsName:  flagBitsName,
		AggregateName: aggregateName,
		BaseTypeName:  baseTypeName,
		Doc:           xmlSpecNodeCommentAttr(typeNode),
	}
}

func (registry VulkanSpecIRBitmaskRegistry) Lookup(flagBitsName string) (VulkanSpecIRBitmaskTypeInfo, bool) {
	info, ok := registry[flagBitsName]
	return info, ok
}
