package main

import (
	"codegen"
	gocode "codegen/go"
	"fmt"
	"foundation/system"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	loaderCommandsFile         = "loader_commands_gen.go"
	vulkanLoaderGeneratorTool  = "gpuarch Vulkan loader generator"
	vulkanLoaderBindingsImport = "gpuarch/vulkan/bindings"

	vulkanLoaderGlobalCommandsType   = "VulkanGlobalCommands"
	vulkanLoaderInstanceCommandsType = "VulkanInstanceCommands"
	vulkanLoaderDeviceCommandsType   = "VulkanDeviceCommands"
	vulkanLoaderCommandMappingType   = "vulkanCommandMapping"
)

func specToLoaderConvert(ir VulkanSpecIR, loaderDir string, corpus VulkanRefpageCorpus) error {
	if err := generateLoaderCommandsContent(ir.Commands, loaderDir, corpus); err != nil {
		return err
	}
	if err := generateLoaderManifestContent(ir.Commands, loaderDir); err != nil {
		return err
	}
	return vulkanLoaderFormat(loaderDir)
}

func generateLoaderCommandsContent(commands []VulkanSpecIRCommand, loaderDir string, corpus VulkanRefpageCorpus) error {
	if len(commands) == 0 {
		return nil
	}

	byTier := map[VulkanSpecIRCommandLoaderTier][]VulkanSpecIRCommand{
		VulkanCommandLoaderTierGlobal:   {},
		VulkanCommandLoaderTierInstance: {},
		VulkanCommandLoaderTierDevice:   {},
	}
	for _, command := range commands {
		if command.GoFieldName == "" || command.PFNTypeName == "" {
			continue
		}
		byTier[command.LoaderTier] = append(byTier[command.LoaderTier], command)
	}

	elements := loaderFilePreamble()
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderCommandMappingTypeDecl()))
	blankLine(&elements)

	tierSpecs := []struct {
		tier     VulkanSpecIRCommandLoaderTier
		typeName string
		section  string
		loadName string
	}{
		{VulkanCommandLoaderTierGlobal, vulkanLoaderGlobalCommandsType, "Global commands", "VulkanGlobalCommandsLoad"},
		{VulkanCommandLoaderTierInstance, vulkanLoaderInstanceCommandsType, "Instance commands", "VulkanInstanceCommandsLoad"},
		{VulkanCommandLoaderTierDevice, vulkanLoaderDeviceCommandsType, "Device commands", "VulkanDeviceCommandsLoad"},
	}

	for _, spec := range tierSpecs {
		tierCommands := byTier[spec.tier]
		if len(tierCommands) == 0 {
			continue
		}
		sort.Slice(tierCommands, func(i, j int) bool {
			return tierCommands[i].GoFieldName < tierCommands[j].GoFieldName
		})
		elements = append(elements, gocode.FileElementFrom(
			vulkanLoaderCommandsStructDecl(spec.typeName, spec.section, tierCommands, corpus),
		))
		blankLine(&elements)
		elements = append(elements, gocode.FileElementFrom(
			vulkanLoaderCommandsLoadFuncDecl(spec.loadName, spec.typeName, spec.tier, tierCommands),
		))
		blankLine(&elements)
	}

	return writeLoaderFile(loaderDir, loaderCommandsFile, elements)
}

func loaderFilePreamble() []codegen.FileElement {
	elements := make([]codegen.FileElement, 0, 8)
	elements = append(elements, gocode.GoGeneratedFileHeader(vulkanLoaderGeneratorTool, time.Now())...)
	elements = append(elements, gocode.FileElementFrom(gocode.DeclPackage("loader")))
	elements = append(elements, gocode.FileElementFrom(gocode.DeclImportBlock(
		"fmt",
		vulkanLoaderBindingsImport,
		"syscore",
	)))
	blankLine(&elements)
	return elements
}

func vulkanLoaderCommandMappingTypeDecl() gocode.TypeStructDecl {
	return gocode.DeclTypeStruct(vulkanLoaderCommandMappingType, []gocode.StructFieldDecl{
		gocode.StructFieldTypeDoc("target", gocode.TypeExprNamed("any"), "target is the command field to bind."),
		gocode.StructFieldTypeDoc("cName", gocode.TypeExprPointer(gocode.TypeExprNamed("byte")), "cName is the NUL-terminated Vulkan entry point name."),
		gocode.StructFieldTypeDoc("debugName", gocode.TypeExprNamed("string"), "debugName is the entry point name for error messages."),
	}, "vulkanCommandMapping describes one proc addr bind target.")
}

func vulkanLoaderCommandsStructDecl(
	typeName string,
	section string,
	commands []VulkanSpecIRCommand,
	corpus VulkanRefpageCorpus,
) gocode.TypeStructDecl {
	fields := make([]gocode.StructFieldDecl, 0, len(commands))
	for commandIndex, command := range commands {
		leading := []codegen.Node{gocode.LayoutBlankLineNode()}
		if commandIndex == 0 && section != "" {
			leading = append(leading, gocode.LineComment{Text: section})
		}

		pfnType, err := gocode.TypeExprFromGoTypeString(vulkanLoaderQualifyBindingsType(command.PFNTypeName))
		if err != nil {
			continue
		}

		refpageName := command.Name
		if refpageName == "" {
			refpageName = command.PFNTypeName
		}
		composed := vulkanSpecIRDocCompose(refpageName, "", command.XMLDoc, corpus, "")
		doc := vulkanLoaderDocFormatHolderField(command.GoFieldName, refpageName, composed)

		fields = append(fields, gocode.StructFieldTypeDoc(
			command.GoFieldName,
			pfnType,
			doc,
			leading...,
		))
	}

	return gocode.DeclTypeStruct(typeName, fields, typeName+" holds loaded Vulkan commands resolved from the ICD.")
}

func vulkanLoaderCommandsLoadFuncDecl(
	loadName string,
	holderType string,
	tier VulkanSpecIRCommandLoaderTier,
	commands []VulkanSpecIRCommand,
) gocode.FuncDecl {
	params := []gocode.ParamType{
		{Name: "module", Type: gocode.TypeExprPointer(gocode.TypeExprNamed("VulkanModule"))},
	}
	switch tier {
	case VulkanCommandLoaderTierInstance:
		params = append(params,
			gocode.ParamType{Name: "instance", Type: gocode.TypeExprNamed("bindings.VkInstance")},
			gocode.ParamType{Name: "dest", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(holderType))},
		)
	case VulkanCommandLoaderTierDevice:
		params = append(params,
			gocode.ParamType{Name: "device", Type: gocode.TypeExprNamed("bindings.VkDevice")},
			gocode.ParamType{Name: "dest", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(holderType))},
		)
	default:
		params = append(params, gocode.ParamType{Name: "dest", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(holderType))})
	}
	returns := []gocode.TypeExpr{gocode.TypeExprNamed("error")}

	body := vulkanLoaderCommandsLoadBody(holderType, tier, commands)
	return gocode.DeclFunc(loadName, params, returns, body)
}

func vulkanLoaderCommandsLoadBody(
	holderType string,
	tier VulkanSpecIRCommandLoaderTier,
	commands []VulkanSpecIRCommand,
) gocode.BlockStmt {
	holderIdent := gocode.ExprIdent("dest")
	mappingElements := make([]codegen.Expr, 0, len(commands))
	for _, command := range commands {
		mappingElements = append(mappingElements, gocode.ExprCompositeLit(vulkanLoaderCommandMappingType, []gocode.FieldInit{
			{
				Name: "target",
				Value: gocode.ExprAddressOf(
					gocode.ExprSelector(holderIdent, command.GoFieldName),
				),
			},
			{Name: "cName", Value: gocode.ExprCall(
				gocode.ExprSelector(gocode.ExprIdent("syscore"), "SYSCORE_C_StringToCStringFirstByte"),
				gocode.ExprStringLit(command.Name),
			)},
			{Name: "debugName", Value: gocode.ExprStringLit(command.Name)},
		}))
	}

	stmts := []codegen.Stmt{
		gocode.StmtShortVarDecl(
			"mappings",
			gocode.ExprSliceCompositeLit(vulkanLoaderCommandMappingType, mappingElements),
		),
	}

	stmts = append(stmts, vulkanLoaderBindLoopStmts(tier)...)
	stmts = append(stmts, gocode.StmtReturn(gocode.ExprNil()))
	return gocode.StmtBlock(stmts...)
}

func vulkanLoaderBindLoopStmts(tier VulkanSpecIRCommandLoaderTier) []codegen.Stmt {
	moduleIdent := gocode.ExprIdent("module")
	commandIdent := gocode.ExprIdent("command")
	cName := gocode.ExprSelector(commandIdent, "cName")

	var resolveAddr codegen.Expr
	switch tier {
	case VulkanCommandLoaderTierGlobal:
		resolveAddr = gocode.ExprCall(
			gocode.ExprSelector(moduleIdent, "GetInstanceProcAddr"),
			gocode.ExprIntLit(0),
			cName,
		)
	case VulkanCommandLoaderTierInstance:
		resolveAddr = gocode.ExprCall(
			gocode.ExprSelector(moduleIdent, "GetInstanceProcAddr"),
			gocode.ExprIdent("instance"),
			cName,
		)
	case VulkanCommandLoaderTierDevice:
		resolveAddr = gocode.ExprCall(
			gocode.ExprSelector(moduleIdent, "GetDeviceProcAddr"),
			gocode.ExprIdent("device"),
			cName,
		)
	}

	return []codegen.Stmt{
		gocode.StmtRangeLoop("mappings", "command", gocode.StmtBlock(
			gocode.StmtShortVarDecl("addr", resolveAddr),
			gocode.StmtIf(
				gocode.ExprBinary(gocode.BinaryOpEq, gocode.ExprIdent("addr"), gocode.ExprIntLit(0)),
				gocode.StmtBlock(
					gocode.StmtReturn(
						gocode.ExprCall(
							gocode.ExprIdent("fmt.Errorf"),
							gocode.ExprStringLit("vulkan loader: missing entry point %q"),
							gocode.ExprSelector(commandIdent, "debugName"),
						),
					),
				),
			),
			gocode.StmtIfWithInit(
				gocode.StmtShortVarDecl(
					"err",
					gocode.ExprCall(
						gocode.ExprSelector(gocode.ExprIdent("syscore"), "SYSCORE_Pure_FunctionBindAddress"),
						gocode.ExprSelector(commandIdent, "target"),
						gocode.ExprIdent("addr"),
					),
				),
				gocode.ExprBinary(gocode.BinaryOpNe, gocode.ExprIdent("err"), gocode.ExprNil()),
				gocode.StmtBlock(
					gocode.StmtReturn(
						gocode.ExprCall(
							gocode.ExprIdent("fmt.Errorf"),
							gocode.ExprStringLit("vulkan loader: bind %q: %w"),
							gocode.ExprSelector(commandIdent, "debugName"),
							gocode.ExprIdent("err"),
						),
					),
				),
			),
		)),
	}
}

func vulkanLoaderQualifyBindingsType(goType string) string {
	prefix := ""
	remaining := goType
	for strings.HasPrefix(remaining, "*") {
		prefix += "*"
		remaining = remaining[1:]
	}
	for strings.HasPrefix(remaining, "[]") {
		prefix += "[]"
		remaining = remaining[2:]
	}
	if strings.HasPrefix(remaining, "PFN_") || strings.HasPrefix(remaining, "Vk") {
		return prefix + "bindings." + remaining
	}
	return goType
}

func writeLoaderFile(loaderDir string, fileName string, elements []codegen.FileElement) error {
	file := gocode.DeclFile(elements...)
	content, err := gocode.GoFileRenderWithOptions(file, gocode.RenderOptionsEnumBindings())
	if err != nil {
		return fmt.Errorf("render %s: %w", fileName, err)
	}
	outputPath := system.PathJoin(loaderDir, fileName)
	if err := system.FileWriteString(outputPath, content); err != nil {
		return fmt.Errorf("write %s to %s: %w", fileName, outputPath, err)
	}
	return nil
}

func vulkanLoaderFormat(loaderDir string) error {
	cmd := exec.Command("gofmt", "-w", loaderDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt %s: %w\n%s", loaderDir, err, output)
	}
	return nil
}
