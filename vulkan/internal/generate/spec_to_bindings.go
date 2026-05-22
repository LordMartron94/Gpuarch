package main

import (
	"codegen"
	gocode "codegen/go"
	"fmt"
	"foundation/system"
	"time"
)

const (
	enumBindingFile             = "bindings_enums_gen.go"
	vulkanBindingsGeneratorTool = "gpuarch Vulkan bindings generator"
	vulkanEnumBaseTypeName      = "VkEnum"
)

func specToBindingsConvert(ir VulkanSpecIR, bindingsDir string) error {
	if err := generateEnumContent(ir.Enums, bindingsDir); err != nil {
		return err
	}

	return nil
}

func generateEnumContent(enums []VulkanSpecIREnum, bindingsDir string) error {
	elements := make([]codegen.FileElement, 0, len(enums)*3+8)
	elements = append(elements, gocode.GoGeneratedFileHeader(vulkanBindingsGeneratorTool, time.Now())...)
	elements = append(elements, gocode.FileElementFrom(gocode.DeclPackage("bindings")))
	blankLine(&elements)
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

	file := gocode.DeclFile(elements...)

	content, err := gocode.GoFileRenderWithOptions(file, gocode.RenderOptionsEnumBindings())
	if err != nil {
		return fmt.Errorf("render enum bindings: %w", err)
	}

	outputPath := system.PathJoin(bindingsDir, enumBindingFile)

	if err := system.FileWriteString(outputPath, content); err != nil {
		return fmt.Errorf("write enum bindings to %s: %w", outputPath, err)
	}

	return nil
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

func blankLine(elements *[]codegen.FileElement) {
	*elements = append(*elements, gocode.GoBlankLine())
}
