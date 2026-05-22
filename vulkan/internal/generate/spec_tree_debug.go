package main

import (
	"fmt"
	"foundation/system"
	"path/filepath"
	"strings"
)

const (
	xmlSpecDebugMaxChildrenListed = 40
	xmlSpecDebugMaxTextRunes      = 120
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
	return xmlSpecTreeDebugOptions{
		MaxChildrenListed: xmlSpecDebugMaxChildrenListed,
		MaxTextRunes:      xmlSpecDebugMaxTextRunes,
	}
}

/*
xmlSpecTreeDebugRender formats a parsed registry tree and a short structural summary for human inspection.

[Parameters]
sourcePath labels the registry XML in the dump header. outputPath labels the debug file path in the header.
root is the parsed tree from xmlSpecTreeParse or vulkanRegistrySpecTreeLoadFromFile.

[Returns]
A multi-line text dump. Does not mutate root.
*/
func xmlSpecTreeDebugRender(sourcePath string, outputPath string, root *xmlSpecNode, options xmlSpecTreeDebugOptions) string {
	var builder strings.Builder

	builder.WriteString("Vulkan Registry XML tree (debug)\n")
	builder.WriteString(fmt.Sprintf("Source: %s\n", sourcePath))
	builder.WriteString(fmt.Sprintf("Debug output: %s\n", outputPath))
	builder.WriteString("\n")
	builder.WriteString("=== Registry summary (depth-1 sections) ===\n")
	xmlSpecTreeSummaryWrite(&builder, root)
	builder.WriteString("\n")
	builder.WriteString("=== Element tree ===\n")
	xmlSpecTreeNodeWrite(&builder, root, 0, options)
	builder.WriteString("\n")

	return builder.String()
}

/*
vulkanRegistrySpecDebugTreeWrite renders a parsed registry tree and writes it to outputPath.

[Parameters]
sourcePath labels the registry XML in the dump header. outputPath is the destination .txt file.
root must be non-nil.

[Returns]
nil on success. An error when render or write fails.

[Side Effects]
Overwrites outputPath when it already exists.
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

func xmlSpecTreeSummaryWrite(builder *strings.Builder, root *xmlSpecNode) {
	if root == nil {
		builder.WriteString("(empty document)\n")
		return
	}

	builder.WriteString(fmt.Sprintf("Root: <%s> (%d top-level children)\n", root.Name, len(root.Children)))
	for _, child := range root.Children {
		builder.WriteString(fmt.Sprintf("  <%s>", child.Name))
		if len(child.Attrs) > 0 {
			builder.WriteString(" ")
			builder.WriteString(xmlSpecAttrsInline(child.Attrs))
		}
		builder.WriteString(fmt.Sprintf(" → %d children", len(child.Children)))
		if text := xmlSpecTextTruncate(child.Text, xmlSpecDebugMaxTextRunes); text != "" {
			builder.WriteString(fmt.Sprintf(" text=%q", text))
		}
		builder.WriteByte('\n')
	}
}

func xmlSpecTreeNodeWrite(builder *strings.Builder, node *xmlSpecNode, depth int, options xmlSpecTreeDebugOptions) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)
	builder.WriteString(indent)
	builder.WriteString(fmt.Sprintf("<%s>", node.Name))

	if len(node.Attrs) > 0 {
		builder.WriteString(" ")
		builder.WriteString(xmlSpecAttrsInline(node.Attrs))
	}

	childCount := len(node.Children)
	if childCount > 0 {
		builder.WriteString(fmt.Sprintf(" children=%d", childCount))
	}

	if text := xmlSpecTextTruncate(node.Text, options.MaxTextRunes); text != "" {
		builder.WriteString(fmt.Sprintf(" text=%q", text))
	}
	builder.WriteByte('\n')

	listed := childCount
	if options.MaxChildrenListed > 0 && uint64(childCount) > options.MaxChildrenListed {
		listed = int(options.MaxChildrenListed)
	}

	for i := 0; i < listed; i++ {
		xmlSpecTreeNodeWrite(builder, node.Children[i], depth+1, options)
	}

	if listed < childCount {
		builder.WriteString(indent)
		builder.WriteString("  ")
		builder.WriteString(fmt.Sprintf("... %d more <%s> children not listed\n", childCount-listed, node.Name))
	}
}

func xmlSpecAttrsInline(attrs []xmlSpecAttr) string {
	parts := make([]string, len(attrs))
	for i, attr := range attrs {
		parts[i] = fmt.Sprintf("%s=%q", attr.Name, attr.Value)
	}
	return strings.Join(parts, " ")
}

func xmlSpecTextTruncate(text string, maxRunes int) string {
	text = strings.TrimSpace(text)
	if text == "" || maxRunes <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}
