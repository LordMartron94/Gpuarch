package main

import (
	"codegen"
	gocode "codegen/go"
	"fmt"
	"foundation/system"
	"sort"
	"strconv"
	"strings"
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

	if err := writeLoaderFile(loaderDir, loaderManifestFile, elements); err != nil {
		return err
	}
	return writeVulkanCommandsManifestTargetTable(loaderDir, filtered)
}

func loaderManifestFilePreamble() []codegen.FileElement {
	elements := make([]codegen.FileElement, 0, 8)
	elements = append(elements, gocode.GoGeneratedFileHeader(vulkanLoaderGeneratorTool, time.Now())...)
	elements = append(elements, gocode.FileElementFrom(gocode.DeclPackage("loader")))
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

	return gocode.DeclFunc(
		funcName,
		[]gocode.ParamType{
			{Name: "manifest", Type: gocode.TypeExprPointer(gocode.TypeExprNamed("VulkanCommandManifest"))},
		},
		[]gocode.TypeExpr{gocode.TypeExprNamed("error")},
		gocode.StmtBlock(
			gocode.StmtReturn(
				gocode.ExprCall(
					gocode.ExprIdent("VulkanCommandManifestAdd"),
					gocode.ExprIdent("manifest"),
					gocode.ExprIdent(fieldConst),
				),
			),
		),
	)
}

func writeVulkanCommandsManifestTargetTable(loaderDir string, commands []VulkanSpecIRCommand) error {
	var body strings.Builder
	body.WriteString("\n/*\nvulkanCommandsManifestTargetByField maps VulkanCommandManifestField to a PFN holder on VulkanCommands.\n*/\n")
	body.WriteString("var vulkanCommandsManifestTargetByField = [...]func(*VulkanCommands) any{\n")
	body.WriteString("\tnil,\n")
	for _, command := range commands {
		endpoint := vulkanLoaderEndpointGoName(command.CommandEndpoint)
		body.WriteString(fmt.Sprintf(
			"\tfunc(c *VulkanCommands) any { return &c.%s.%s },\n",
			endpoint,
			command.GoFieldName,
		))
	}
	body.WriteString("}\n")

	path := system.PathJoin(loaderDir, loaderManifestFile)
	existingBytes, err := system.FileReadAllBytes(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", loaderManifestFile, err)
	}
	if err := system.FileWriteString(path, string(existingBytes)+body.String()); err != nil {
		return fmt.Errorf("append manifest target table: %w", err)
	}
	return nil
}
