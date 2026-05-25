package main

import (
	"fmt"
	"foundation/spec/xml"
	"foundation/system"
	"path/filepath"
)

/*
xmlSpecTreeDebugOptions controls how much of a large registry document is expanded in debug output.
*/
type xmlSpecTreeDebugOptions struct {
	MaxChildrenListed uint64
	MaxTextRunes      int
}

/*
xmlSpecTreeDebugDefaultOptions returns limits suited to vk.xml debug dumps.
*/
func xmlSpecTreeDebugDefaultOptions() xmlSpecTreeDebugOptions {
	defaults := xml.DebugDefaultOptions("Vulkan Registry XML tree (debug)")
	return xmlSpecTreeDebugOptions{
		MaxChildrenListed: defaults.MaxChildrenListed,
		MaxTextRunes:      defaults.MaxTextRunes,
	}
}

func xmlSpecTreeDebugRender(sourcePath string, outputPath string, root *xmlSpecNode, options xmlSpecTreeDebugOptions) string {
	return xml.TreeDebugRender(sourcePath, outputPath, root, xml.DebugOptions{
		Title:             "Vulkan Registry XML tree (debug)",
		MaxChildrenListed: options.MaxChildrenListed,
		MaxTextRunes:      options.MaxTextRunes,
	})
}

/*
vulkanRegistrySpecDebugTreeWrite renders a parsed registry tree and writes it to outputPath.
*/
func vulkanRegistrySpecDebugTreeWrite(sourcePath string, outputPath string, root *xmlSpecNode, options xmlSpecTreeDebugOptions) error {
	absSourcePath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("resolve spec XML path: %w", err)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve debug tree output path: %w", err)
	}

	rendered := xmlSpecTreeDebugRender(absSourcePath, absOutputPath, root, options)

	if err := system.FileWriteString(absOutputPath, rendered); err != nil {
		return fmt.Errorf("write Vulkan registry debug tree to %s: %w", absOutputPath, err)
	}

	return nil
}
