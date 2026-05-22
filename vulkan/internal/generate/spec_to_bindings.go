package main

import (
	"codegen"
	gocode "codegen/go"
	"fmt"
	"foundation/system"
	"time"
)

const (
	basetypeBindingFile         = "bindings_basetypes_gen.go"
	enumBindingFile             = "bindings_enums_gen.go"
	flagsBindingFile            = "bindings_flags_gen.go"
	vulkanBindingsGeneratorTool = "gpuarch Vulkan bindings generator"
	vulkanEnumBaseTypeName      = "VkEnum"
)

func specToBindingsConvert(ir VulkanSpecIR, bindingsDir string) error {
	if err := generateBasetypeContent(ir.Basetypes, bindingsDir); err != nil {
		return err
	}
	if err := generateEnumContent(ir.Enums, bindingsDir); err != nil {
		return err
	}
	if err := generateFlagsContent(ir.Flags, bindingsDir); err != nil {
		return err
	}
	return nil
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

func generateBasetypeContent(basetypes []VulkanSpecIRBasetype, bindingsDir string) error {
	elements := bindingFilePreamble()

	for _, basetype := range basetypes {
		elements = append(elements, gocode.FileElementFrom(
			gocode.DeclTypeDefined(
				basetype.Name,
				gocode.TypeExprNamed(basetype.Underlying),
				true,
				gocode.GoDocFormatExported(basetype.Name, basetype.Doc),
			),
		))
	}

	return writeBindingFile(bindingsDir, basetypeBindingFile, elements)
}

func generateEnumContent(enums []VulkanSpecIREnum, bindingsDir string) error {
	elements := bindingFilePreamble()
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			vulkanEnumBaseTypeName,
			gocode.TypeExprNamed("int32"),
			true,
			gocode.GoDocFormatExported(vulkanEnumBaseTypeName, "is the base type for Vulkan enumerated types in the registry."),
		),
	))
	blankLine(&elements)

	for _, enum := range enums {
		elements = append(elements, vulkanSpecIREnumBindingElements(enum)...)
	}

	return writeBindingFile(bindingsDir, enumBindingFile, elements)
}

func generateFlagsContent(flags []VulkanSpecIRFlags, bindingsDir string) error {
	elements := bindingFilePreamble()

	for _, flagType := range flags {
		elements = append(elements, vulkanSpecIRFlagsBindingElements(flagType)...)
	}

	return writeBindingFile(bindingsDir, flagsBindingFile, elements)
}

func vulkanSpecIREnumBindingElements(enum VulkanSpecIREnum) []codegen.FileElement {
	if enum.Name == "" || len(enum.Values) == 0 {
		return nil
	}

	elements := make([]codegen.FileElement, 0, 3)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			enum.Name,
			gocode.TypeExprNamed(vulkanEnumBaseTypeName),
			false,
			gocode.GoDocFormatExported(enum.Name, enum.Doc),
		),
	))

	specs := make([]gocode.ConstSpec, 0, len(enum.Values))
	enumType := gocode.TypeExprNamedPtr(enum.Name)
	for _, value := range enum.Values {
		if value.Key == "" || value.Value == "" {
			continue
		}
		specs = append(specs, gocode.ConstSpecNew(
			value.Key,
			enumType,
			value.Value,
			gocode.GoDocFormatExported(value.Key, value.Doc),
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

func vulkanSpecIRFlagsBindingElements(flagType VulkanSpecIRFlags) []codegen.FileElement {
	if flagType.Name == "" || flagType.AggregateName == "" {
		return nil
	}

	elements := make([]codegen.FileElement, 0, 3)
	elements = append(elements, gocode.FileElementFrom(
		gocode.DeclTypeDefined(
			flagType.AggregateName,
			gocode.TypeExprNamed(flagType.BaseTypeName),
			false,
			gocode.GoDocFormatExported(flagType.AggregateName, flagType.Doc),
		),
	))

	if len(flagType.Values) > 0 {
		specs := make([]gocode.ConstSpec, 0, len(flagType.Values))
		flagsType := gocode.TypeExprNamedPtr(flagType.AggregateName)
		for _, value := range flagType.Values {
			specs = append(specs, gocode.ConstSpecNew(
				value.Key,
				flagsType,
				value.Value,
				gocode.GoDocFormatExported(value.Key, value.Doc),
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
