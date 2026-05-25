package main

import (
	"strings"

	specxml "foundation/spec/xml"
)

type xmlSpecNode = specxml.Node

type xmlSpecAttr = specxml.Attr

func xmlSpecTreeParse(content []byte) (*xmlSpecNode, error) {
	return specxml.TreeParse(content)
}

func xmlSpecNodeAttrValue(node *xmlSpecNode, attrName string) string {
	return node.Attr(attrName)
}

func xmlSpecNodeDirectCommentsCollect(node *xmlSpecNode) string {
	return node.DirectComments()
}

func xmlSpecNodeCommentText(commentNode *xmlSpecNode) string {
	if commentNode == nil {
		return ""
	}
	return strings.TrimSpace(commentNode.Text)
}

func xmlSpecNodeCommentAttr(node *xmlSpecNode) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Attr("comment"))
}
