package main

import (
	"flag"
	"fmt"
	"foundation/system"
	"path/filepath"
)

func bindingOutputAbsPath(bindingOutputDir string) string {
	bindingOutputAbs, err := filepath.Abs(bindingOutputDir)
	if err != nil {
		panic(fmt.Errorf("resolve binding output path: %w", err))
	}
	return bindingOutputAbs
}

const specName string = "vulkan_spec.xml"
const specDebug string = "vulkan_spec_debug.txt"

func main() {
	var specOutputDir string
	var bindingOutputDir string
	var refpageDir string
	var refpageSkipFetch bool
	var refpageUpdate bool
	flag.StringVar(&specOutputDir, "specOutputDir", "", "directory for vulkan_spec.xml (fetched when missing or when bindings are not requested)")
	flag.StringVar(&bindingOutputDir, "bindingOutputDir", "", "directory for generated Vulkan bindings (reuses existing spec when present)")
	flag.StringVar(&refpageDir, "refpageDir", "", "directory for cached Khronos man page HTML (defaults to <specOutputDir>/refpages)")
	flag.BoolVar(&refpageSkipFetch, "refpageSkipFetch", false, "never fetch man pages; use refpage cache only")
	flag.BoolVar(&refpageUpdate, "refpageUpdate", false, "re-fetch cached man pages when the registry reports newer content")
	flag.Parse()

	if specOutputDir == "" && bindingOutputDir == "" {
		panic("gpuarch generate: provide -specOutputDir and/or -bindingOutputDir")
	}

	if bindingOutputDir != "" {
		if specOutputDir == "" {
			panic("gpuarch generate: -bindingOutputDir requires -specOutputDir")
		}
		runBindingGeneration(specOutputDir, bindingOutputDir, refpageDir, refpageSkipFetch, refpageUpdate)
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

func runBindingGeneration(specOutputDir string, bindingOutputDir string, refpageDir string, refpageSkipFetch bool, refpageUpdate bool) {
	bindingOutputAbs := bindingOutputAbsPath(bindingOutputDir)
	system.DirCreate(bindingOutputAbs, true)

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

	ir := vulkanSpecIRBuild(root)

	if refpageDir == "" {
		refpageDir = system.PathJoin(specOutputDir, "refpages")
	}

	refpageNames := vulkanRefpageCorpusNamesFromIR(ir)
	corpus, err := vulkanRefpageCorpusEnsure(refpageDir, paths.SpecFile, refpageNames, vulkanRefpageFetchOptions{
		SkipNetwork: refpageSkipFetch,
		UpdateStale: refpageUpdate,
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d Vulkan man pages from %s\n", len(corpus), refpageDir)

	if err := specToBindingsConvert(ir, bindingOutputDir, corpus); err != nil {
		panic(err)
	}

	loaderOutputDir := system.PathJoin(system.PathDir(bindingOutputAbs), "loader")
	system.DirCreate(loaderOutputDir, true)
	if err := specToLoaderConvert(ir, loaderOutputDir, corpus); err != nil {
		panic(err)
	}

	fmt.Printf("Wrote Vulkan bindings to %s\n", bindingOutputDir)
	fmt.Printf("Wrote Vulkan loader to %s\n", loaderOutputDir)
	fmt.Printf("Formatted Vulkan bindings and loader with gofmt\n")
}
