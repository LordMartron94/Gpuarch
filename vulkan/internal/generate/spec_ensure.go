package main

import (
	"fmt"
	"foundation/system"
	"path/filepath"
)

/*
vulkanRegistrySpecPaths holds resolved spec and debug artifact paths under a spec output directory.
*/
type vulkanRegistrySpecPaths struct {
	SpecFile  string
	DebugFile string
}

/*
vulkanRegistrySpecPathsResolve maps specOutputDir to the canonical spec XML and debug tree file paths.

[Parameters]
specOutputDir is the directory that holds vulkan_spec.xml.

[Returns]
Resolved absolute paths and nil error on success.
*/
func vulkanRegistrySpecPathsResolve(specOutputDir string) (vulkanRegistrySpecPaths, error) {
	specOutputDirAbs, err := filepath.Abs(specOutputDir)
	if err != nil {
		return vulkanRegistrySpecPaths{}, fmt.Errorf("resolve spec output path: %w", err)
	}

	return vulkanRegistrySpecPaths{
		SpecFile:  system.PathJoin(specOutputDirAbs, specName),
		DebugFile: system.PathJoin(specOutputDirAbs, specDebug),
	}, nil
}

/*
vulkanRegistrySpecFetchWriteAndDebug downloads the registry XML, writes the spec file, and writes the debug tree dump.

[Parameters]
paths must contain writable SpecFile and DebugFile locations. Parent directories must already exist.

[Returns]
The parsed registry tree and nil error on success.

[Side Effects]
Performs a network GET. Overwrites SpecFile and DebugFile when they already exist.
*/
func vulkanRegistrySpecFetchWriteAndDebug(paths vulkanRegistrySpecPaths) (*xmlSpecNode, error) {
	content, err := vulkanRegistrySpecFetch(VulkanRegistrySpecURL)
	if err != nil {
		return nil, err
	}

	if err := system.FileWriteBytes(paths.SpecFile, content); err != nil {
		return nil, fmt.Errorf("write Vulkan registry spec to %s: %w", paths.SpecFile, err)
	}

	root, err := xmlSpecTreeParse(content)
	if err != nil {
		return nil, err
	}

	if err := vulkanRegistrySpecDebugTreeWrite(paths.SpecFile, paths.DebugFile, root, xmlSpecTreeDebugDefaultOptions()); err != nil {
		return nil, err
	}

	return root, nil
}

/*
vulkanRegistrySpecEnsure returns a parsed registry tree, reusing an on-disk spec when present.

[Parameters]
specOutputDir is the directory that holds vulkan_spec.xml. The directory is created when missing.

[Returns]
The parsed tree, whether the spec was loaded or freshly fetched, and nil error on success.

[Side Effects]
When vulkan_spec.xml is absent, performs a network GET and writes spec and debug artifacts.
*/
func vulkanRegistrySpecEnsure(specOutputDir string) (*xmlSpecNode, bool, error) {
	paths, err := vulkanRegistrySpecPathsResolve(specOutputDir)
	if err != nil {
		return nil, false, err
	}

	system.DirCreate(filepath.Dir(paths.SpecFile), true)

	if system.FileExists(paths.SpecFile) {
		content, err := system.FileReadAllBytes(paths.SpecFile)
		if err != nil {
			return nil, false, fmt.Errorf("read Vulkan registry spec at %s: %w", paths.SpecFile, err)
		}

		root, err := xmlSpecTreeParse(content)
		if err != nil {
			return nil, false, err
		}

		return root, true, nil
	}

	root, err := vulkanRegistrySpecFetchWriteAndDebug(paths)
	if err != nil {
		return nil, false, err
	}

	return root, false, nil
}
