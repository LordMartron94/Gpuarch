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
	flag.StringVar(&specOutputDir, "specOutputDir", "", "download vk.xml from the Khronos registry and write it to this file path")
	flag.StringVar(&bindingOutputDir, "bindingOutputDir", "", "render generated Vulkan command bindings to this Go file path")
	flag.Parse()

	if specOutputDir != "" {
		if bindingOutputDir != "" {
			fmt.Println("skipping binding output generation for now as it is not yet implemented")
			return
		}

		specOutputDirAbs, err := filepath.Abs(specOutputDir)
		if err != nil {
			panic(fmt.Errorf("resolve spec output path: %w", err))
		}

		system.DirCreate(specOutputDirAbs, true)

		specOutputFile := system.PathJoin(specOutputDirAbs, specName)
		specDebugOutputFile := system.PathJoin(specOutputDirAbs, specDebug)

		if err := vulkanRegistrySpecWriteToFile(specOutputFile); err != nil {
			panic(err)
		}

		if err := vulkanRegistrySpecWriteDebugTree(specOutputFile, specDebugOutputFile); err != nil {
			panic(err)
		}

		fmt.Printf("Fetched Vulkan registry spec to %s\n", specOutputFile)
		fmt.Printf("Wrote Vulkan registry debug tree to %s\n", specDebugOutputFile)
		return
	}

	panic("vulkangen: provide -specOutputDir, or -bindingOutputDir")
}
