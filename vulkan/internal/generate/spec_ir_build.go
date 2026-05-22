package main

import "sort"

/*
vulkanSpecIRBuild constructs the registry IR from a parsed XML document tree.

[Parameters]
root is the document root from xmlSpecTreeParse.

[Returns]
A populated VulkanSpecIR with sorted entries. Does not mutate root.
*/
func vulkanSpecIRBuild(root *xmlSpecNode) VulkanSpecIR {
	basetypes := vulkanSpecIRBasetypesCollect(root)
	constants := vulkanSpecIRConstantsCollect(root)
	bitmaskRegistry := vulkanSpecIRBitmaskRegistryCollect(root)
	enums := vulkanSpecIREnumsCollect(root)
	flags := vulkanSpecIRFlagsCollect(root, bitmaskRegistry)
	flags = vulkanSpecIRExtensionEnumsApply(root, &enums, flags)
	handles := vulkanSpecIRHandlesCollect(root)
	typeRegistry := vulkanSpecIRTypeRegistryCollect(root)
	funcpointers := vulkanSpecIRFuncpointersCollect(root, typeRegistry)
	commands := vulkanSpecIRCommandsCollect(root, typeRegistry)
	structs := vulkanSpecIRStructsCollect(root, typeRegistry)
	apiVersions := vulkanSpecIRApiVersionsCollect(root)
	extensions := vulkanSpecIRExtensionsCollect(root)
	features := vulkanSpecIRRegistryFeaturesCollect(apiVersions, extensions)

	sort.Slice(basetypes, func(i, j int) bool { return basetypes[i].Name < basetypes[j].Name })
	sort.Slice(funcpointers, func(i, j int) bool { return funcpointers[i].Name < funcpointers[j].Name })
	sort.Slice(commands, func(i, j int) bool { return commands[i].PFNTypeName < commands[j].PFNTypeName })
	sort.Slice(structs, func(i, j int) bool {
		if structs[i].AliasOf != structs[j].AliasOf {
			return structs[i].AliasOf == ""
		}
		return structs[i].Name < structs[j].Name
	})
	sort.Slice(enums, func(i, j int) bool { return enums[i].Name < enums[j].Name })
	sort.Slice(flags, func(i, j int) bool {
		return flags[i].AggregateName < flags[j].AggregateName
	})

	return VulkanSpecIR{
		Basetypes:    basetypes,
		Constants:    constants,
		Enums:        enums,
		Flags:        flags,
		Funcpointers: funcpointers,
		Commands:     commands,
		Handles:      handles,
		Structs:      structs,
		ApiVersions:  apiVersions,
		Extensions:   extensions,
		Features:     features,
	}
}
