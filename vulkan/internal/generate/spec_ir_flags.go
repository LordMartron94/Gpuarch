package main

import (
	"strconv"
	"strings"
)

const (
	vulkanSpecIRFlagsBase32 = "VkFlags"
	vulkanSpecIRFlagsBase64 = "VkFlags64"
)

func vulkanSpecIRFlagsCollect(root *xmlSpecNode) []VulkanSpecIRFlags {
	if root == nil {
		return nil
	}

	var flags []VulkanSpecIRFlags
	vulkanSpecIRFlagsCollectWalk(root, &flags)
	return flags
}

func vulkanSpecIRFlagsCollectWalk(node *xmlSpecNode, flags *[]VulkanSpecIRFlags) {
	if node.Name == "enums" && xmlSpecNodeAttrValue(node, "type") == "bitmask" {
		if flagType := vulkanSpecIRFlagsBuild(node); flagType.Name != "" {
			*flags = append(*flags, flagType)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRFlagsCollectWalk(child, flags)
	}
}

func vulkanSpecIRFlagsBuild(enumsNode *xmlSpecNode) VulkanSpecIRFlags {
	name := xmlSpecNodeAttrValue(enumsNode, "name")
	if name == "" {
		return VulkanSpecIRFlags{}
	}

	bitwidth := 32
	if xmlSpecNodeAttrValue(enumsNode, "bitwidth") == "64" {
		bitwidth = 64
	}

	flagType := VulkanSpecIRFlags{
		Name:          name,
		AggregateName: vulkanSpecIRFlagsAggregateName(name),
		Doc:           xmlSpecNodeCommentAttr(enumsNode),
		Bitwidth:      bitwidth,
		BaseTypeName:  vulkanSpecIRFlagsBaseTypeName(bitwidth),
	}

	var sectionDoc string
	for _, child := range enumsNode.Children {
		switch child.Name {
		case "comment":
			sectionDoc = xmlSpecNodeCommentText(child)

		case "enum":
			if xmlSpecNodeAttrValue(child, "alias") != "" {
				continue
			}
			value := vulkanSpecIRFlagsValueBuild(child, sectionDoc)
			if value.Key == "" || value.Value == "" {
				continue
			}
			flagType.Values = append(flagType.Values, value)
		}
	}

	return flagType
}

func vulkanSpecIRFlagsValueBuild(enumNode *xmlSpecNode, sectionDoc string) VulkanSpecIRFlagsValue {
	return VulkanSpecIRFlagsValue{
		Key:   xmlSpecNodeAttrValue(enumNode, "name"),
		Value: vulkanSpecIRFlagsValueExpr(enumNode),
		Doc:   vulkanSpecIRDocJoin(sectionDoc, xmlSpecNodeCommentAttr(enumNode)),
	}
}

func vulkanSpecIRFlagsValueExpr(enumNode *xmlSpecNode) string {
	if value := xmlSpecNodeAttrValue(enumNode, "value"); value != "" {
		return value
	}

	bitpos := xmlSpecNodeAttrValue(enumNode, "bitpos")
	if bitpos == "" {
		return ""
	}

	if _, err := strconv.Atoi(bitpos); err != nil {
		return ""
	}

	return "1 << " + bitpos
}

func vulkanSpecIRFlagsAggregateName(flagBitsName string) string {
	if strings.HasSuffix(flagBitsName, "FlagBits2") {
		return flagBitsName[:len(flagBitsName)-len("FlagBits2")] + "Flags2"
	}
	if strings.HasSuffix(flagBitsName, "FlagBits") {
		return flagBitsName[:len(flagBitsName)-len("FlagBits")] + "Flags"
	}
	return flagBitsName + "Flags"
}

func vulkanSpecIRFlagsBaseTypeName(bitwidth int) string {
	if bitwidth == 64 {
		return vulkanSpecIRFlagsBase64
	}
	return vulkanSpecIRFlagsBase32
}
