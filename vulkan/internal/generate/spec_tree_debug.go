package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"foundation/system"
	"io"
	"path/filepath"
	"strings"
)

const (
	xmlSpecDebugMaxChildrenListed = 40
	xmlSpecDebugMaxTextRunes      = 120
)

/*
xmlSpecAttr is one attribute on an XML element in the Vulkan registry tree.
*/
type xmlSpecAttr struct {
	Name  string
	Value string
}

/*
xmlSpecNode is a node in the registry XML tree built by encoding/xml token decoding.
*/
type xmlSpecNode struct {
	Name     string
	Attrs    []xmlSpecAttr
	Children []*xmlSpecNode
	Text     string
}

/*
xmlSpecTreeParseOptions controls how much of a large registry document is expanded in debug output.
*/
type xmlSpecTreeParseOptions struct {
	MaxChildrenListed uint64
	MaxTextRunes      int
}

/*
xmlSpecTreeDefaultOptions returns limits suited to vk.xml debug dumps.
*/
func xmlSpecTreeDefaultOptions() xmlSpecTreeParseOptions {
	return xmlSpecTreeParseOptions{
		MaxChildrenListed: xmlSpecDebugMaxChildrenListed,
		MaxTextRunes:      xmlSpecDebugMaxTextRunes,
	}
}

/*
xmlSpecTreeParse builds a document tree from raw registry XML using encoding/xml.

[Parameters]
content is the full vk.xml bytes. No Vulkan-specific interpretation is applied.

[Returns]
The root element node and nil error on success.
*/
func xmlSpecTreeParse(content []byte) (*xmlSpecNode, error) {
	decoder := xml.NewDecoder(bytes.NewReader(content))

	var stack []*xmlSpecNode
	var root *xmlSpecNode

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decode Vulkan registry XML: %w", err)
		}

		switch element := token.(type) {
		case xml.StartElement:
			node := &xmlSpecNode{
				Name:  element.Name.Local,
				Attrs: xmlSpecAttrsFromXML(element.Attr),
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("decode Vulkan registry XML: unexpected end element </%s>", element.Name.Local)
			}
			stack = stack[:len(stack)-1]

		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := strings.TrimSpace(string(element))
			if text == "" {
				continue
			}
			current := stack[len(stack)-1]
			if current.Text == "" {
				current.Text = text
			} else {
				current.Text += " " + text
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("decode Vulkan registry XML: document has no root element")
	}

	return root, nil
}

func xmlSpecAttrsFromXML(attrs []xml.Attr) []xmlSpecAttr {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]xmlSpecAttr, len(attrs))
	for i, attr := range attrs {
		out[i] = xmlSpecAttr{
			Name:  attr.Name.Local,
			Value: attr.Value,
		}
	}
	return out
}

/*
xmlSpecTreeDebugRender formats the parsed registry tree and a short structural summary for human inspection.
*/
func xmlSpecTreeDebugRender(specXMLPath string, outputPath string, root *xmlSpecNode, options xmlSpecTreeParseOptions) string {
	var builder strings.Builder

	builder.WriteString("Vulkan Registry XML tree (debug)\n")
	builder.WriteString(fmt.Sprintf("Source: %s\n", specXMLPath))
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

func xmlSpecTreeNodeWrite(builder *strings.Builder, node *xmlSpecNode, depth int, options xmlSpecTreeParseOptions) {
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

/*
vulkanRegistrySpecWriteDebugTree reads registry XML from specXMLPath, parses it with encoding/xml, and writes a sibling .txt tree dump.

[Parameters]
specXMLPath must point to an existing vk.xml file. Output is written next to it with the same basename and a .txt extension.

[Returns]
nil on success. An error when read, parse, or write fails.
*/
func vulkanRegistrySpecWriteDebugTree(specXMLPath string, outputPath string) error {
	absXMLPath, err := filepath.Abs(specXMLPath)
	if err != nil {
		return fmt.Errorf("resolve spec XML path: %w", err)
	}

	content, err := system.FileReadAllBytes(absXMLPath)
	if err != nil {
		return fmt.Errorf("read Vulkan registry spec at %s: %w", absXMLPath, err)
	}

	root, err := xmlSpecTreeParse(content)
	if err != nil {
		return err
	}

	rendered := xmlSpecTreeDebugRender(absXMLPath, outputPath, root, xmlSpecTreeDefaultOptions())

	if err := system.FileWriteString(outputPath, rendered); err != nil {
		return fmt.Errorf("write Vulkan registry debug tree to %s: %w", outputPath, err)
	}

	return nil
}
