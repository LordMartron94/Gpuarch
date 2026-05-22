package main

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
		Name:   xmlSpecNodeAttrValue(enumsNode, "name"),
		XMLDoc: xmlSpecNodeDirectCommentsCollect(enumsNode),
	}

	var rawValues []vulkanSpecIRRawValue
	var sectionDoc string
	for _, child := range enumsNode.Children {
		switch child.Name {
		case "comment":
			sectionDoc = xmlSpecNodeCommentText(child)

		case "enum":
			raw := vulkanSpecIREnumRawValueBuild(child, sectionDoc)
			if raw.Key == "" {
				continue
			}
			rawValues = append(rawValues, raw)
		}
	}

	enumType.Values = vulkanSpecIRRawValuesResolveEnumValues(rawValues)
	return enumType
}

func vulkanSpecIREnumRawValueBuild(enumNode *xmlSpecNode, sectionDoc string) vulkanSpecIRRawValue {
	return vulkanSpecIRRawValue{
		Key:         xmlSpecNodeAttrValue(enumNode, "name"),
		Value:       xmlSpecNodeAttrValue(enumNode, "value"),
		AliasTarget: xmlSpecNodeAttrValue(enumNode, "alias"),
		Doc:         vulkanSpecIRDocJoin(sectionDoc, xmlSpecNodeCommentAttr(enumNode)),
	}
}
