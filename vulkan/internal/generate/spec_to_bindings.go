package main

import (
	"codegen"
	gocode "codegen/go"
	"fmt"
	"foundation/system"
	"os/exec"
	"time"
)

const (
	basetypeBindingFile         = "bindings_basetypes_gen.go"
	constantsBindingFile        = "bindings_constants_gen.go"
	enumBindingFile             = "bindings_enums_gen.go"
	flagsBindingFile            = "bindings_flags_gen.go"
	funcpointerBindingFile      = "bindings_funcpointers_gen.go"
	handlesBindingFile          = "bindings_handles_gen.go"
	structsBindingFile          = "bindings_structs_gen.go"
	vulkanBindingsGeneratorTool = "gpuarch Vulkan bindings generator"
	vulkanEnumBaseTypeName      = "VkEnum"
)

func specToBindingsConvert(ir VulkanSpecIR, bindingsDir string, corpus VulkanRefpageCorpus) error {
	if err := generateBasetypeContent(ir.Basetypes, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateConstantsContent(ir.Constants, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateEnumContent(ir.Enums, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateFlagsContent(ir.Flags, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateFuncpointerContent(ir.Funcpointers, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateHandlesContent(ir.Handles, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateStructsContent(ir.Structs, bindingsDir, corpus); err != nil {
		return err
	}
	if err := generateRegistryContent(ir, bindingsDir, corpus); err != nil {
		return err
	}
	return vulkanBindingsFormat(bindingsDir)
}

func generateStructsContent(structs []VulkanSpecIRStruct, bindingsDir string, corpus VulkanRefpageCorpus) error {
	if len(structs) == 0 {
		return nil
	}

	needsUnsafe := false
	for _, aggregate := range structs {
		for _, field := range aggregate.Fields {
			if field.NeedsUnsafe {
				needsUnsafe = true
				break
			}
		}
	}

	elements := bindingFilePreamble()
	if needsUnsafe {
		elements = append(elements, gocode.FileElementFrom(gocode.DeclImportBlock("unsafe")))
		blankLine(&elements)
	}

	for _, aggregate := range structs {
		elements = append(elements, vulkanSpecIRStructBindingElements(aggregate, corpus)...)
	}

	return writeBindingFile(bindingsDir, structsBindingFile, elements)
}

func vulkanSpecIRStructBindingElements(aggregate VulkanSpecIRStruct, corpus VulkanRefpageCorpus) []codegen.FileElement {
	if aggregate.Name == "" {
		return nil
	}

	doc := vulkanSpecIRDocCompose(aggregate.Name, aggregate.AliasOf, aggregate.XMLDoc, corpus, "")

	if aggregate.AliasOf != "" {
		elements := make([]codegen.FileElement, 0, 2)
		elements = append(elements, gocode.FileElementFrom(
			gocode.DeclTypeDefined(
				aggregate.Name,
				gocode.TypeExprNamed(aggregate.AliasOf),
				true,
				vulkanSpecIRDocFormatExported(aggregate.Name, doc),
			),
		))
		blankLine(&elements)
		return elements
	}

	fields := make([]gocode.StructFieldDecl, 0, len(aggregate.Fields))
	for _, field := range aggregate.Fields {
		if field.Name == "" || field.GoType == "" {
			continue
		}
		typ, err := gocode.TypeExprFromGoTypeString(field.GoType)
		if err != nil {
			continue
		}
		fieldDoc := vulkanSpecIRDocCompose(
			aggregate.Name,
			aggregate.AliasOf,
			field.XMLDoc,
			corpus,
			field.VulkanMemberName,
		)
		formattedDoc := vulkanSpecIRDocFormatStructField(field.Name, fieldDoc)
		fields = append(fields, gocode.StructFieldTypeDoc(
			field.Name,
			typ,
			formattedDoc,
		))
	}

	elements := make([]codegen.FileElement, 0, 2)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeStruct(aggregate.Name, fields, vulkanSpecIRDocFormatExported(aggregate.Name, doc)),
	))
	blankLine(&elements)
	return elements
}

/*
vulkanBindingsFormat runs gofmt on all generated Go files in bindingsDir.
*/
func vulkanBindingsFormat(bindingsDir string) error {
	cmd := exec.Command("gofmt", "-w", bindingsDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt %s: %w\n%s", bindingsDir, err, output)
	}
	return nil
}

func generateFuncpointerContent(funcpointers []VulkanSpecIRFuncpointer, bindingsDir string, corpus VulkanRefpageCorpus) error {
	if len(funcpointers) == 0 {
		return nil
	}

	needsUnsafe := false
	for _, fn := range funcpointers {
		if vulkanSpecIRFuncpointerNeedsUnsafe(fn) {
			needsUnsafe = true
			break
		}
	}

	elements := bindingFilePreamble()
	if needsUnsafe {
		elements = append(elements, gocode.FileElementFrom(gocode.DeclImportBlock("unsafe")))
		blankLine(&elements)
	}

	for _, fn := range funcpointers {
		elements = append(elements, vulkanSpecIRFuncpointerBindingElements(fn, corpus)...)
	}

	return writeBindingFile(bindingsDir, funcpointerBindingFile, elements)
}

func vulkanSpecIRFuncpointerBindingElements(fn VulkanSpecIRFuncpointer, corpus VulkanRefpageCorpus) []codegen.FileElement {
	if fn.Name == "" {
		return nil
	}

	sig, err := vulkanSpecIRFuncpointerGoTypeExpr(fn)
	if err != nil {
		return nil
	}

	doc := vulkanSpecIRDocCompose(fn.Name, "", fn.XMLDoc, corpus, "")
	elements := make([]codegen.FileElement, 0, 2)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			fn.Name,
			sig,
			false,
			vulkanSpecIRDocFormatExported(fn.Name, doc),
		),
	))
	blankLine(&elements)
	return elements
}

func generateHandlesContent(handles []VulkanSpecIRHandle, bindingsDir string, corpus VulkanRefpageCorpus) error {
	if len(handles) == 0 {
		return nil
	}

	elements := bindingFilePreamble()
	for _, handle := range handles {
		elements = append(elements, vulkanSpecIRHandleBindingElements(handle, corpus)...)
	}
	return writeBindingFile(bindingsDir, handlesBindingFile, elements)
}

func vulkanSpecIRHandleBindingElements(handle VulkanSpecIRHandle, corpus VulkanRefpageCorpus) []codegen.FileElement {
	if handle.Name == "" || handle.Underlying == "" {
		return nil
	}

	underlying := gocode.TypeExprNamed(handle.Underlying)
	if handle.AliasOf != "" {
		underlying = gocode.TypeExprNamed(handle.AliasOf)
	}

	doc := vulkanSpecIRDocCompose(handle.Name, handle.AliasOf, handle.XMLDoc, corpus, "")

	elements := make([]codegen.FileElement, 0, 2)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			handle.Name,
			underlying,
			handle.AliasOf != "",
			vulkanSpecIRDocFormatExported(handle.Name, doc),
		),
	))
	blankLine(&elements)
	return elements
}

func writeBindingFile(bindingsDir string, fileName string, elements []codegen.FileElement) error {
	file := gocode.DeclFile(elements...)

	content, err := gocode.GoFileRenderWithOptions(file, gocode.RenderOptionsEnumBindings())
	if err != nil {
		return fmt.Errorf("render %s: %w", fileName, err)
	}

	outputPath := system.PathJoin(bindingsDir, fileName)
	if err := system.FileWriteString(outputPath, content); err != nil {
		return fmt.Errorf("write %s to %s: %w", fileName, outputPath, err)
	}

	return nil
}

func bindingFilePreamble() []codegen.FileElement {
	elements := make([]codegen.FileElement, 0, 4)
	elements = append(elements, gocode.GoGeneratedFileHeader(vulkanBindingsGeneratorTool, time.Now())...)
	elements = append(elements, gocode.FileElementFrom(gocode.DeclPackage("bindings")))
	blankLine(&elements)
	return elements
}

func generateBasetypeContent(basetypes []VulkanSpecIRBasetype, bindingsDir string, corpus VulkanRefpageCorpus) error {
	elements := bindingFilePreamble()

	for _, basetype := range basetypes {
		doc := vulkanSpecIRDocCompose(basetype.Name, "", basetype.XMLDoc, corpus, "")
		elements = append(elements, gocode.FileElementFrom(
			gocode.DeclTypeDefined(
				basetype.Name,
				gocode.TypeExprNamed(basetype.Underlying),
				true,
				vulkanSpecIRDocFormatExported(basetype.Name, doc),
			),
		))
	}

	return writeBindingFile(bindingsDir, basetypeBindingFile, elements)
}

func generateConstantsContent(constants VulkanSpecIRConstants, bindingsDir string, corpus VulkanRefpageCorpus) error {
	if len(constants.Values) == 0 {
		return nil
	}

	elements := bindingFilePreamble()
	elements = append(elements, vulkanSpecIRConstantsBindingElements(constants, corpus)...)
	return writeBindingFile(bindingsDir, constantsBindingFile, elements)
}

func vulkanSpecIRConstantsBindingElements(constants VulkanSpecIRConstants, corpus VulkanRefpageCorpus) []codegen.FileElement {
	specs := make([]gocode.ConstSpec, 0, len(constants.Values))
	for _, value := range constants.Values {
		if value.Key == "" || value.Value == "" {
			continue
		}

		var typ *gocode.TypeExpr
		if value.GoType != "" {
			typ = gocode.TypeExprNamedPtr(value.GoType)
		}

		doc := vulkanSpecIRDocCompose(value.Key, "", value.XMLDoc, corpus, "")
		specs = append(specs, gocode.ConstSpecNew(
			value.Key,
			typ,
			value.Value,
			vulkanSpecIRDocFormatExported(value.Key, doc),
		))
	}

	if len(specs) == 0 {
		return nil
	}

	groupDoc := ""
	if constants.Name != "" {
		groupDoc = vulkanSpecIRDocCompose(constants.Name, "", constants.XMLDoc, corpus, "")
	}

	elements := make([]codegen.FileElement, 0, 2)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclConstGroup(specs, groupDoc, true),
	))
	blankLine(&elements)
	return elements
}

func generateEnumContent(enums []VulkanSpecIREnum, bindingsDir string, corpus VulkanRefpageCorpus) error {
	elements := bindingFilePreamble()
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			vulkanEnumBaseTypeName,
			gocode.TypeExprNamed("int32"),
			true,
			vulkanSpecIRDocFormatExported(vulkanEnumBaseTypeName, "is the base type for Vulkan enumerated types in the registry."),
		),
	))
	blankLine(&elements)

	for _, enum := range enums {
		elements = append(elements, vulkanSpecIREnumBindingElements(enum, corpus)...)
	}

	return writeBindingFile(bindingsDir, enumBindingFile, elements)
}

func generateFlagsContent(flags []VulkanSpecIRFlags, bindingsDir string, corpus VulkanRefpageCorpus) error {
	elements := bindingFilePreamble()

	for _, flagType := range flags {
		elements = append(elements, vulkanSpecIRFlagsBindingElements(flagType, corpus)...)
	}

	return writeBindingFile(bindingsDir, flagsBindingFile, elements)
}

func vulkanSpecIREnumBindingElements(enum VulkanSpecIREnum, corpus VulkanRefpageCorpus) []codegen.FileElement {
	if enum.Name == "" || len(enum.Values) == 0 {
		return nil
	}

	doc := vulkanSpecIRDocCompose(enum.Name, "", enum.XMLDoc, corpus, "")
	elements := make([]codegen.FileElement, 0, 3)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			enum.Name,
			gocode.TypeExprNamed(vulkanEnumBaseTypeName),
			false,
			vulkanSpecIRDocFormatExported(enum.Name, doc),
		),
	))

	specs := make([]gocode.ConstSpec, 0, len(enum.Values))
	enumType := gocode.TypeExprNamedPtr(enum.Name)
	for _, value := range enum.Values {
		if value.Key == "" || value.Value == "" {
			continue
		}
		valueDoc := vulkanSpecIRDocCompose(value.Key, "", value.XMLDoc, corpus, "")
		specs = append(specs, gocode.ConstSpecNew(
			value.Key,
			enumType,
			value.Value,
			vulkanSpecIRDocFormatExported(value.Key, valueDoc),
		))
	}

	if len(specs) > 0 {
		elements = append(elements, gocode.FileElementFrom(
			gocode.DeclConstGroup(specs, "", true),
		))
	}

	blankLine(&elements)
	return elements
}

func vulkanSpecIRFlagsBindingElements(flagType VulkanSpecIRFlags, corpus VulkanRefpageCorpus) []codegen.FileElement {
	if flagType.AggregateName == "" {
		return nil
	}

	doc := vulkanSpecIRDocCompose(flagType.AggregateName, "", flagType.XMLDoc, corpus, "")
	elements := make([]codegen.FileElement, 0, 3)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			flagType.AggregateName,
			gocode.TypeExprNamed(flagType.BaseTypeName),
			false,
			vulkanSpecIRDocFormatExported(flagType.AggregateName, doc),
		),
	))

	if len(flagType.Values) > 0 {
		specs := make([]gocode.ConstSpec, 0, len(flagType.Values))
		flagsType := gocode.TypeExprNamedPtr(flagType.AggregateName)
		for _, value := range flagType.Values {
			valueDoc := vulkanSpecIRDocCompose(value.Key, "", value.XMLDoc, corpus, "")
			specs = append(specs, gocode.ConstSpecNew(
				value.Key,
				flagsType,
				value.Value,
				vulkanSpecIRDocFormatExported(value.Key, valueDoc),
			))
		}
		elements = append(elements, gocode.FileElementFrom(
			gocode.DeclConstGroup(specs, "", true),
		))
	}

	blankLine(&elements)
	return elements
}

func blankLine(elements *[]codegen.FileElement) {
	*elements = append(*elements, gocode.GoBlankLine())
}
