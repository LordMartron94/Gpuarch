package main

import "sort"

/*
vulkanSpecIRBuild constructs the registry IR from a parsed XML document tree.

[Parameters]
root is the document root from xmlSpecTreeParse.

[Returns]
A populated VulkanSpecIR. Enum entries are sorted by name. Does not mutate root.
*/
func vulkanSpecIRBuild(root *xmlSpecNode) VulkanSpecIR {
	enums := vulkanSpecIREnumsCollect(root)
	sort.Slice(enums, func(i, j int) bool {
		return enums[i].Name < enums[j].Name
	})
	return VulkanSpecIR{Enums: enums}
}

func vulkanSpecIREnumsCollect(root *xmlSpecNode) []VulkanSpecIREnum {
	if root == nil {
		return nil
	}

	var enums []VulkanSpecIREnum
	vulkanSpecIREnumsCollectWalk(root, &enums)
	return enums
}

func vulkanSpecIREnumsCollectWalk(node *xmlSpecNode, enums *[]VulkanSpecIREnum) {
	if node.Name == "enums" && xmlSpecNodeAttrValue(node, "type") == "enum" {
		if enumType := vulkanSpecIREnumBuild(node); enumType.Name != "" {
			*enums = append(*enums, enumType)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIREnumsCollectWalk(child, enums)
	}
}

func vulkanSpecIREnumBuild(enumsNode *xmlSpecNode) VulkanSpecIREnum {
	enumType := VulkanSpecIREnum{
		Name: xmlSpecNodeAttrValue(enumsNode, "name"),
		Doc:  xmlSpecNodeCommentAttr(enumsNode),
	}

	var sectionDoc string
	for _, child := range enumsNode.Children {
		switch child.Name {
		case "comment":
			sectionDoc = xmlSpecNodeCommentText(child)

		case "enum":
			value := vulkanSpecIREnumValueBuild(child, sectionDoc)
			if value.Key == "" || value.Value == "" {
				continue
			}
			enumType.Values = append(enumType.Values, value)
		}
	}

	return enumType
}

func vulkanSpecIREnumValueBuild(enumNode *xmlSpecNode, sectionDoc string) VulkanSpecIREnumValue {
	return VulkanSpecIREnumValue{
		Key:   xmlSpecNodeAttrValue(enumNode, "name"),
		Value: xmlSpecNodeAttrValue(enumNode, "value"),
		Doc:   vulkanSpecIRDocJoin(sectionDoc, xmlSpecNodeCommentAttr(enumNode)),
	}
}
