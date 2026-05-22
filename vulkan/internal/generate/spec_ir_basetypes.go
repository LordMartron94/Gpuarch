package main

import "strings"

/*
vulkanSpecIRCTypeGo maps a Vulkan registry C type name (for example uint32_t) to a Go type name.
*/
func vulkanSpecIRCTypeGo(cType string) (goType string, ok bool) {
	goType, ok = vulkanSpecIRCTypeToGo[cType]
	return goType, ok
}

var vulkanSpecIRCTypeToGo = map[string]string{
	"uint8_t":  "uint8",
	"uint16_t": "uint16",
	"uint32_t": "uint32",
	"uint64_t": "uint64",
	"int8_t":   "int8",
	"int16_t":  "int16",
	"int32_t":  "int32",
	"int64_t":  "int64",
	"size_t":   "uintptr",
	"float":    "float32",
	"double":   "float64",
}

func vulkanSpecIRBasetypesCollect(root *xmlSpecNode) []VulkanSpecIRBasetype {
	if root == nil {
		return nil
	}

	var basetypes []VulkanSpecIRBasetype
	vulkanSpecIRBasetypesCollectWalk(root, &basetypes)
	return basetypes
}

func vulkanSpecIRBasetypesCollectWalk(node *xmlSpecNode, basetypes *[]VulkanSpecIRBasetype) {
	if node.Name == "type" && xmlSpecNodeAttrValue(node, "category") == "basetype" {
		if basetype := vulkanSpecIRBasetypeBuild(node); basetype.Name != "" {
			*basetypes = append(*basetypes, basetype)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRBasetypesCollectWalk(child, basetypes)
	}
}

func vulkanSpecIRBasetypeBuild(typeNode *xmlSpecNode) VulkanSpecIRBasetype {
	var name string
	var underlying string

	for _, child := range typeNode.Children {
		switch child.Name {
		case "name":
			name = strings.TrimSpace(child.Text)
			if name == "" {
				name = xmlSpecNodeAttrValue(child, "name")
			}
		case "type":
			underlying = strings.TrimSpace(child.Text)
			if underlying == "" {
				underlying = xmlSpecNodeAttrValue(child, "type")
			}
		}
	}

	if name == "" || underlying == "" {
		return VulkanSpecIRBasetype{}
	}
	if !strings.HasPrefix(name, "Vk") {
		return VulkanSpecIRBasetype{}
	}

	goType, ok := vulkanSpecIRCTypeToGo[underlying]
	if !ok {
		return VulkanSpecIRBasetype{}
	}

	return VulkanSpecIRBasetype{
		Name:       name,
		Underlying: goType,
		XMLDoc:     xmlSpecNodeDirectCommentsCollect(typeNode),
	}
}
