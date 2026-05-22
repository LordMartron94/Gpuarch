package main

import (
	"strconv"
	"strings"
)

const (
	vulkanSpecIRFlagsBase32 = "VkFlags"
	vulkanSpecIRFlagsBase64 = "VkFlags64"
)

func vulkanSpecIRFlagsCollect(root *xmlSpecNode, registry VulkanSpecIRBitmaskRegistry) []VulkanSpecIRFlags {
	if root == nil {
		return nil
	}

	byAggregate := make(map[string]VulkanSpecIRFlags)
	vulkanSpecIRFlagsCollectWalk(root, registry, byAggregate)
	vulkanSpecIRFlagsMergeRegistry(registry, byAggregate)

	flags := make([]VulkanSpecIRFlags, 0, len(byAggregate))
	for _, flagType := range byAggregate {
		flags = append(flags, flagType)
	}
	return flags
}

func vulkanSpecIRFlagsCollectWalk(node *xmlSpecNode, registry VulkanSpecIRBitmaskRegistry, byAggregate map[string]VulkanSpecIRFlags) {
	if node.Name == "enums" && xmlSpecNodeAttrValue(node, "type") == "bitmask" {
		if flagType := vulkanSpecIRFlagsBuild(node, registry); flagType.AggregateName != "" {
			byAggregate[flagType.AggregateName] = flagType
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRFlagsCollectWalk(child, registry, byAggregate)
	}
}

func vulkanSpecIRFlagsMergeRegistry(registry VulkanSpecIRBitmaskRegistry, byAggregate map[string]VulkanSpecIRFlags) {
	for _, info := range registry {
		if info.AggregateName == "" {
			continue
		}
		if _, exists := byAggregate[info.AggregateName]; exists {
			continue
		}
		byAggregate[info.AggregateName] = VulkanSpecIRFlags{
			Name:          info.FlagBitsName,
			AggregateName: info.AggregateName,
			Doc:           info.Doc,
			Bitwidth:      vulkanSpecIRFlagsBitwidthFromBase(info.BaseTypeName),
			BaseTypeName:  info.BaseTypeName,
		}
	}
}

func vulkanSpecIRFlagsBuild(enumsNode *xmlSpecNode, registry VulkanSpecIRBitmaskRegistry) VulkanSpecIRFlags {
	flagBitsName := xmlSpecNodeAttrValue(enumsNode, "name")
	if flagBitsName == "" {
		return VulkanSpecIRFlags{}
	}

	flagType := VulkanSpecIRFlags{
		Name: flagBitsName,
		Doc:  xmlSpecNodeCommentAttr(enumsNode),
	}

	if info, ok := registry.Lookup(flagBitsName); ok {
		flagType.AggregateName = info.AggregateName
		flagType.BaseTypeName = info.BaseTypeName
		flagType.Bitwidth = vulkanSpecIRFlagsBitwidthFromBase(info.BaseTypeName)
		if flagType.Doc == "" {
			flagType.Doc = info.Doc
		}
	} else {
		flagType.AggregateName = vulkanSpecIRFlagsAggregateName(flagBitsName)
		flagType.Bitwidth = 32
		if xmlSpecNodeAttrValue(enumsNode, "bitwidth") == "64" {
			flagType.Bitwidth = 64
		}
		flagType.BaseTypeName = vulkanSpecIRFlagsBaseTypeName(flagType.Bitwidth)
	}

	var rawValues []vulkanSpecIRRawValue
	var sectionDoc string
	for _, child := range enumsNode.Children {
		switch child.Name {
		case "comment":
			sectionDoc = xmlSpecNodeCommentText(child)

		case "enum":
			raw := vulkanSpecIRFlagsRawValueBuild(child, sectionDoc)
			if raw.Key == "" {
				continue
			}
			rawValues = append(rawValues, raw)
		}
	}

	flagType.Values = vulkanSpecIRRawValuesResolveFlagsValues(rawValues)
	return flagType
}

func vulkanSpecIRFlagsRawValueBuild(enumNode *xmlSpecNode, sectionDoc string) vulkanSpecIRRawValue {
	aliasTarget := xmlSpecNodeAttrValue(enumNode, "alias")
	value := vulkanSpecIRFlagsValueExpr(enumNode)

	return vulkanSpecIRRawValue{
		Key:         xmlSpecNodeAttrValue(enumNode, "name"),
		Value:       value,
		AliasTarget: aliasTarget,
		Doc:         vulkanSpecIRDocJoin(sectionDoc, xmlSpecNodeCommentAttr(enumNode)),
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

/*
vulkanSpecIRFlagsAggregateName derives the Vk*Flags type name from a Vk*FlagBits enums block name.

Extension suffixes after FlagBits (for example VkCompositeAlphaFlagBitsKHR) are preserved.
*/
func vulkanSpecIRFlagsAggregateName(flagBitsName string) string {
	if strings.Contains(flagBitsName, "FlagBits2") {
		return strings.Replace(flagBitsName, "FlagBits2", "Flags2", 1)
	}
	if strings.Contains(flagBitsName, "FlagBits") {
		return strings.Replace(flagBitsName, "FlagBits", "Flags", 1)
	}
	return flagBitsName + "Flags"
}

func vulkanSpecIRFlagsFlagBitsNameFromAggregate(aggregateName string) string {
	if strings.HasSuffix(aggregateName, "Flags2") {
		return aggregateName[:len(aggregateName)-len("Flags2")] + "FlagBits2"
	}
	if strings.HasSuffix(aggregateName, "Flags") {
		return aggregateName[:len(aggregateName)-len("Flags")] + "FlagBits"
	}
	return aggregateName + "FlagBits"
}

func vulkanSpecIRFlagsBitwidthFromBase(baseTypeName string) int {
	if baseTypeName == vulkanSpecIRFlagsBase64 {
		return 64
	}
	return 32
}

func vulkanSpecIRFlagsBaseTypeName(bitwidth int) string {
	if bitwidth == 64 {
		return vulkanSpecIRFlagsBase64
	}
	return vulkanSpecIRFlagsBase32
}
