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

	vulkanLoaderCommandsType       = "VulkanCommands"
	vulkanLoaderCommandMappingType = "vulkanCommandMapping"
)

var vulkanLoaderEndpointOrder = []VulkanSpecIRCommandEndpoint{
	VulkanCommandEndpointGlobal,
	VulkanCommandEndpointInstance,
	VulkanCommandEndpointPhysicalDevice,
	VulkanCommandEndpointSurface,
	VulkanCommandEndpointSwapchain,
	VulkanCommandEndpointDevice,
	VulkanCommandEndpointQueue,
	VulkanCommandEndpointCommandBuffer,
}

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

	byEndpoint := map[VulkanSpecIRCommandEndpoint][]VulkanSpecIRCommand{}
	for _, command := range commands {
		if command.GoFieldName == "" || command.PFNTypeName == "" {
			continue
		}
		endpoint := command.CommandEndpoint
		if endpoint == "" {
			endpoint = VulkanCommandEndpointInstance
		}
		byEndpoint[endpoint] = append(byEndpoint[endpoint], command)
	}

	elements := loaderFilePreamble()
	blankLine(&elements)
	elements = append(elements, gocode.FileElementFrom(vulkanLoaderCommandMappingTypeDecl()))
	blankLine(&elements)

	for _, endpoint := range vulkanLoaderEndpointOrder {
		endpointCommands := byEndpoint[endpoint]
		if len(endpointCommands) == 0 {
			continue
		}
		sort.Slice(endpointCommands, func(i, j int) bool {
			return endpointCommands[i].GoFieldName < endpointCommands[j].GoFieldName
		})
		typeName := vulkanLoaderEndpointTypeName(endpoint)
		section := vulkanLoaderEndpointSectionLabel(endpoint)
		elements = append(elements, gocode.FileElementFrom(
			vulkanLoaderCommandsStructDecl(typeName, section, endpointCommands, corpus),
		))
		blankLine(&elements)
	}

	elements = append(elements, gocode.FileElementFrom(vulkanLoaderCommandsRootStructDecl(byEndpoint)))
	blankLine(&elements)

	tierSpecs := []struct {
		tier     VulkanSpecIRCommandLoaderTier
		loadName string
	}{
		{VulkanCommandLoaderTierGlobal, "VulkanCommandsLoadGlobal"},
		{VulkanCommandLoaderTierInstance, "VulkanCommandsLoadInstance"},
		{VulkanCommandLoaderTierDevice, "VulkanCommandsLoadDevice"},
	}

	for _, spec := range tierSpecs {
		tierCommands := vulkanLoaderCommandsForTier(commands, spec.tier)
		if len(tierCommands) == 0 {
			continue
		}
		sort.Slice(tierCommands, func(i, j int) bool {
			return tierCommands[i].GoFieldName < tierCommands[j].GoFieldName
		})
		elements = append(elements, gocode.FileElementFrom(
			vulkanLoaderCommandsTierLoadFuncDecl(spec.loadName, spec.tier, tierCommands),
		))
		blankLine(&elements)
	}

	return writeLoaderFile(loaderDir, loaderCommandsFile, elements)
}

func vulkanLoaderCommandsForTier(commands []VulkanSpecIRCommand, tier VulkanSpecIRCommandLoaderTier) []VulkanSpecIRCommand {
	filtered := make([]VulkanSpecIRCommand, 0, len(commands))
	for _, command := range commands {
		if command.GoFieldName == "" || command.PFNTypeName == "" {
			continue
		}
		if command.LoaderTier == tier {
			filtered = append(filtered, command)
		}
	}
	return filtered
}

func vulkanLoaderEndpointTypeName(endpoint VulkanSpecIRCommandEndpoint) string {
	return vulkanLoaderCommandsType + vulkanLoaderEndpointGoName(endpoint) + "Endpoint"
}

func vulkanLoaderEndpointGoName(endpoint VulkanSpecIRCommandEndpoint) string {
	switch endpoint {
	case VulkanCommandEndpointGlobal:
		return "Global"
	case VulkanCommandEndpointInstance:
		return "Instance"
	case VulkanCommandEndpointPhysicalDevice:
		return "PhysicalDevice"
	case VulkanCommandEndpointSurface:
		return "Surface"
	case VulkanCommandEndpointSwapchain:
		return "Swapchain"
	case VulkanCommandEndpointDevice:
		return "Device"
	case VulkanCommandEndpointQueue:
		return "Queue"
	case VulkanCommandEndpointCommandBuffer:
		return "CommandBuffer"
	default:
		return "Instance"
	}
}

func vulkanLoaderEndpointSectionLabel(endpoint VulkanSpecIRCommandEndpoint) string {
	switch endpoint {
	case VulkanCommandEndpointGlobal:
		return "Global commands"
	case VulkanCommandEndpointInstance:
		return "Instance commands"
	case VulkanCommandEndpointPhysicalDevice:
		return "Physical device commands"
	case VulkanCommandEndpointSurface:
		return "Surface commands"
	case VulkanCommandEndpointSwapchain:
		return "Swapchain commands"
	case VulkanCommandEndpointDevice:
		return "Device commands"
	case VulkanCommandEndpointQueue:
		return "Queue commands"
	case VulkanCommandEndpointCommandBuffer:
		return "Command buffer commands"
	default:
		return "Commands"
	}
}

func vulkanLoaderCommandsRootStructDecl(byEndpoint map[VulkanSpecIRCommandEndpoint][]VulkanSpecIRCommand) gocode.TypeStructDecl {
	fields := make([]gocode.StructFieldDecl, 0, len(vulkanLoaderEndpointOrder))
	for endpointIndex, endpoint := range vulkanLoaderEndpointOrder {
		if len(byEndpoint[endpoint]) == 0 {
			continue
		}
		goName := vulkanLoaderEndpointGoName(endpoint)
		typeName := vulkanLoaderEndpointTypeName(endpoint)
		leading := []codegen.Node(nil)
		if endpointIndex == 0 {
			leading = []codegen.Node{gocode.LayoutBlankLineNode()}
		}
		fields = append(fields, gocode.StructFieldTypeDoc(
			goName,
			gocode.TypeExprNamed(typeName),
			goName+" groups Vulkan commands whose primary handle is "+vulkanLoaderEndpointSectionLabel(endpoint)+".",
			leading...,
		))
	}
	return gocode.DeclTypeStruct(
		vulkanLoaderCommandsType,
		fields,
		"VulkanCommands holds loaded Vulkan entry points from the ICD, grouped by semantic endpoint.",
	)
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

func vulkanLoaderCommandsFieldSelector(holderIdent codegen.Expr, command VulkanSpecIRCommand) codegen.Expr {
	endpointName := vulkanLoaderEndpointGoName(command.CommandEndpoint)
	return gocode.ExprSelector(
		gocode.ExprSelector(holderIdent, endpointName),
		command.GoFieldName,
	)
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

func vulkanLoaderCommandsTierLoadFuncDecl(
	loadName string,
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
			gocode.ParamType{Name: "commands", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(vulkanLoaderCommandsType))},
		)
	case VulkanCommandLoaderTierDevice:
		params = append(params,
			gocode.ParamType{Name: "device", Type: gocode.TypeExprNamed("bindings.VkDevice")},
			gocode.ParamType{Name: "commands", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(vulkanLoaderCommandsType))},
		)
	default:
		params = append(params, gocode.ParamType{Name: "commands", Type: gocode.TypeExprPointer(gocode.TypeExprNamed(vulkanLoaderCommandsType))})
	}
	returns := []gocode.TypeExpr{gocode.TypeExprNamed("error")}

	body := vulkanLoaderCommandsLoadBody(tier, commands)
	return gocode.DeclFunc(loadName, params, returns, body)
}

func vulkanLoaderCommandsLoadBody(
	tier VulkanSpecIRCommandLoaderTier,
	commands []VulkanSpecIRCommand,
) gocode.BlockStmt {
	holderIdent := gocode.ExprIdent("commands")
	mappingElements := make([]codegen.Expr, 0, len(commands))
	for _, command := range commands {
		mappingElements = append(mappingElements, gocode.ExprCompositeLit(vulkanLoaderCommandMappingType, []gocode.FieldInit{
			{
				Name:  "target",
				Value: gocode.ExprAddressOf(vulkanLoaderCommandsFieldSelector(holderIdent, command)),
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

func vulkanLoaderMissingEntryPointReturnStmt(tier VulkanSpecIRCommandLoaderTier, commandIdent codegen.Expr) []codegen.Stmt {
	var instanceArg codegen.Expr
	var deviceArg codegen.Expr
	switch tier {
	case VulkanCommandLoaderTierInstance:
		instanceArg = gocode.ExprIdent("instance")
		deviceArg = gocode.ExprIntLit(0)
	case VulkanCommandLoaderTierDevice:
		instanceArg = gocode.ExprIntLit(0)
		deviceArg = gocode.ExprIdent("device")
	default:
		instanceArg = gocode.ExprIntLit(0)
		deviceArg = gocode.ExprIntLit(0)
	}

	return []codegen.Stmt{
		gocode.StmtReturn(
			gocode.ExprCall(
				gocode.ExprIdent("vulkanLoaderEntryPointResolveError"),
				gocode.ExprSelector(commandIdent, "debugName"),
				vulkanLoaderBindTierIntLit(tier),
				instanceArg,
				deviceArg,
			),
		),
	}
}

func vulkanLoaderBindTierIntLit(tier VulkanSpecIRCommandLoaderTier) *gocode.IntLitExpr {
	switch tier {
	case VulkanCommandLoaderTierGlobal:
		return gocode.ExprIntLit(vulkanCommandLoaderTierGlobal)
	case VulkanCommandLoaderTierDevice:
		return gocode.ExprIntLit(vulkanCommandLoaderTierDevice)
	default:
		return gocode.ExprIntLit(vulkanCommandLoaderTierInstance)
	}
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
				gocode.StmtBlock(vulkanLoaderMissingEntryPointReturnStmt(tier, commandIdent)...),
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
