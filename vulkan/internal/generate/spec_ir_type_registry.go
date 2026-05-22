package main

import "strings"

/*
VulkanSpecIRTypeRegistry maps Vulkan registry type names to the Go type names used in bindings.
*/
type VulkanSpecIRTypeRegistry map[string]string

func vulkanSpecIRTypeRegistryCollect(root *xmlSpecNode) VulkanSpecIRTypeRegistry {
	registry := make(VulkanSpecIRTypeRegistry)
	if root == nil {
		return registry
	}
	vulkanSpecIRTypeRegistryCollectWalk(root, registry)
	return registry
}

func vulkanSpecIRTypeRegistryCollectWalk(node *xmlSpecNode, registry VulkanSpecIRTypeRegistry) {
	if node.Name == "type" {
		category := xmlSpecNodeAttrValue(node, "category")
		name := xmlSpecNodeAttrValue(node, "name")
		alias := xmlSpecNodeAttrValue(node, "alias")
		if name != "" && alias != "" {
			switch category {
			case "enum", "bitmask", "struct", "handle", "basetype":
				registry[name] = alias
			}
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRTypeRegistryCollectWalk(child, registry)
	}
}

func vulkanSpecIRTypeRegistryResolve(registry VulkanSpecIRTypeRegistry, typeName string) string {
	if typeName == "" {
		return ""
	}

	resolved := typeName
	for {
		alias, ok := registry[resolved]
		if !ok || alias == "" || alias == resolved {
			break
		}
		resolved = alias
	}

	if strings.Contains(resolved, "FlagBits") {
		return vulkanSpecIRFlagsAggregateName(resolved)
	}

	return resolved
}
