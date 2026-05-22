package main

import (
	"sort"
	"strings"
)

const (
	vulkanSpecIRHandleDispatchableUnderlying    = "uintptr"
	vulkanSpecIRHandleNonDispatchableUnderlying = "uint64"
	vulkanSpecIRHandleMacroDispatchable         = "VK_DEFINE_HANDLE"
	vulkanSpecIRHandleMacroNonDispatchable      = "VK_DEFINE_NON_DISPATCHABLE_HANDLE"
)

func vulkanSpecIRHandlesCollect(root *xmlSpecNode) []VulkanSpecIRHandle {
	if root == nil {
		return nil
	}

	var handles []VulkanSpecIRHandle
	vulkanSpecIRHandlesCollectWalk(root, &handles)
	vulkanSpecIRHandlesResolveAliases(handles)
	sort.Slice(handles, func(i, j int) bool {
		if handles[i].AliasOf != handles[j].AliasOf {
			return handles[i].AliasOf == ""
		}
		return handles[i].Name < handles[j].Name
	})
	return handles
}

func vulkanSpecIRHandlesCollectWalk(node *xmlSpecNode, handles *[]VulkanSpecIRHandle) {
	if node.Name == "type" && xmlSpecNodeAttrValue(node, "category") == "handle" {
		if handle := vulkanSpecIRHandleBuild(node); handle.Name != "" {
			*handles = append(*handles, handle)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRHandlesCollectWalk(child, handles)
	}
}

func vulkanSpecIRHandleBuild(typeNode *xmlSpecNode) VulkanSpecIRHandle {
	name := vulkanSpecIRHandleNameParse(typeNode)
	if name == "" || !strings.HasPrefix(name, "Vk") {
		return VulkanSpecIRHandle{}
	}

	macro := vulkanSpecIRHandleMacroParse(typeNode)
	dispatchable := macro == vulkanSpecIRHandleMacroDispatchable
	underlying := vulkanSpecIRHandleNonDispatchableUnderlying
	if dispatchable {
		underlying = vulkanSpecIRHandleDispatchableUnderlying
	}

	return VulkanSpecIRHandle{
		Name:           name,
		Underlying:     underlying,
		Dispatchable:   dispatchable,
		Parent:         xmlSpecNodeAttrValue(typeNode, "parent"),
		ObjectTypeEnum: xmlSpecNodeAttrValue(typeNode, "objtypeenum"),
		AliasOf:        xmlSpecNodeAttrValue(typeNode, "alias"),
		XMLDoc:         vulkanSpecIRHandleXMLDoc(typeNode, dispatchable),
	}
}

func vulkanSpecIRHandleNameParse(typeNode *xmlSpecNode) string {
	if name := xmlSpecNodeAttrValue(typeNode, "name"); name != "" {
		return name
	}

	for _, child := range typeNode.Children {
		if child.Name != "name" {
			continue
		}
		if name := strings.TrimSpace(child.Text); name != "" {
			return name
		}
	}

	return vulkanSpecIRHandleNameFromMacro(typeNode)
}

func vulkanSpecIRHandleNameFromMacro(typeNode *xmlSpecNode) string {
	macro := vulkanSpecIRHandleMacroParse(typeNode)
	if macro == "" {
		return ""
	}

	for _, child := range typeNode.Children {
		if child.Name != "name" {
			continue
		}
		if name := strings.TrimSpace(child.Text); name != "" {
			return name
		}
	}

	return ""
}

func vulkanSpecIRHandleMacroParse(typeNode *xmlSpecNode) string {
	for _, child := range typeNode.Children {
		if child.Name != "type" {
			continue
		}
		if macro := strings.TrimSpace(child.Text); macro != "" {
			return macro
		}
	}
	return ""
}

func vulkanSpecIRHandleXMLDoc(typeNode *xmlSpecNode, dispatchable bool) string {
	parts := make([]string, 0, 4)
	if comment := xmlSpecNodeDirectCommentsCollect(typeNode); comment != "" {
		parts = append(parts, comment)
	}
	if dispatchable {
		parts = append(parts, "is a dispatchable Vulkan handle.")
	} else {
		parts = append(parts, "is a non-dispatchable Vulkan handle.")
	}
	if parent := xmlSpecNodeAttrValue(typeNode, "parent"); parent != "" {
		parts = append(parts, "Parent object type is "+parent+".")
	}
	if objectType := xmlSpecNodeAttrValue(typeNode, "objtypeenum"); objectType != "" {
		parts = append(parts, "Object type token is "+objectType+".")
	}
	return vulkanSpecIRDocJoin(parts...)
}

func vulkanSpecIRHandlesResolveAliases(handles []VulkanSpecIRHandle) {
	byName := make(map[string]VulkanSpecIRHandle, len(handles))
	for _, handle := range handles {
		if handle.AliasOf == "" {
			byName[handle.Name] = handle
		}
	}

	for i := range handles {
		if handles[i].AliasOf == "" {
			continue
		}
		target, ok := byName[handles[i].AliasOf]
		if !ok {
			continue
		}
		handles[i].Underlying = target.Underlying
		handles[i].Dispatchable = target.Dispatchable
		if handles[i].Parent == "" {
			handles[i].Parent = target.Parent
		}
		if handles[i].ObjectTypeEnum == "" {
			handles[i].ObjectTypeEnum = target.ObjectTypeEnum
		}
	}
}
