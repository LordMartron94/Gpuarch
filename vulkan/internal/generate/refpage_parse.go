package main

import (
	"regexp"
	"strings"
)

const vulkanRefpageSidebarBlockMarker = `<div class="sidebarblock">`

var (
	vulkanRefpageArticlePattern            = regexp.MustCompile(`(?is)<article[^>]*class="doc"[^>]*>(.*)</article>`)
	vulkanRefpageNameSectionPattern        = regexp.MustCompile(`(?is)<h2[^>]*id="_name"[^>]*>.*?</h2>\s*<div class="sectionbody">\s*<div class="paragraph">\s*<p>(.*?)</p>`)
	vulkanRefpageMembersSectionPattern     = regexp.MustCompile(`(?is)<h2[^>]*id="_members"[^>]*>.*?<div class="sectionbody">(.*)<h2[^>]*id="_description"`)
	vulkanRefpageDescriptionSectionPattern = regexp.MustCompile(`(?is)<h2[^>]*id="_description"[^>]*>.*?<div class="sectionbody">(.*?)<h2[^>]*id="_see_also"`)
	vulkanRefpageMemberCodePattern         = regexp.MustCompile(`<code>([a-zA-Z_][a-zA-Z0-9_]*)</code>`)
	vulkanRefpageSidebarTitlePattern       = regexp.MustCompile(`(?is)<div class="title">(.*?)</div>`)
	vulkanRefpageVUIDPrefixPattern         = regexp.MustCompile(`(?m)^VUID-[A-Za-z0-9-]+\s+`)
)

/*
vulkanRefpageHTMLParse extracts structured documentation from a Khronos registry man page HTML document.
*/
func vulkanRefpageHTMLParse(pageHTML string) VulkanRefpageDoc {
	doc := VulkanRefpageDoc{Members: make(map[string]string)}

	article := pageHTML
	if match := vulkanRefpageArticlePattern.FindStringSubmatch(pageHTML); len(match) >= 2 {
		article = match[1]
	}

	if match := vulkanRefpageNameSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		nameLine := vulkanSpecIRDocHTMLToText(match[1])
		doc.Summary = vulkanRefpageSummaryFromNameLine(nameLine)
	}

	if match := vulkanRefpageDescriptionSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		vulkanRefpageDescriptionSectionParse(match[1], &doc)
	}

	if match := vulkanRefpageMembersSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		vulkanRefpageMembersParse(match[1], doc.Members)
	}

	return doc
}

func vulkanRefpageDescriptionSectionParse(sectionHTML string, doc *VulkanRefpageDoc) {
	blocks := strings.Split(sectionHTML, vulkanRefpageSidebarBlockMarker)
	if len(blocks) > 0 {
		doc.Description = vulkanSpecIRDocHTMLToText(blocks[0])
	}

	for _, block := range blocks[1:] {
		titleMatch := vulkanRefpageSidebarTitlePattern.FindStringSubmatch(block)
		if len(titleMatch) < 2 {
			continue
		}

		title := vulkanSpecIRDocHTMLToText(titleMatch[1])
		rules := vulkanRefpageSidebarListItemsParse(block)
		switch strings.TrimSpace(title) {
		case "Valid Usage":
			doc.ValidUsageExplicit = rules
		case "Valid Usage (Implicit)":
			doc.ValidUsageImplicit = rules
		}
	}
}

func vulkanRefpageSidebarListItemsParse(blockHTML string) []string {
	chunks := strings.Split(blockHTML, "<li>")
	rules := make([]string, 0, len(chunks)-1)
	for _, chunk := range chunks[1:] {
		if end := strings.Index(chunk, "</li>"); end >= 0 {
			chunk = chunk[:end]
		}

		rule := vulkanRefpageValidUsageItemClean(vulkanSpecIRDocHTMLToText(chunk))
		if rule != "" {
			rules = append(rules, rule)
		}
	}
	return rules
}

func vulkanRefpageValidUsageItemClean(text string) string {
	text = strings.TrimSpace(text)
	text = vulkanRefpageVUIDPrefixPattern.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func vulkanRefpageMembersParse(sectionHTML string, members map[string]string) {
	chunks := strings.Split(sectionHTML, "<li>")
	for _, chunk := range chunks[1:] {
		if end := strings.Index(chunk, "</li>"); end >= 0 {
			chunk = chunk[:end]
		}

		codeMatch := vulkanRefpageMemberCodePattern.FindStringSubmatch(chunk)
		if len(codeMatch) < 2 {
			continue
		}

		memberName := codeMatch[1]
		memberDoc := vulkanSpecIRDocHTMLToText(chunk)
		memberDoc = strings.TrimSpace(strings.TrimPrefix(memberDoc, memberName))
		memberDoc = strings.TrimLeft(memberDoc, ":- \t")
		if memberDoc != "" {
			members[memberName] = memberDoc
		}
	}
}

func vulkanRefpageSummaryFromNameLine(nameLine string) string {
	nameLine = strings.TrimSpace(nameLine)
	if nameLine == "" {
		return ""
	}

	if idx := strings.Index(nameLine, " - "); idx >= 0 {
		return strings.TrimSpace(nameLine[idx+len(" - "):])
	}

	return nameLine
}
