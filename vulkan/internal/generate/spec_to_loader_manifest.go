package main

import (
	"codegen"
	gocode "codegen/go"
	"sort"
	"strconv"
	"time"
)

const (
	loaderManifestFile              = "loader_manifest_gen.go"
	vulkanCommandLoaderTierGlobal   = 0
	vulkanCommandLoaderTierInstance = 1
	vulkanCommandLoaderTierDevice   = 2
)

func generateLoaderManifestContent(commands []VulkanSpecIRCommand, loaderDir string) error {
	filtered := make([]VulkanSpecIRCommand, 0, len(commands))
	for _, command := range commands {
		if command.GoFieldName == "" || command.PFNTypeName == "" || command.Name == "" {
			continue
		}
		filtered = append(filtered, command)
	}
	if len(filtered) == 0 {
		return nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].GoFieldName < filtered[j].GoFieldName
	})

	elements := loaderManifestFilePreamble()
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestFieldTypeDecl()))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestFieldConstantsDecl(filtered)))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestCatalogEntryStructDecl()))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestCatalogDecl(filtered)))
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestEntryTypeDecl()))
	blankLine(&elements)
	elements = append(elements, vulkanLoaderManifestTypedAddElements(filtered)...)

	return writeLoaderFile(loaderDir, loaderManifestFile, elements)
}

func loaderManifestFilePreamble() []codegen.FileElement {
	elements := make([]codegen.FileElement, 0, 8)
	elements = append(elements, gocode.GoGeneratedFileHeader(vulkanLoaderGeneratorTool, time.Now())...)
	elements = append(elements, gocode.FileElementFrom(gocode.DeclPackage("loader")))
	elements = append(elements, gocode.FileElementFrom(gocode.DeclImportBlock(
		vulkanLoaderBindingsImport,
	)))
	blankLine(&elements)
	return elements
}

func vulkanLoaderManifestFieldTypeDecl() gocode.TypeDefinedDecl {
	return gocode.DeclTypeDefined(
		"VulkanCommandManifestField",
		gocode.TypeExprNamed("uint32"),
		true,
		"VulkanCommandManifestField identifies one Vulkan command for selective runtime loading.",
	)
}

func vulkanLoaderManifestFieldConstantName(goFieldName string) string {
	return "VulkanCommandManifestField" + goFieldName
}

func vulkanLoaderManifestFieldConstantsDecl(commands []VulkanSpecIRCommand) gocode.ConstGroupDecl {
	invalidType := gocode.TypeExprNamed("VulkanCommandManifestField")
	specs := []gocode.ConstSpec{
		{
			Name:  "VulkanCommandManifestFieldInvalid",
			Type:  &invalidType,
			Value: "0",
		},
		{
			Name:  "vulkanCommandManifestFieldMax",
			Type:  &invalidType,
			Value: strconv.Itoa(len(commands)),
		},
	}
	for fieldIndex, command := range commands {
		specs = append(specs, gocode.ConstSpec{
			Name:  vulkanLoaderManifestFieldConstantName(command.GoFieldName),
			Type:  &invalidType,
			Value: strconv.Itoa(fieldIndex + 1),
		})
	}
	return gocode.DeclConstGroup(
		specs,
		"VulkanCommandManifestField constants select commands for VulkanCommandManifestAdd at runtime.",
		false,
	)
}

func vulkanLoaderManifestCatalogEntryStructDecl() gocode.TypeStructDecl {
	return gocode.DeclTypeStruct("vulkanCommandCatalogEntry", []gocode.StructFieldDecl{
		gocode.StructFieldTypeDoc("tier", gocode.TypeExprNamed("int"), "tier selects global, instance, or device proc addr resolution."),
		gocode.StructFieldTypeDoc("vkName", gocode.TypeExprNamed("string"), "vkName is the Vulkan entry point name."),
	}, "vulkanCommandCatalogEntry is registry metadata used when loading a runtime manifest.")
}

func vulkanLoaderManifestCatalogDecl(commands []VulkanSpecIRCommand) gocode.VarDecl {
	entries := make([]codegen.Expr, 0, len(commands)+1)
	entries = append(entries, gocode.ExprCompositeLit("vulkanCommandCatalogEntry", nil))

	for _, command := range commands {
		entries = append(entries, gocode.ExprCompositeLit("vulkanCommandCatalogEntry", []gocode.FieldInit{
			{Name: "tier", Value: vulkanLoaderManifestTierIntLit(command.LoaderTier)},
			{Name: "vkName", Value: gocode.ExprStringLit(command.Name)},
		}))
	}

	return gocode.DeclVar(
		"vulkanCommandCatalog",
		gocode.ExprSliceCompositeLit("vulkanCommandCatalogEntry", entries),
		"vulkanCommandCatalog maps VulkanCommandManifestField values to loader metadata.",
	)
}

func vulkanLoaderManifestTierIntLit(tier VulkanSpecIRCommandLoaderTier) *gocode.IntLitExpr {
	switch tier {
	case VulkanCommandLoaderTierGlobal:
		return gocode.ExprIntLit(vulkanCommandLoaderTierGlobal)
	case VulkanCommandLoaderTierDevice:
		return gocode.ExprIntLit(vulkanCommandLoaderTierDevice)
	default:
		return gocode.ExprIntLit(vulkanCommandLoaderTierInstance)
	}
}

func vulkanLoaderManifestEntryTypeDecl() gocode.TypeStructDecl {
	return gocode.DeclTypeStruct("vulkanCommandManifestEntry", []gocode.StructFieldDecl{
		gocode.StructFieldTypeDoc("field", gocode.TypeExprNamed("VulkanCommandManifestField"), "field selects the catalog entry."),
		gocode.StructFieldTypeDoc("target", gocode.TypeExprNamed("any"), "target is the PFN field pointer to bind."),
	}, "vulkanCommandManifestEntry is one command registered on VulkanCommandManifest.")
}

func vulkanLoaderManifestTypedAddElements(commands []VulkanSpecIRCommand) []codegen.FileElement {
	elements := make([]codegen.FileElement, 0, len(commands))
	for commandIndex, command := range commands {
		if commandIndex > 0 {
			blankLine(&elements)
		}
		elements = append(elements, gocode.FileElementFrom(vulkanLoaderManifestTypedAddDecl(command)))
	}
	return elements
}

func vulkanLoaderManifestTypedAddDecl(command VulkanSpecIRCommand) gocode.FuncDecl {
	funcName := "VulkanCommandManifestAdd" + command.GoFieldName
	fieldConst := vulkanLoaderManifestFieldConstantName(command.GoFieldName)
	pfnType, err := gocode.TypeExprFromGoTypeString(vulkanLoaderQualifyBindingsType(command.PFNTypeName))
	if err != nil {
		pfnType = gocode.TypeExprNamed(command.PFNTypeName)
	}

	return gocode.DeclFunc(
		funcName,
		[]gocode.ParamType{
			{Name: "manifest", Type: gocode.TypeExprPointer(gocode.TypeExprNamed("VulkanCommandManifest"))},
			{Name: "target", Type: gocode.TypeExprPointer(pfnType)},
		},
		[]gocode.TypeExpr{gocode.TypeExprNamed("error")},
		gocode.StmtBlock(
			gocode.StmtReturn(
				gocode.ExprCall(
					gocode.ExprIdent("VulkanCommandManifestAdd"),
					gocode.ExprIdent("manifest"),
					gocode.ExprIdent(fieldConst),
					gocode.ExprIdent("target"),
				),
			),
		),
	)
}
