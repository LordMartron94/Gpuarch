package main

import (
	"flag"
	"fmt"
	"foundation/system"
	"path/filepath"
)

const specName string = "vulkan_spec.xml"
const specDebug string = "vulkan_spec_debug.txt"

func main() {
	var specOutputDir string
	var bindingOutputDir string
	flag.StringVar(&specOutputDir, "specOutputDir", "", "directory for vulkan_spec.xml (fetched when missing or when bindings are not requested)")
	flag.StringVar(&bindingOutputDir, "bindingOutputDir", "", "directory for generated Vulkan bindings (reuses existing spec when present)")
	flag.Parse()

	if specOutputDir == "" && bindingOutputDir == "" {
		panic("gpuarch generate: provide -specOutputDir and/or -bindingOutputDir")
	}

	if bindingOutputDir != "" {
		if specOutputDir == "" {
			panic("gpuarch generate: -bindingOutputDir requires -specOutputDir")
		}
		runBindingGeneration(specOutputDir, bindingOutputDir)
		return
	}

	runSpecFetch(specOutputDir)
}

func runSpecFetch(specOutputDir string) {
	paths, err := vulkanRegistrySpecPathsResolve(specOutputDir)
	if err != nil {
		panic(err)
	}

	system.DirCreate(filepath.Dir(paths.SpecFile), true)

	if _, err := vulkanRegistrySpecFetchWriteAndDebug(paths); err != nil {
		panic(err)
	}

	fmt.Printf("Fetched Vulkan registry spec to %s\n", paths.SpecFile)
	fmt.Printf("Wrote Vulkan registry debug tree to %s\n", paths.DebugFile)
}

func runBindingGeneration(specOutputDir string, bindingOutputDir string) {
	_, err := filepath.Abs(bindingOutputDir)
	if err != nil {
		panic(fmt.Errorf("resolve binding output path: %w", err))
	}

	root, loadedExisting, err := vulkanRegistrySpecEnsure(specOutputDir)
	if err != nil {
		panic(err)
	}

	paths, err := vulkanRegistrySpecPathsResolve(specOutputDir)
	if err != nil {
		panic(err)
	}

	if loadedExisting {
		fmt.Printf("Using existing Vulkan registry spec at %s\n", paths.SpecFile)
	} else {
		fmt.Printf("Fetched Vulkan registry spec to %s\n", paths.SpecFile)
		fmt.Printf("Wrote Vulkan registry debug tree to %s\n", paths.DebugFile)
	}

	vulkanBindingEnumTypeNamesPrint(root)
}
