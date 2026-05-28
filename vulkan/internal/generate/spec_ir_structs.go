package main

import (
	"regexp"
	"strings"
	"unicode"
)

var vulkanSpecIRMemberArraySizePattern = regexp.MustCompile(`\[([^\]]+)\]`)

func vulkanSpecIRStructsCollect(root *xmlSpecNode, registry VulkanSpecIRTypeRegistry) []VulkanSpecIRStruct {
	if root == nil {
		return nil
	}

	var structs []VulkanSpecIRStruct
	vulkanSpecIRStructsCollectWalk(root, registry, &structs)
	vulkanSpecIRStructsResolveAliases(structs)
	return structs
}

func vulkanSpecIRStructsCollectWalk(node *xmlSpecNode, registry VulkanSpecIRTypeRegistry, structs *[]VulkanSpecIRStruct) {
	if node.Name == "type" {
		category := xmlSpecNodeAttrValue(node, "category")
		switch category {
		case "struct", "union":
			if aggregate := vulkanSpecIRStructBuild(node, registry, category == "union"); aggregate.Name != "" {
				*structs = append(*structs, aggregate)
			}
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRStructsCollectWalk(child, registry, structs)
	}
}

func vulkanSpecIRStructBuild(typeNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry, isUnion bool) VulkanSpecIRStruct {
	name := xmlSpecNodeAttrValue(typeNode, "name")
	if name == "" || !strings.HasPrefix(name, "Vk") {
		return VulkanSpecIRStruct{}
	}

	aggregate := VulkanSpecIRStruct{
		Name:    name,
		IsUnion: isUnion,
		AliasOf: xmlSpecNodeAttrValue(typeNode, "alias"),
		XMLDoc:  xmlSpecNodeDirectCommentsCollect(typeNode),
	}

	seenFields := make(map[string]struct{})
	for _, child := range typeNode.Children {
		if child.Name != "member" {
			continue
		}
		field, ok := vulkanSpecIRStructFieldBuild(child, registry)
		if !ok {
			continue
		}
		if _, exists := seenFields[field.Name]; exists {
			continue
		}
		seenFields[field.Name] = struct{}{}
		aggregate.Fields = append(aggregate.Fields, field)
	}

	return aggregate
}

func vulkanSpecIRStructFieldBuild(memberNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) (VulkanSpecIRStructField, bool) {
	name := vulkanSpecIRMemberFieldName(memberNode)
	if name == "" {
		return VulkanSpecIRStructField{}, false
	}

	goType, ok := vulkanSpecIRMemberGoTypeString(memberNode, registry)
	if !ok || goType == "" {
		return VulkanSpecIRStructField{}, false
	}

	return VulkanSpecIRStructField{
		Name:             vulkanSpecIRMemberGoFieldName(name),
		VulkanMemberName: name,
		GoType:           goType,
		XMLDoc:           xmlSpecNodeDirectCommentsCollect(memberNode),
		NeedsUnsafe:      strings.Contains(goType, "unsafe.Pointer"),
	}, true
}

func vulkanSpecIRMemberFieldName(memberNode *xmlSpecNode) string {
	for _, child := range memberNode.Children {
		if child.Name == "name" {
			if name := strings.TrimSpace(child.Text); name != "" {
				return name
			}
		}
	}
	return ""
}

func vulkanSpecIRMemberGoFieldName(name string) string {
	if name == "" {
		return ""
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func vulkanSpecIRMemberGoTypeString(memberNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) (string, bool) {
	baseType := vulkanSpecIRMemberBaseType(memberNode)
	if baseType == "" {
		return "", false
	}

	baseType = vulkanSpecIRTypeRegistryResolve(registry, baseType)
	goBase, ok := vulkanSpecIRMemberCTypeGo(baseType)
	if !ok {
		return "", false
	}

	pointerDepth := vulkanSpecIRMemberPointerDepth(memberNode)
	goType := vulkanSpecIRPointerGoType(goBase, pointerDepth)

	if arrayLen := vulkanSpecIRMemberArrayLength(memberNode); arrayLen != "" {
		goType = "[" + arrayLen + "]" + goType
	}

	return goType, true
}

func vulkanSpecIRMemberBaseType(memberNode *xmlSpecNode) string {
	for _, child := range memberNode.Children {
		if child.Name == "type" {
			return strings.TrimSpace(child.Text)
		}
	}
	return ""
}

func vulkanSpecIRMemberPointerDepth(memberNode *xmlSpecNode) int {
	return strings.Count(vulkanSpecIRMemberRawText(memberNode), "*")
}

func vulkanSpecIRMemberArrayLength(memberNode *xmlSpecNode) string {
	for _, child := range memberNode.Children {
		if child.Name == "enum" {
			if length := strings.TrimSpace(child.Text); length != "" && vulkanSpecIRMemberArrayLengthLiteral(length) {
				return length
			}
		}
	}

	matches := vulkanSpecIRMemberArraySizePattern.FindStringSubmatch(vulkanSpecIRMemberDeclarationText(memberNode))
	if len(matches) >= 2 {
		length := strings.TrimSpace(matches[1])
		if vulkanSpecIRMemberArrayLengthLiteral(length) {
			return length
		}
	}
	return ""
}

func vulkanSpecIRMemberArrayLengthLiteral(length string) bool {
	if length == "" {
		return false
	}
	if unicode.IsDigit(rune(length[0])) {
		return true
	}
	return strings.HasPrefix(length, "VK_")
}

func vulkanSpecIRMemberDeclarationText(memberNode *xmlSpecNode) string {
	var parts []string
	if text := strings.TrimSpace(memberNode.Text); text != "" {
		parts = append(parts, text)
	}
	for _, child := range memberNode.Children {
		if child.Name == "comment" {
			continue
		}
		if text := strings.TrimSpace(child.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func vulkanSpecIRMemberRawText(memberNode *xmlSpecNode) string {
	return vulkanSpecIRMemberDeclarationText(memberNode)
}

func vulkanSpecIRMemberCTypeGo(cType string) (string, bool) {
	cType = strings.TrimSpace(cType)
	if cType == "" {
		return "", false
	}

	if goType, ok := vulkanSpecIRCTypeGo(cType); ok {
		return goType, true
	}

	switch cType {
	case "char":
		return "byte", true
	case "void":
		return "unsafe.Pointer", true
	case "VkDeviceSize":
		return "uint64", true
	case "VkDeviceAddress":
		return "uint64", true
	case "VkFlags":
		return "uint32", true
	case "VkFlags64":
		return "uint64", true
	case "VkSampleMask":
		return "uint32", true
	}

	if strings.HasPrefix(cType, "Vk") {
		return cType, true
	}

	if strings.HasPrefix(cType, "PFN_") {
		return cType, true
	}

	return "", false
}

func vulkanSpecIRStructsResolveAliases(structs []VulkanSpecIRStruct) {
	byName := make(map[string]VulkanSpecIRStruct, len(structs))
	for _, aggregate := range structs {
		if aggregate.AliasOf == "" {
			byName[aggregate.Name] = aggregate
		}
	}

	for i := range structs {
		if structs[i].AliasOf == "" {
			continue
		}
		target, ok := byName[structs[i].AliasOf]
		if !ok {
			continue
		}
		structs[i].IsUnion = target.IsUnion
		if structs[i].XMLDoc == "" {
			structs[i].XMLDoc = target.XMLDoc
		}
	}
}
