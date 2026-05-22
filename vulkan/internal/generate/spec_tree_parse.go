package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
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
