package main

import "strings"

/*
xmlSpecNodeCommentAttr returns the comment attribute on node when present.
*/
func xmlSpecNodeCommentAttr(node *xmlSpecNode) string {
	return strings.TrimSpace(xmlSpecNodeAttrValue(node, "comment"))
}

/*
xmlSpecNodeCommentText returns trimmed character data on a <comment> element.
*/
func xmlSpecNodeCommentText(node *xmlSpecNode) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Text)
}

/*
vulkanSpecIRDocJoin merges documentation fragments, omitting empty parts.

[Returns]
Fragments joined with a blank line. An empty string when no fragments are present.
*/
func vulkanSpecIRDocJoin(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			nonEmpty = append(nonEmpty, part)
		}
	}
	return strings.Join(nonEmpty, "\n\n")
}
