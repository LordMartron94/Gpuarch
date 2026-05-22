package main

import (
	"codegen"
	gocode "codegen/go"
	"strconv"
	"strings"
)

const registryBindingFile = "bindings_registry_gen.go"

const (
	vkRegistryFeatureRefTypeName = "VkRegistryFeatureRef"
	vkRegistryExtensionTypeName  = "VkRegistryExtension"
	vkRegistryApiVersionTypeName = "VkRegistryApiVersion"
	vkRegistryFeatureTypeName    = "VkRegistryFeature"
	vkRegistryExtensionsVarName  = "VkRegistryExtensions"
	vkRegistryApiVersionsVarName = "VkRegistryApiVersions"
	vkRegistryFeaturesVarName    = "VkRegistryFeatures"
)

func generateRegistryContent(ir VulkanSpecIR, bindingsDir string, _ VulkanRefpageCorpus) error {
	elements := bindingFilePreamble()
	elements = append(elements, vulkanRegistryTypeElements()...)
	blankLine(&elements)
	elements = append(elements, vulkanRegistryExtensionNameConstants(ir.Extensions)...)
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclVar(
			vkRegistryExtensionsVarName,
			vulkanRegistryExtensionsComposite(ir.Extensions),
			"VkRegistryExtensions lists every extension entry from the Khronos Vulkan registry vk.xml.",
		),
	))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclVar(
			vkRegistryApiVersionsVarName,
			vulkanRegistryApiVersionsComposite(ir.ApiVersions),
			"VkRegistryApiVersions lists core and internal API version feature blocks from vk.xml.",
		),
	))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclVar(
			vkRegistryFeaturesVarName,
			vulkanRegistryFeaturesComposite(ir.Features),
			"VkRegistryFeatures lists capability flags and the structs that expose them.",
		),
	))
	return writeBindingFile(bindingsDir, registryBindingFile, elements)
}

func vulkanRegistryTypeElements() []codegen.FileElement {
	return []codegen.FileElement{
		gocode.FileElementFrom(gocode.DeclTypeStruct(
			vkRegistryFeatureRefTypeName,
			[]gocode.StructFieldDecl{
				gocode.StructFieldTypeDoc("Name", gocode.TypeExprNamed("string"), "Name is the Vulkan feature struct member (for example timelineSemaphore)."),
				gocode.StructFieldTypeDoc("StructType", gocode.TypeExprNamed("string"), "StructType is the VkPhysicalDevice*Features struct that contains the member."),
				gocode.StructFieldTypeDoc("RequireDepends", gocode.TypeExprNamed("string"), "RequireDepends is the optional extension or feature dependency on the parent require block."),
			},
			"VkRegistryFeatureRef links a capability flag to the struct that exposes it.",
		)),
		gocode.FileElementFrom(gocode.DeclTypeStruct(
			vkRegistryExtensionTypeName,
			[]gocode.StructFieldDecl{
				gocode.StructFieldTypeDoc("Name", gocode.TypeExprNamed("string"), "Name is the extension name (for example VK_KHR_surface)."),
				gocode.StructFieldTypeDoc("Number", gocode.TypeExprNamed("uint32"), "Number is the registry extension number."),
				gocode.StructFieldTypeDoc("ExtensionType", gocode.TypeExprNamed("string"), "ExtensionType is instance or device."),
				gocode.StructFieldTypeDoc("EnableName", gocode.TypeExprNamed("string"), "EnableName is the string passed to vkEnumerateDeviceExtensionProperties."),
				gocode.StructFieldTypeDoc("SpecVersion", gocode.TypeExprNamed("uint32"), "SpecVersion is the extension spec version from vk.xml."),
				gocode.StructFieldTypeDoc("Depends", gocode.TypeExprNamed("string"), "Depends lists required extensions or features."),
				gocode.StructFieldTypeDoc("Supported", gocode.TypeExprNamed("string"), "Supported lists APIs that expose the extension."),
				gocode.StructFieldTypeDoc("Ratified", gocode.TypeExprNamed("string"), "Ratified lists ratification status per API."),
				gocode.StructFieldTypeDoc("Platform", gocode.TypeExprNamed("string"), "Platform is set for platform-specific WSI extensions."),
				gocode.StructFieldTypeDoc("Author", gocode.TypeExprNamed("string"), "Author is the vendor tag from vk.xml."),
				gocode.StructFieldTypeDoc("Contact", gocode.TypeExprNamed("string"), "Contact is the registry contact field."),
				gocode.StructFieldTypeDoc("NoFeatures", gocode.TypeExprNamed("bool"), "NoFeatures is true when the extension defines no feature struct members."),
				gocode.StructFieldType("Features", vulkanRegistryFeatureRefSliceType()),
			},
			"VkRegistryExtension describes one extension from the Khronos Vulkan registry.",
		)),
		gocode.FileElementFrom(gocode.DeclTypeStruct(
			vkRegistryApiVersionTypeName,
			[]gocode.StructFieldDecl{
				gocode.StructFieldTypeDoc("Name", gocode.TypeExprNamed("string"), "Name is the feature block name (for example VK_VERSION_1_3)."),
				gocode.StructFieldTypeDoc("Number", gocode.TypeExprNamed("string"), "Number is the API version number from vk.xml."),
				gocode.StructFieldTypeDoc("Depends", gocode.TypeExprNamed("string"), "Depends lists required feature blocks."),
				gocode.StructFieldTypeDoc("Comment", gocode.TypeExprNamed("string"), "Comment is the registry comment on the feature block."),
				gocode.StructFieldTypeDoc("Internal", gocode.TypeExprNamed("bool"), "Internal is true for apitype=internal feature blocks."),
				gocode.StructFieldType("Features", vulkanRegistryFeatureRefSliceType()),
			},
			"VkRegistryApiVersion describes a VK_VERSION or internal base/compute/graphics feature block.",
		)),
		gocode.FileElementFrom(gocode.DeclTypeStruct(
			vkRegistryFeatureTypeName,
			[]gocode.StructFieldDecl{
				gocode.StructFieldTypeDoc("Name", gocode.TypeExprNamed("string"), "Name is the Vulkan feature struct member."),
				gocode.StructFieldTypeDoc("StructType", gocode.TypeExprNamed("string"), "StructType is the VkPhysicalDevice*Features struct containing the member."),
				gocode.StructFieldTypeDoc("Provider", gocode.TypeExprNamed("string"), "Provider is the extension or API version feature block that introduced the capability."),
				gocode.StructFieldTypeDoc("RequireDepends", gocode.TypeExprNamed("string"), "RequireDepends is the optional dependency on the parent require block."),
			},
			"VkRegistryFeature is one registry capability entry with its provider.",
		)),
	}
}

func vulkanRegistryExtensionNameConstants(extensions []VulkanSpecIRExtension) []codegen.FileElement {
	specs := make([]gocode.ConstSpec, 0, len(extensions))
	for _, extension := range extensions {
		if extension.ExtensionNameConst == "" || extension.EnableName == "" {
			continue
		}
		specs = append(specs, gocode.ConstSpecNew(
			extension.ExtensionNameConst,
			nil,
			quoteGoString(extension.EnableName),
			vulkanSpecIRDocFormatExported(extension.ExtensionNameConst, ""),
		))
	}
	if len(specs) == 0 {
		return nil
	}
	return []codegen.FileElement{
		gocode.FileElementFrom(gocode.DeclConstGroup(
			specs,
			"Extension enable strings from the Khronos Vulkan registry.",
			true,
		)),
	}
}

func vulkanRegistryExtensionsComposite(extensions []VulkanSpecIRExtension) *gocode.CompositeLitExpr {
	elements := make([]codegen.Expr, 0, len(extensions))
	for _, extension := range extensions {
		elements = append(elements, vulkanRegistryExtensionLiteral(extension))
	}
	return gocode.ExprSliceCompositeLit(vkRegistryExtensionTypeName, elements)
}

func vulkanRegistryExtensionLiteral(extension VulkanSpecIRExtension) *gocode.CompositeLitExpr {
	featureElements := make([]codegen.Expr, 0, len(extension.Features))
	for _, feature := range extension.Features {
		featureElements = append(featureElements, vulkanRegistryFeatureRefLiteral(feature))
	}

	return gocode.ExprCompositeLit(vkRegistryExtensionTypeName, []gocode.FieldInit{
		{Name: "Name", Value: gocode.ExprStringLit(extension.Name)},
		{Name: "Number", Value: gocode.ExprIntLit(int64(vulkanRegistryParseUint32(extension.Number)))},
		{Name: "ExtensionType", Value: gocode.ExprStringLit(extension.ExtensionType)},
		{Name: "EnableName", Value: gocode.ExprStringLit(extension.EnableName)},
		{Name: "SpecVersion", Value: gocode.ExprIntLit(int64(vulkanRegistryParseUint32(extension.SpecVersion)))},
		{Name: "Depends", Value: gocode.ExprStringLit(extension.Depends)},
		{Name: "Supported", Value: gocode.ExprStringLit(extension.Supported)},
		{Name: "Ratified", Value: gocode.ExprStringLit(extension.Ratified)},
		{Name: "Platform", Value: gocode.ExprStringLit(extension.Platform)},
		{Name: "Author", Value: gocode.ExprStringLit(extension.Author)},
		{Name: "Contact", Value: gocode.ExprStringLit(extension.Contact)},
		{Name: "NoFeatures", Value: gocode.ExprBoolLit(extension.NoFeatures)},
		{Name: "Features", Value: gocode.ExprSliceCompositeLit(vkRegistryFeatureRefTypeName, featureElements)},
	})
}

func vulkanRegistryApiVersionsComposite(versions []VulkanSpecIRApiVersion) *gocode.CompositeLitExpr {
	elements := make([]codegen.Expr, 0, len(versions))
	for _, version := range versions {
		elements = append(elements, vulkanRegistryApiVersionLiteral(version))
	}
	return gocode.ExprSliceCompositeLit(vkRegistryApiVersionTypeName, elements)
}

func vulkanRegistryApiVersionLiteral(version VulkanSpecIRApiVersion) *gocode.CompositeLitExpr {
	featureElements := make([]codegen.Expr, 0, len(version.Features))
	for _, feature := range version.Features {
		featureElements = append(featureElements, vulkanRegistryFeatureRefLiteral(feature))
	}

	return gocode.ExprCompositeLit(vkRegistryApiVersionTypeName, []gocode.FieldInit{
		{Name: "Name", Value: gocode.ExprStringLit(version.Name)},
		{Name: "Number", Value: gocode.ExprStringLit(version.Number)},
		{Name: "Depends", Value: gocode.ExprStringLit(version.Depends)},
		{Name: "Comment", Value: gocode.ExprStringLit(version.Comment)},
		{Name: "Internal", Value: gocode.ExprBoolLit(version.Internal)},
		{Name: "Features", Value: gocode.ExprSliceCompositeLit(vkRegistryFeatureRefTypeName, featureElements)},
	})
}

func vulkanRegistryFeaturesComposite(features []VulkanSpecIRFeature) *gocode.CompositeLitExpr {
	elements := make([]codegen.Expr, 0, len(features))
	for _, feature := range features {
		elements = append(elements, gocode.ExprCompositeLit(vkRegistryFeatureTypeName, []gocode.FieldInit{
			{Name: "Name", Value: gocode.ExprStringLit(feature.Name)},
			{Name: "StructType", Value: gocode.ExprStringLit(feature.StructType)},
			{Name: "Provider", Value: gocode.ExprStringLit(feature.Provider)},
			{Name: "RequireDepends", Value: gocode.ExprStringLit(feature.RequireDepends)},
		}))
	}
	return gocode.ExprSliceCompositeLit(vkRegistryFeatureTypeName, elements)
}

func vulkanRegistryFeatureRefLiteral(feature VulkanSpecIRFeatureRef) *gocode.CompositeLitExpr {
	return gocode.ExprCompositeLit(vkRegistryFeatureRefTypeName, []gocode.FieldInit{
		{Name: "Name", Value: gocode.ExprStringLit(feature.Name)},
		{Name: "StructType", Value: gocode.ExprStringLit(feature.StructType)},
		{Name: "RequireDepends", Value: gocode.ExprStringLit(feature.RequireDepends)},
	})
}

func vulkanRegistryParseUint32(value string) uint32 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(parsed)
}

func quoteGoString(value string) string {
	return strconv.Quote(value)
}

func vulkanRegistryFeatureRefSliceType() gocode.TypeExpr {
	typ, err := gocode.TypeExprFromGoTypeString("[]" + vkRegistryFeatureRefTypeName)
	if err != nil {
		return gocode.TypeExprNamed("[]" + vkRegistryFeatureRefTypeName)
	}
	return typ
}
