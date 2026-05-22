package main

import (
	"fmt"
	"sort"
)

/*
xmlSpecNodeAttrValue returns the value of attrName on node, or an empty string when absent.
*/
func xmlSpecNodeAttrValue(node *xmlSpecNode, attrName string) string {
	if node == nil {
		return ""
	}
	for _, attr := range node.Attrs {
		if attr.Name == attrName {
			return attr.Value
		}
	}
	return ""
}

/*
vulkanRegistrySpecEnumTypeNamesCollect returns Vulkan enum type names from <enums type="enum" name="..."> blocks.

[Parameters]
root is the parsed registry document from xmlSpecTreeParse.

[Returns]
Enum type names (for example VkImageLayout). The slice is unsorted. Does not mutate root.
*/
func vulkanRegistrySpecEnumTypeNamesCollect(root *xmlSpecNode) []string {
	if root == nil {
		return nil
	}

	var names []string
	vulkanRegistrySpecEnumTypeNamesCollectWalk(root, &names)
	return names
}

func vulkanRegistrySpecEnumTypeNamesCollectWalk(node *xmlSpecNode, names *[]string) {
	if node.Name == "enums" && xmlSpecNodeAttrValue(node, "type") == "enum" {
		if name := xmlSpecNodeAttrValue(node, "name"); name != "" {
			*names = append(*names, name)
		}
	}

	for _, child := range node.Children {
		vulkanRegistrySpecEnumTypeNamesCollectWalk(child, names)
	}
}

/*
vulkanBindingEnumTypeNamesPrint writes all Vulkan enum type names from the registry tree to stdout.

[Parameters]
root is the parsed registry document.

[Side Effects]
Writes to stdout.
*/
func vulkanBindingEnumTypeNamesPrint(root *xmlSpecNode) {
	names := vulkanRegistrySpecEnumTypeNamesCollect(root)
	sort.Strings(names)

	fmt.Printf("Vulkan enum types (%d):\n", len(names))
	for _, name := range names {
		fmt.Println(name)
	}
}
