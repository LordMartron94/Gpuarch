package main

import (
	"sort"
	"strconv"
	"strings"
)

var vulkanSpecIRConstantCSuffixToGo = []struct {
	suffix string
	goType string
}{
	{suffix: "ULL", goType: "uint64"},
	{suffix: "LLU", goType: "uint64"},
	{suffix: "UL", goType: "uint64"},
	{suffix: "LU", goType: "uint64"},
	{suffix: "LL", goType: "int64"},
	{suffix: "U", goType: "uint32"},
	{suffix: "L", goType: "int64"},
}

func vulkanSpecIRConstantsCollect(root *xmlSpecNode) VulkanSpecIRConstants {
	if root == nil {
		return VulkanSpecIRConstants{}
	}

	var result VulkanSpecIRConstants
	vulkanSpecIRConstantsCollectWalk(root, &result)
	return result
}

func vulkanSpecIRConstantsCollectWalk(node *xmlSpecNode, result *VulkanSpecIRConstants) {
	if node.Name == "enums" && xmlSpecNodeAttrValue(node, "type") == "constants" {
		*result = vulkanSpecIRConstantsBuild(node)
	}

	for _, child := range node.Children {
		vulkanSpecIRConstantsCollectWalk(child, result)
	}
}

func vulkanSpecIRConstantsBuild(enumsNode *xmlSpecNode) VulkanSpecIRConstants {
	block := VulkanSpecIRConstants{
		Name: xmlSpecNodeAttrValue(enumsNode, "name"),
		Doc:  xmlSpecNodeCommentAttr(enumsNode),
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
			value := vulkanSpecIRConstantValueBuild(child, sectionDoc)
			if value.Key == "" || value.Value == "" {
				continue
			}
			block.Values = append(block.Values, value)
		}
	}

	sort.Slice(block.Values, func(i, j int) bool { return block.Values[i].Key < block.Values[j].Key })
	return block
}

func vulkanSpecIRConstantValueBuild(enumNode *xmlSpecNode, sectionDoc string) VulkanSpecIRConstantValue {
	cType := xmlSpecNodeAttrValue(enumNode, "type")
	rawValue := xmlSpecNodeAttrValue(enumNode, "value")
	goValue, goType := vulkanSpecIRConstantGoLiteral(cType, rawValue)

	return VulkanSpecIRConstantValue{
		Key:    xmlSpecNodeAttrValue(enumNode, "name"),
		Value:  goValue,
		GoType: goType,
		Doc:    vulkanSpecIRDocJoin(sectionDoc, xmlSpecNodeCommentAttr(enumNode)),
	}
}

/*
vulkanSpecIRConstantGoLiteral converts a registry API constant value attribute to a Go const initializer.

Uses the registry type attribute when present; otherwise infers width from C integer suffixes on literals.
*/
func vulkanSpecIRConstantGoLiteral(cType string, rawValue string) (goLiteral string, goType string) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return "", ""
	}

	if goType, ok := vulkanSpecIRCTypeGo(cType); ok {
		if goType == "float32" || goType == "float64" {
			if literal, ok := vulkanSpecIRConstantParseFloatLiteral(rawValue); ok {
				return literal, goType
			}
		}
	}

	if operand, suffix, ok := vulkanSpecIRConstantParseBitwiseNot(rawValue); ok {
		widthType, _ := vulkanSpecIRCTypeGo(cType)
		if widthType == "" {
			widthType = vulkanSpecIRConstantGoTypeFromSuffix(suffix)
		}
		if widthType == "" {
			widthType = "uint32"
		}
		return "^" + widthType + "(" + operand + ")", ""
	}

	if literal, ok := vulkanSpecIRConstantParseFloatLiteral(rawValue); ok {
		goType = "float32"
		if mapped, ok := vulkanSpecIRCTypeGo(cType); ok {
			goType = mapped
		}
		return literal, goType
	}

	return rawValue, ""
}

/*
vulkanSpecIRConstantParseBitwiseNot recognizes C parenthesized bitwise NOT forms such as (~0U).
*/
func vulkanSpecIRConstantParseBitwiseNot(rawValue string) (operand string, suffix string, ok bool) {
	if len(rawValue) < 4 || rawValue[0] != '(' || rawValue[len(rawValue)-1] != ')' {
		return "", "", false
	}

	inner := strings.TrimSpace(rawValue[1 : len(rawValue)-1])
	if len(inner) < 2 || inner[0] != '~' {
		return "", "", false
	}

	operandPart := strings.TrimSpace(inner[1:])
	operand, suffix = vulkanSpecIRConstantSplitNumericSuffix(operandPart)
	if operand == "" || !vulkanSpecIRConstantLooksNumeric(operand) {
		return "", "", false
	}

	return operand, suffix, true
}

func vulkanSpecIRConstantSplitNumericSuffix(operandPart string) (operand string, suffix string) {
	operandPart = strings.TrimSpace(operandPart)
	if operandPart == "" {
		return "", ""
	}

	for _, entry := range vulkanSpecIRConstantCSuffixToGo {
		if !strings.HasSuffix(operandPart, entry.suffix) {
			continue
		}
		candidate := strings.TrimSpace(operandPart[:len(operandPart)-len(entry.suffix)])
		if candidate != "" && vulkanSpecIRConstantLooksNumeric(candidate) {
			return candidate, entry.suffix
		}
	}

	if vulkanSpecIRConstantLooksNumeric(operandPart) {
		return operandPart, ""
	}

	return "", ""
}

func vulkanSpecIRConstantGoTypeFromSuffix(suffix string) string {
	for _, entry := range vulkanSpecIRConstantCSuffixToGo {
		if entry.suffix == suffix {
			return entry.goType
		}
	}
	return ""
}

func vulkanSpecIRConstantParseFloatLiteral(rawValue string) (literal string, ok bool) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return "", false
	}

	if strings.HasSuffix(rawValue, "F") || strings.HasSuffix(rawValue, "f") {
		literal = strings.TrimSpace(rawValue[:len(rawValue)-1])
		if literal == "" {
			return "", false
		}
		return literal, true
	}

	return "", false
}

func vulkanSpecIRConstantLooksNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		_, err := strconv.ParseUint(s[2:], 16, 64)
		return err == nil
	}

	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}
