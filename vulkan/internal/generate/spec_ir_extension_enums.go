package main

import "strconv"

const (
	vulkanSpecIREnumExtensionBase      = 1000000000
	vulkanSpecIREnumExtensionBlockSize = 1000
)

/*
vulkanSpecIRExtensionEnumsApply merges <enum> entries from feature and extension require sections into
collected enum and flag types. Entries inside <enums> blocks are already collected elsewhere.
*/
func vulkanSpecIRExtensionEnumsApply(root *xmlSpecNode, enums *[]VulkanSpecIREnum, flags []VulkanSpecIRFlags) []VulkanSpecIRFlags {
	enumByName := vulkanSpecIREnumIndexByName(*enums)
	flagsByFlagBits, flagsByAggregate := vulkanSpecIRFlagsIndex(flags)

	if root != nil {
		vulkanSpecIRExtensionEnumsWalk(root, nil, "", enumByName, flagsByFlagBits, flagsByAggregate)
		vulkanSpecIRAliasesFinalizeWalk(root, nil, enumByName, flagsByFlagBits, flagsByAggregate)
	}

	return vulkanSpecIRFlagsIndexFlatten(flagsByAggregate)
}

func vulkanSpecIRAliasesFinalizeWalk(
	node *xmlSpecNode,
	parent *xmlSpecNode,
	enumByName map[string]*VulkanSpecIREnum,
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
) {
	if node.Name == "enum" && parent != nil && xmlSpecNodeAttrValue(node, "alias") != "" {
		extends := xmlSpecNodeAttrValue(node, "extends")
		if extends == "" && parent.Name == "enums" {
			extends = xmlSpecNodeAttrValue(parent, "name")
		}
		vulkanSpecIRExtensionEnumApplyAlias(
			xmlSpecNodeAttrValue(node, "name"),
			xmlSpecNodeAttrValue(node, "alias"),
			xmlSpecNodeCommentAttr(node),
			extends,
			enumByName,
			flagsByFlagBits,
			flagsByAggregate,
		)
	}

	for _, child := range node.Children {
		vulkanSpecIRAliasesFinalizeWalk(child, node, enumByName, flagsByFlagBits, flagsByAggregate)
	}
}

func vulkanSpecIREnumIndexByName(enums []VulkanSpecIREnum) map[string]*VulkanSpecIREnum {
	index := make(map[string]*VulkanSpecIREnum, len(enums))
	for i := range enums {
		index[enums[i].Name] = &enums[i]
	}
	return index
}

func vulkanSpecIRFlagsIndex(flags []VulkanSpecIRFlags) (map[string]*VulkanSpecIRFlags, map[string]*VulkanSpecIRFlags) {
	byFlagBits := make(map[string]*VulkanSpecIRFlags, len(flags))
	byAggregate := make(map[string]*VulkanSpecIRFlags, len(flags))
	for i := range flags {
		flagType := &flags[i]
		if flagType.Name != "" {
			byFlagBits[flagType.Name] = flagType
		}
		if flagType.AggregateName != "" {
			byAggregate[flagType.AggregateName] = flagType
		}
	}
	return byFlagBits, byAggregate
}

func vulkanSpecIRFlagsIndexFlatten(byAggregate map[string]*VulkanSpecIRFlags) []VulkanSpecIRFlags {
	flags := make([]VulkanSpecIRFlags, 0, len(byAggregate))
	for _, flagType := range byAggregate {
		flags = append(flags, *flagType)
	}
	return flags
}

func vulkanSpecIRExtensionEnumsWalk(
	node *xmlSpecNode,
	parent *xmlSpecNode,
	extensionNumber string,
	enumByName map[string]*VulkanSpecIREnum,
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
) {
	if node.Name == "extension" {
		if number := xmlSpecNodeAttrValue(node, "number"); number != "" {
			extensionNumber = number
		}
	}

	if node.Name == "enum" && parent != nil && parent.Name != "enums" {
		vulkanSpecIRExtensionEnumApply(node, extensionNumber, enumByName, flagsByFlagBits, flagsByAggregate)
	}

	for _, child := range node.Children {
		vulkanSpecIRExtensionEnumsWalk(child, node, extensionNumber, enumByName, flagsByFlagBits, flagsByAggregate)
	}
}

func vulkanSpecIRExtensionEnumApply(
	enumNode *xmlSpecNode,
	extensionNumber string,
	enumByName map[string]*VulkanSpecIREnum,
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
) {
	key := xmlSpecNodeAttrValue(enumNode, "name")
	extends := xmlSpecNodeAttrValue(enumNode, "extends")
	if key == "" {
		return
	}

	doc := xmlSpecNodeCommentAttr(enumNode)

	if aliasTarget := xmlSpecNodeAttrValue(enumNode, "alias"); aliasTarget != "" {
		if extends == "" {
			return
		}
		vulkanSpecIRExtensionEnumApplyAlias(key, aliasTarget, doc, extends, enumByName, flagsByFlagBits, flagsByAggregate)
		return
	}

	if extends == "" {
		return
	}

	if bitpos := xmlSpecNodeAttrValue(enumNode, "bitpos"); bitpos != "" {
		value := vulkanSpecIRFlagsValueExpr(enumNode)
		if value == "" {
			return
		}
		vulkanSpecIRFlagsAppendValue(flagsByFlagBits, flagsByAggregate, extends, VulkanSpecIRFlagsValue{
			Key:   key,
			Value: value,
			Doc:   doc,
		})
		return
	}

	value := xmlSpecNodeAttrValue(enumNode, "value")
	if value == "" {
		value = vulkanSpecIREnumExtensionValue(enumNode, extensionNumber)
	}
	if value == "" {
		return
	}

	if vulkanSpecIRFlagsLookup(flagsByFlagBits, flagsByAggregate, extends) != nil {
		vulkanSpecIRFlagsAppendValue(flagsByFlagBits, flagsByAggregate, extends, VulkanSpecIRFlagsValue{
			Key:   key,
			Value: value,
			Doc:   doc,
		})
		return
	}

	vulkanSpecIREnumAppendValue(enumByName, extends, VulkanSpecIREnumValue{
		Key:   key,
		Value: value,
		Doc:   doc,
	})
}

func vulkanSpecIRExtensionEnumApplyAlias(
	key string,
	aliasTarget string,
	doc string,
	extends string,
	enumByName map[string]*VulkanSpecIREnum,
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
) {
	if enumType, ok := enumByName[extends]; ok {
		if value := vulkanSpecIREnumFindValue(enumType.Values, aliasTarget); value != "" {
			vulkanSpecIREnumAppendValue(enumByName, extends, VulkanSpecIREnumValue{Key: key, Value: value, Doc: doc})
		}
		return
	}

	if flagType := vulkanSpecIRFlagsLookup(flagsByFlagBits, flagsByAggregate, extends); flagType != nil {
		if value := vulkanSpecIREnumFindValueFlags(flagType.Values, aliasTarget); value != "" {
			vulkanSpecIRFlagsAppendValue(flagsByFlagBits, flagsByAggregate, extends, VulkanSpecIRFlagsValue{
				Key: key, Value: value, Doc: doc,
			})
		}
	}
}

func vulkanSpecIREnumExtensionValue(enumNode *xmlSpecNode, extensionNumber string) string {
	extnumber := xmlSpecNodeAttrValue(enumNode, "extnumber")
	if extnumber == "" {
		extnumber = extensionNumber
	}
	offset := xmlSpecNodeAttrValue(enumNode, "offset")
	if extnumber == "" || offset == "" {
		return ""
	}

	ext, errExt := strconv.Atoi(extnumber)
	off, errOff := strconv.Atoi(offset)
	if errExt != nil || errOff != nil {
		return ""
	}

	value := vulkanSpecIREnumExtensionBase + (ext-1)*vulkanSpecIREnumExtensionBlockSize + off
	if xmlSpecNodeAttrValue(enumNode, "dir") == "-" {
		value = -value
	}

	return strconv.Itoa(value)
}

func vulkanSpecIREnumAppendValue(enumByName map[string]*VulkanSpecIREnum, enumName string, value VulkanSpecIREnumValue) {
	enumType, ok := enumByName[enumName]
	if !ok || enumType == nil {
		return
	}
	if vulkanSpecIREnumHasValue(enumType.Values, value.Key) {
		return
	}
	enumType.Values = append(enumType.Values, value)
}

func vulkanSpecIRFlagsAppendValue(
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
	flagBitsName string,
	value VulkanSpecIRFlagsValue,
) {
	flagType := vulkanSpecIRFlagsLookup(flagsByFlagBits, flagsByAggregate, flagBitsName)
	if flagType == nil {
		return
	}
	if vulkanSpecIREnumHasValueFlags(flagType.Values, value.Key) {
		return
	}
	flagType.Values = append(flagType.Values, value)
}

func vulkanSpecIRFlagsLookup(
	flagsByFlagBits map[string]*VulkanSpecIRFlags,
	flagsByAggregate map[string]*VulkanSpecIRFlags,
	name string,
) *VulkanSpecIRFlags {
	if flagType, ok := flagsByFlagBits[name]; ok {
		return flagType
	}
	aggregateName := vulkanSpecIRFlagsAggregateName(name)
	if flagType, ok := flagsByAggregate[aggregateName]; ok {
		return flagType
	}
	if flagType, ok := flagsByAggregate[name]; ok {
		return flagType
	}
	return nil
}

func vulkanSpecIREnumFindValue(values []VulkanSpecIREnumValue, key string) string {
	for _, value := range values {
		if value.Key == key {
			return value.Value
		}
	}
	return ""
}

func vulkanSpecIREnumFindValueFlags(values []VulkanSpecIRFlagsValue, key string) string {
	for _, value := range values {
		if value.Key == key {
			return value.Value
		}
	}
	return ""
}

func vulkanSpecIREnumHasValue(values []VulkanSpecIREnumValue, key string) bool {
	for _, value := range values {
		if value.Key == key {
			return true
		}
	}
	return false
}

func vulkanSpecIREnumHasValueFlags(values []VulkanSpecIRFlagsValue, key string) bool {
	for _, value := range values {
		if value.Key == key {
			return true
		}
	}
	return false
}
