package main

import (
	"sort"
	"strings"
)

func vulkanSpecIRCommandsCollect(root *xmlSpecNode, registry VulkanSpecIRTypeRegistry) []VulkanSpecIRCommand {
	if root == nil {
		return nil
	}

	byName := make(map[string]VulkanSpecIRCommand)
	var aliasOnly []VulkanSpecIRCommand

	vulkanSpecIRCommandsCollectWalk(root, registry, byName, &aliasOnly)
	vulkanSpecIRCommandsApplyAliases(byName, aliasOnly)

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
			XMLDoc:      xmlSpecNodeDirectCommentsCollect(cmdNode),
		})
		return
	}

	name, returnGoType, returnsUnsafe, params, ok := vulkanSpecIRProtoParamsBuild(cmdNode, registry, vulkanSpecIRCommandParamInclude)
	if !ok || !strings.HasPrefix(name, "vk") {
		return
	}

	command := VulkanSpecIRCommand{
		Name:          name,
		PFNTypeName:   vulkanSpecIRCommandPFNTypeName(name),
		ReturnGoType:  returnGoType,
		ReturnsUnsafe: returnsUnsafe,
		Params:        params,
		XMLDoc:        xmlSpecNodeDirectCommentsCollect(cmdNode),
		APIPriority:   vulkanSpecIRCommandPriority(cmdNode),
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
