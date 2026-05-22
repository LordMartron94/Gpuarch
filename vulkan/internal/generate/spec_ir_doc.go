package main

import (
	"html"
	"regexp"
	"strings"
)

const vulkanRegistryManPageBaseURL = "https://registry.khronos.org/vulkan/specs/latest/man/html/"

var vulkanSpecIRDocHTMLTagPattern = regexp.MustCompile(`<[^>]+>`)
var vulkanSpecIRDocWhitespacePattern = regexp.MustCompile(`[ \t]+\n`)

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
xmlSpecNodeDirectCommentsCollect returns documentation from the comment attribute and direct <comment> children.
*/
func xmlSpecNodeDirectCommentsCollect(node *xmlSpecNode) string {
	if node == nil {
		return ""
	}

	var parts []string
	if comment := xmlSpecNodeCommentAttr(node); comment != "" {
		parts = append(parts, comment)
	}

	for _, child := range node.Children {
		if child.Name == "comment" {
			if text := xmlSpecNodeCommentText(child); text != "" {
				parts = append(parts, text)
			}
		}
	}

	return vulkanSpecIRDocJoin(parts...)
}

/*
vulkanSpecIRDocJoin merges documentation fragments, omitting empty parts.
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

/*
vulkanRegistryManPageURL returns the Khronos registry man page URL for a Vulkan API name.
*/
func vulkanRegistryManPageURL(vulkanName string) string {
	vulkanName = strings.TrimSpace(vulkanName)
	if vulkanName == "" {
		return ""
	}
	return vulkanRegistryManPageBaseURL + vulkanName + ".html"
}

/*
vulkanSpecIRDocCompose builds final Go documentation from XML comments, optional refpage prose, and a registry link.

[Parameters]
vulkanName is the registry symbol (for example VkPhysicalDeviceFeatures or VK_SUCCESS). xmlDoc is tier-1 text from vk.xml.
corpus supplies tier-3 refpage prose when present. memberCName is the C member name for struct fields; empty for types and constants.

[Returns]
Documentation text before GoDocFormatExported is applied at render time.
*/
func vulkanSpecIRDocCompose(
	vulkanName string,
	aliasOf string,
	xmlDoc string,
	corpus VulkanRefpageCorpus,
	memberCName string,
) string {
	refpage := corpus.Lookup(vulkanName, aliasOf)

	var parts []string

	if memberCName == "" {
		if summary := strings.TrimSpace(refpage.Summary); summary != "" {
			parts = append(parts, "[Context]\n"+summary)
		} else if xmlDoc != "" {
			parts = append(parts, "[Context]\n"+xmlDoc)
		} else if body := strings.TrimSpace(refpage.Description); body != "" && len(refpage.Members) == 0 {
			parts = append(parts, "[Context]\n"+body)
		}
	} else {
		if memberDoc := strings.TrimSpace(refpage.Members[memberCName]); memberDoc != "" {
			parts = append(parts, memberDoc)
		}
		if xmlDoc != "" {
			if len(parts) == 0 {
				parts = append(parts, xmlDoc)
			} else if !strings.Contains(parts[0], xmlDoc) {
				parts = append(parts, xmlDoc)
			}
		}
	}

	if memberCName == "" {
		if url := vulkanRegistryManPageURL(vulkanName); url != "" {
			parts = append(parts, "[Reference]\n"+url)
		}
	}

	return vulkanSpecIRDocJoin(parts...)
}

const vulkanSpecIRDocGeneratedLead = "is generated from the Khronos Vulkan registry."

/*
vulkanSpecIRDocFormatExported formats documentation for generated Vulkan bindings.

The first sentence names the symbol and states the registry source so linters accept the GoDoc comment. Section tags such as [Context] and [Reference] follow in later paragraphs.
*/
func vulkanSpecIRDocFormatExported(subject string, doc string) string {
	subject = strings.TrimSpace(subject)
	doc = strings.TrimSpace(doc)
	if subject == "" {
		return doc
	}

	prefix := subject + " " + vulkanSpecIRDocGeneratedLead
	if doc == "" {
		return prefix
	}

	if vulkanSpecIRDocHasLeadSentence(subject, doc) {
		return doc
	}

	if strings.HasPrefix(doc, "[") {
		return prefix + "\n\n" + doc
	}

	return prefix + "\n\n" + doc
}

func vulkanSpecIRDocHasLeadSentence(subject string, doc string) bool {
	if !strings.HasPrefix(doc, subject) {
		return false
	}
	remainder := strings.TrimSpace(doc[len(subject):])
	if remainder == "" {
		return false
	}
	if strings.HasPrefix(remainder, "[") {
		return false
	}
	return remainder[0] != '\n'
}

/*
vulkanSpecIRDocFormatStructField formats documentation for generated struct fields.

When the registry supplies member prose, the field name opens the first sentence (for example MinX is the x position...). Additional paragraphs are separated by a blank line. The generated-registry lead is used only when no prose is available.
*/
func vulkanSpecIRDocFormatStructField(subject string, doc string) string {
	subject = strings.TrimSpace(subject)
	doc = strings.TrimSpace(doc)
	if subject == "" {
		return doc
	}
	if doc == "" {
		return subject + " " + vulkanSpecIRDocGeneratedLead
	}
	if vulkanSpecIRDocHasLeadSentence(subject, doc) {
		return doc
	}
	if strings.HasPrefix(doc, "[") {
		return subject + " " + vulkanSpecIRDocGeneratedLead + "\n\n" + doc
	}

	paragraphs := strings.Split(doc, "\n\n")
	paragraphs[0] = subject + " " + strings.TrimSpace(paragraphs[0])
	return strings.Join(paragraphs, "\n\n")
}

/*
vulkanSpecIRDocHTMLToText converts refpage HTML fragments to plain text suitable for GoDoc.
*/
func vulkanSpecIRDocHTMLToText(fragment string) string {
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return ""
	}

	fragment = strings.ReplaceAll(fragment, "<br>", "\n")
	fragment = strings.ReplaceAll(fragment, "<br/>", "\n")
	fragment = strings.ReplaceAll(fragment, "<br />", "\n")
	fragment = strings.ReplaceAll(fragment, "</p>", "\n\n")
	fragment = strings.ReplaceAll(fragment, "</li>", "\n")
	fragment = strings.ReplaceAll(fragment, "<li>", "\n")
	fragment = strings.ReplaceAll(fragment, "</div>", "\n")
	fragment = vulkanSpecIRDocHTMLTagPattern.ReplaceAllString(fragment, "")
	fragment = html.UnescapeString(fragment)
	fragment = strings.ReplaceAll(fragment, "\r\n", "\n")
	fragment = vulkanSpecIRDocWhitespacePattern.ReplaceAllString(fragment, "\n")
	fragment = strings.ReplaceAll(fragment, "\n\n\n", "\n\n")

	lines := strings.Split(fragment, "\n")
	trimmed := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if len(trimmed) > 0 && trimmed[len(trimmed)-1] != "" {
				trimmed = append(trimmed, "")
			}
			continue
		}
		trimmed = append(trimmed, line)
	}

	return strings.TrimSpace(strings.Join(trimmed, "\n"))
}
