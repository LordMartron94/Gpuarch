package main

/*
vulkanSpecIRProtoParamsBuild parses a registry <proto> and <param> list into Go signature parts.

When paramInclude is non-nil, only matching <param> nodes are collected (for example to skip
vulkansc-only duplicates inside a shared Vulkan command definition).
*/
func vulkanSpecIRProtoParamsBuild(
	node *xmlSpecNode,
	registry VulkanSpecIRTypeRegistry,
	paramInclude func(*xmlSpecNode) bool,
) (name string, returnGoType string, returnsUnsafe bool, params []VulkanSpecIRFuncpointerParam, ok bool) {
	var protoNode *xmlSpecNode
	params = make([]VulkanSpecIRFuncpointerParam, 0, len(node.Children))

	for _, child := range node.Children {
		switch child.Name {
		case "proto":
			protoNode = child
		case "param":
			if paramInclude != nil && !paramInclude(child) {
				continue
			}
			if param, paramOK := vulkanSpecIRFuncpointerParamBuild(child, registry); paramOK {
				params = append(params, param)
			}
		}
	}

	if protoNode == nil {
		return "", "", false, nil, false
	}

	name = vulkanSpecIRProtoName(protoNode)
	if name == "" {
		return "", "", false, nil, false
	}

	returnGoType, returnsUnsafe, ok = vulkanSpecIRProtoReturnGoType(protoNode, registry)
	if !ok {
		return "", "", false, nil, false
	}

	return name, returnGoType, returnsUnsafe, params, true
}
