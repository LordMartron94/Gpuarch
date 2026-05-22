package main

import (
	"strings"

	gocode "codegen/go"
)

func vulkanSpecIRFuncpointersCollect(root *xmlSpecNode, registry VulkanSpecIRTypeRegistry) []VulkanSpecIRFuncpointer {
	if root == nil {
		return nil
	}

	var funcpointers []VulkanSpecIRFuncpointer
	vulkanSpecIRFuncpointersCollectWalk(root, registry, &funcpointers)
	return funcpointers
}

func vulkanSpecIRFuncpointersCollectWalk(node *xmlSpecNode, registry VulkanSpecIRTypeRegistry, funcpointers *[]VulkanSpecIRFuncpointer) {
	if node.Name == "type" && xmlSpecNodeAttrValue(node, "category") == "funcpointer" {
		if fn := vulkanSpecIRFuncpointerBuild(node, registry); fn.Name != "" {
			*funcpointers = append(*funcpointers, fn)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRFuncpointersCollectWalk(child, registry, funcpointers)
	}
}

func vulkanSpecIRFuncpointerBuild(typeNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) VulkanSpecIRFuncpointer {
	var protoNode *xmlSpecNode
	params := make([]VulkanSpecIRFuncpointerParam, 0, len(typeNode.Children))

	for _, child := range typeNode.Children {
		switch child.Name {
		case "proto":
			protoNode = child
		case "param":
			if param, ok := vulkanSpecIRFuncpointerParamBuild(child, registry); ok {
				params = append(params, param)
			}
		}
	}

	if protoNode == nil {
		return VulkanSpecIRFuncpointer{}
	}

	name := vulkanSpecIRProtoName(protoNode)
	if name == "" || !strings.HasPrefix(name, "PFN_") {
		return VulkanSpecIRFuncpointer{}
	}

	returnGoType, returnsUnsafe, ok := vulkanSpecIRProtoReturnGoType(protoNode, registry)
	if !ok {
		return VulkanSpecIRFuncpointer{}
	}

	needsUnsafe := returnsUnsafe
	for _, param := range params {
		if param.NeedsUnsafe {
			needsUnsafe = true
		}
	}
	_ = needsUnsafe

	return VulkanSpecIRFuncpointer{
		Name:          name,
		ReturnGoType:  returnGoType,
		ReturnsUnsafe: returnsUnsafe,
		Params:        params,
		XMLDoc:        xmlSpecNodeDirectCommentsCollect(typeNode),
	}
}

func vulkanSpecIRFuncpointerParamBuild(paramNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) (VulkanSpecIRFuncpointerParam, bool) {
	name := vulkanSpecIRParamName(paramNode)
	if name == "" {
		return VulkanSpecIRFuncpointerParam{}, false
	}

	goType, needsUnsafe, ok := vulkanSpecIRParamGoTypeString(paramNode, registry)
	if !ok {
		return VulkanSpecIRFuncpointerParam{}, false
	}

	return VulkanSpecIRFuncpointerParam{
		Name:        name,
		GoType:      goType,
		NeedsUnsafe: needsUnsafe,
	}, true
}

func vulkanSpecIRProtoName(protoNode *xmlSpecNode) string {
	for _, child := range protoNode.Children {
		if child.Name == "name" {
			if name := strings.TrimSpace(child.Text); name != "" {
				return name
			}
			return strings.TrimSpace(xmlSpecNodeAttrValue(child, "name"))
		}
	}
	return ""
}

func vulkanSpecIRProtoReturnGoType(protoNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) (goType string, needsUnsafe bool, ok bool) {
	baseType := ""
	for _, child := range protoNode.Children {
		if child.Name == "type" {
			baseType = strings.TrimSpace(child.Text)
			if baseType == "" {
				baseType = strings.TrimSpace(xmlSpecNodeAttrValue(child, "type"))
			}
			break
		}
	}
	if baseType == "" {
		return "", false, false
	}

	baseType = vulkanSpecIRTypeRegistryResolve(registry, baseType)
	pointerDepth := strings.Count(vulkanSpecIRProtoRawText(protoNode), "*")

	if baseType == "void" {
		if pointerDepth == 0 {
			return "", false, true
		}
		return "unsafe.Pointer", true, true
	}

	goBase, ok := vulkanSpecIRMemberCTypeGo(baseType)
	if !ok {
		return "", false, false
	}

	goType = goBase
	for i := 0; i < pointerDepth; i++ {
		goType = "*" + goType
	}

	return goType, strings.Contains(goType, "unsafe.Pointer"), true
}

func vulkanSpecIRProtoRawText(protoNode *xmlSpecNode) string {
	var parts []string
	if text := strings.TrimSpace(protoNode.Text); text != "" {
		parts = append(parts, text)
	}
	for _, child := range protoNode.Children {
		if child.Name == "comment" {
			continue
		}
		if text := strings.TrimSpace(child.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func vulkanSpecIRParamName(paramNode *xmlSpecNode) string {
	for _, child := range paramNode.Children {
		if child.Name == "name" {
			if name := strings.TrimSpace(child.Text); name != "" {
				return name
			}
			return strings.TrimSpace(xmlSpecNodeAttrValue(child, "name"))
		}
	}
	return ""
}

func vulkanSpecIRParamGoTypeString(paramNode *xmlSpecNode, registry VulkanSpecIRTypeRegistry) (goType string, needsUnsafe bool, ok bool) {
	baseType := ""
	for _, child := range paramNode.Children {
		if child.Name == "type" {
			baseType = strings.TrimSpace(child.Text)
			if baseType == "" {
				baseType = strings.TrimSpace(xmlSpecNodeAttrValue(child, "type"))
			}
			break
		}
	}
	if baseType == "" {
		return "", false, false
	}

	baseType = vulkanSpecIRTypeRegistryResolve(registry, baseType)
	pointerDepth := strings.Count(vulkanSpecIRParamRawText(paramNode), "*")

	goBase, ok := vulkanSpecIRMemberCTypeGo(baseType)
	if !ok {
		return "", false, false
	}

	goType = goBase
	for i := 0; i < pointerDepth; i++ {
		goType = "*" + goType
	}

	return goType, strings.Contains(goType, "unsafe.Pointer"), true
}

func vulkanSpecIRParamRawText(paramNode *xmlSpecNode) string {
	var parts []string
	if text := strings.TrimSpace(paramNode.Text); text != "" {
		parts = append(parts, text)
	}
	for _, child := range paramNode.Children {
		if child.Name == "comment" {
			continue
		}
		if text := strings.TrimSpace(child.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func vulkanSpecIRFuncpointerGoTypeExpr(fn VulkanSpecIRFuncpointer) (gocode.TypeExpr, error) {
	params := make([]gocode.ParamType, 0, len(fn.Params))
	for _, param := range fn.Params {
		typ, err := gocode.TypeExprFromGoTypeString(param.GoType)
		if err != nil {
			return gocode.TypeExpr{}, err
		}
		params = append(params, gocode.ParamType{Name: param.Name, Type: typ})
	}

	var returns []gocode.TypeExpr
	if fn.ReturnGoType != "" && fn.ReturnGoType != "void" {
		ret, err := gocode.TypeExprFromGoTypeString(fn.ReturnGoType)
		if err != nil {
			return gocode.TypeExpr{}, err
		}
		returns = append(returns, ret)
	}

	return gocode.TypeExprFunc(params, returns), nil
}

func vulkanSpecIRFuncpointerNeedsUnsafe(fn VulkanSpecIRFuncpointer) bool {
	if fn.ReturnsUnsafe {
		return true
	}
	for _, param := range fn.Params {
		if param.NeedsUnsafe {
			return true
		}
	}
	return false
}
