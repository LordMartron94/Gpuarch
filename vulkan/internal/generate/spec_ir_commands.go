package main

import (
	"sort"
	"strings"
	"unicode"
)

func vulkanSpecIRCommandsCollect(root *xmlSpecNode, registry VulkanSpecIRTypeRegistry) []VulkanSpecIRCommand {
	if root == nil {
		return nil
	}

	byName := make(map[string]VulkanSpecIRCommand)
	var aliasOnly []VulkanSpecIRCommand

	vulkanSpecIRCommandsCollectWalk(root, registry, byName, &aliasOnly)
	vulkanSpecIRCommandsApplyAliases(byName, aliasOnly)
	vulkanSpecIRCommandsFinalizeTiers(byName, aliasOnly)
	vulkanSpecIRCommandsFinalizeEndpoints(byName, aliasOnly)

	commands := make([]VulkanSpecIRCommand, 0, len(byName)+len(aliasOnly))
	for _, command := range byName {
		commands = append(commands, command)
	}
	for _, command := range aliasOnly {
		if command.AliasOf != "" {
			commands = append(commands, command)
		}
	}

	sort.Slice(commands, func(i, j int) bool {
		return commands[i].PFNTypeName < commands[j].PFNTypeName
	})
	return commands
}

func vulkanSpecIRCommandsCollectWalk(
	node *xmlSpecNode,
	registry VulkanSpecIRTypeRegistry,
	byName map[string]VulkanSpecIRCommand,
	aliasOnly *[]VulkanSpecIRCommand,
) {
	if node.Name == "command" {
		vulkanSpecIRCommandCollect(node, registry, byName, aliasOnly)
	}

	for _, child := range node.Children {
		vulkanSpecIRCommandsCollectWalk(child, registry, byName, aliasOnly)
	}
}

func vulkanSpecIRCommandCollect(
	cmdNode *xmlSpecNode,
	registry VulkanSpecIRTypeRegistry,
	byName map[string]VulkanSpecIRCommand,
	aliasOnly *[]VulkanSpecIRCommand,
) {
	if !vulkanSpecIRCommandInclude(cmdNode) {
		return
	}

	aliasName := strings.TrimSpace(xmlSpecNodeAttrValue(cmdNode, "alias"))
	if aliasName != "" && !vulkanSpecIRCommandHasProto(cmdNode) {
		name := strings.TrimSpace(xmlSpecNodeAttrValue(cmdNode, "name"))
		if name == "" {
			return
		}
		*aliasOnly = append(*aliasOnly, VulkanSpecIRCommand{
			Name:        name,
			PFNTypeName: vulkanSpecIRCommandPFNTypeName(name),
			AliasOf:     vulkanSpecIRCommandPFNTypeName(aliasName),
			AliasOfName: aliasName,
			GoFieldName: vulkanSpecIRCommandGoFieldName(name),
			XMLDoc:      xmlSpecNodeDirectCommentsCollect(cmdNode),
		})
		return
	}

	name, returnGoType, returnsUnsafe, params, ok := vulkanSpecIRProtoParamsBuild(cmdNode, registry, vulkanSpecIRCommandParamInclude)
	if !ok || !strings.HasPrefix(name, "vk") {
		return
	}

	command := VulkanSpecIRCommand{
		Name:            name,
		PFNTypeName:     vulkanSpecIRCommandPFNTypeName(name),
		ReturnGoType:    returnGoType,
		ReturnsUnsafe:   returnsUnsafe,
		Params:          params,
		LoaderTier:      vulkanSpecIRCommandLoaderTierInfer(name, params),
		CommandEndpoint: vulkanSpecIRCommandEndpointInfer(name, params),
		GoFieldName:     vulkanSpecIRCommandGoFieldName(name),
		XMLDoc:          xmlSpecNodeDirectCommentsCollect(cmdNode),
		APIPriority:     vulkanSpecIRCommandPriority(cmdNode),
	}

	existing, exists := byName[name]
	if !exists || command.APIPriority > existing.APIPriority {
		byName[name] = command
	}
}

func vulkanSpecIRCommandsApplyAliases(
	byName map[string]VulkanSpecIRCommand,
	aliasOnly []VulkanSpecIRCommand,
) {
	for i := range aliasOnly {
		alias := &aliasOnly[i]
		if alias.AliasOfName == "" {
			continue
		}
		if target, ok := byName[alias.AliasOfName]; ok {
			alias.AliasOf = target.PFNTypeName
		}
	}
}

func vulkanSpecIRCommandPFNTypeName(vkName string) string {
	return "PFN_" + vkName
}

func vulkanSpecIRCommandGoFieldName(vkName string) string {
	name := strings.TrimPrefix(vkName, "vk")
	if name == "" {
		return vkName
	}
	runes := []rune(name)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func vulkanSpecIRCommandsFinalizeTiers(
	byName map[string]VulkanSpecIRCommand,
	aliasOnly []VulkanSpecIRCommand,
) {
	for i := range aliasOnly {
		if aliasOnly[i].LoaderTier != "" {
			continue
		}
		if target, ok := byName[aliasOnly[i].AliasOfName]; ok {
			aliasOnly[i].LoaderTier = target.LoaderTier
		} else {
			aliasOnly[i].LoaderTier = VulkanCommandLoaderTierInstance
		}
	}
}

func vulkanSpecIRCommandsFinalizeEndpoints(
	byName map[string]VulkanSpecIRCommand,
	aliasOnly []VulkanSpecIRCommand,
) {
	for i := range aliasOnly {
		if aliasOnly[i].CommandEndpoint != "" {
			continue
		}
		if target, ok := byName[aliasOnly[i].AliasOfName]; ok {
			aliasOnly[i].CommandEndpoint = target.CommandEndpoint
		} else {
			aliasOnly[i].CommandEndpoint = VulkanCommandEndpointInstance
		}
	}
}

func vulkanSpecIRCommandParamBaseType(goType string) string {
	base := strings.TrimPrefix(goType, "*")
	return strings.TrimPrefix(base, "[]")
}

func vulkanSpecIRCommandEndpointInfer(name string, params []VulkanSpecIRFuncpointerParam) VulkanSpecIRCommandEndpoint {
	switch vulkanSpecIRCommandLoaderTierInfer(name, params) {
	case VulkanCommandLoaderTierGlobal:
		return VulkanCommandEndpointGlobal
	}

	for _, param := range params {
		base := vulkanSpecIRCommandParamBaseType(param.GoType)
		switch {
		case base == "VkInstance":
			return VulkanCommandEndpointInstance
		case base == "VkPhysicalDevice":
			return VulkanCommandEndpointPhysicalDevice
		case strings.HasPrefix(base, "VkSurface"):
			return VulkanCommandEndpointSurface
		case base == "VkSwapchainKHR":
			return VulkanCommandEndpointSwapchain
		case base == "VkDevice":
			return VulkanCommandEndpointDevice
		case base == "VkQueue":
			return VulkanCommandEndpointQueue
		case base == "VkCommandBuffer":
			return VulkanCommandEndpointCommandBuffer
		}
	}

	if vulkanSpecIRCommandLoaderTierInfer(name, params) == VulkanCommandLoaderTierDevice {
		return VulkanCommandEndpointDevice
	}
	return VulkanCommandEndpointInstance
}

func vulkanSpecIRCommandLoaderTierInfer(name string, params []VulkanSpecIRFuncpointerParam) VulkanSpecIRCommandLoaderTier {
	switch name {
	case "vkCreateInstance",
		"vkEnumerateInstanceExtensionProperties",
		"vkEnumerateInstanceLayerProperties",
		"vkEnumerateInstanceVersion",
		"vkGetInstanceProcAddr":
		return VulkanCommandLoaderTierGlobal
	}

	for _, param := range params {
		base := strings.TrimPrefix(param.GoType, "*")
		base = strings.TrimPrefix(base, "[]")
		switch base {
		case "VkDevice":
			return VulkanCommandLoaderTierDevice
		case "VkCommandBuffer", "VkQueue":
			return VulkanCommandLoaderTierDevice
		case "VkInstance":
			return VulkanCommandLoaderTierInstance
		case "VkPhysicalDevice":
			return VulkanCommandLoaderTierInstance
		}
	}

	return VulkanCommandLoaderTierInstance
}

func vulkanSpecIRCommandHasProto(cmdNode *xmlSpecNode) bool {
	for _, child := range cmdNode.Children {
		if child.Name == "proto" {
			return true
		}
	}
	return false
}

func vulkanSpecIRCommandInclude(cmdNode *xmlSpecNode) bool {
	export := strings.TrimSpace(xmlSpecNodeAttrValue(cmdNode, "export"))
	if export != "" && !strings.Contains(export, "vulkan") {
		return false
	}
	return vulkanSpecIRCommandVulkanAPI(cmdNode)
}

func vulkanSpecIRCommandParamInclude(paramNode *xmlSpecNode) bool {
	return vulkanSpecIRCommandVulkanAPI(paramNode)
}

func vulkanSpecIRCommandVulkanAPI(cmdNode *xmlSpecNode) bool {
	api := strings.TrimSpace(xmlSpecNodeAttrValue(cmdNode, "api"))
	if api == "" {
		return true
	}
	for _, part := range strings.Split(api, ",") {
		part = strings.TrimSpace(part)
		if part == "vulkan" || part == "vulkanbase" {
			return true
		}
	}
	return false
}

func vulkanSpecIRCommandPriority(cmdNode *xmlSpecNode) int {
	api := strings.TrimSpace(xmlSpecNodeAttrValue(cmdNode, "api"))
	if api == "" || strings.Contains(api, "vulkan") {
		return 2
	}
	if strings.Contains(api, "vulkansc") {
		return 1
	}
	return 0
}
