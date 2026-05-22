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
	vulkanRefpageMemberCodeTextPattern     = regexp.MustCompile(`(?m)^` + "`" + `?([a-zA-Z_][a-zA-Z0-9_]*)` + "`" + `?\s`)
	vulkanRefpageSidebarTitlePattern       = regexp.MustCompile(`(?is)<div class="title">(.*?)</div>`)
	vulkanRefpageVUIDPrefixPattern         = regexp.MustCompile(`(?m)^VUID-[A-Za-z0-9-]+\s+`)
	vulkanRefpageMemberListPattern         = regexp.MustCompile(`(?is)<div class="ulist">\s*<ul>(.*?)</ul>\s*</div>`)
)

/*
vulkanRefpageHTMLParse extracts structured documentation from a Khronos registry man page HTML document.
*/
func vulkanRefpageHTMLParse(pageHTML string) VulkanRefpageDoc {
	doc := VulkanRefpageDoc{Members: make(map[string]VulkanRefpageMemberDoc)}

	article := pageHTML
	if match := vulkanRefpageArticlePattern.FindStringSubmatch(pageHTML); len(match) >= 2 {
		article = match[1]
	}

	if match := vulkanRefpageNameSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		nameLine := vulkanSpecIRDocHTMLToText(match[1])
		doc.Summary = vulkanRefpageSummaryFromNameLine(nameLine)
	}

	if match := vulkanRefpageMembersSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		vulkanRefpageMembersSectionParse(match[1], doc.Members)
	}

	if match := vulkanRefpageDescriptionSectionPattern.FindStringSubmatch(article); len(match) >= 2 {
		vulkanRefpageDescriptionSectionParse(match[1], &doc)
	}

	return doc
}

func vulkanRefpageDescriptionSectionParse(sectionHTML string, doc *VulkanRefpageDoc) {
	blocks := strings.Split(sectionHTML, vulkanRefpageSidebarBlockMarker)
	intro := blocks[0]

	vulkanRefpageDescriptionMemberListsParse(intro, doc.Members)
	doc.Description = vulkanRefpageDescriptionProseWithoutMemberLists(intro)

	for _, block := range blocks[1:] {
		titleMatch := vulkanRefpageSidebarTitlePattern.FindStringSubmatch(block)
		if len(titleMatch) < 2 {
			continue
		}

		title := vulkanSpecIRDocHTMLToText(titleMatch[1])
		rules := vulkanRefpageSidebarListItemsParse(block)
		switch strings.TrimSpace(title) {
		case "Valid Usage":
			doc.ValidUsageExplicit = append(doc.ValidUsageExplicit, rules...)
		case "Valid Usage (Implicit)":
			doc.ValidUsageImplicit = append(doc.ValidUsageImplicit, rules...)
		}
	}

	vulkanRefpageMemberValidUsageDistribute(doc)
}

func vulkanRefpageDescriptionMemberListsParse(introHTML string, members map[string]VulkanRefpageMemberDoc) {
	for _, listHTML := range vulkanRefpageMemberListPattern.FindAllString(introHTML, -1) {
		vulkanRefpageMembersListParse(listHTML, members)
	}
}

func vulkanRefpageDescriptionProseWithoutMemberLists(introHTML string) string {
	prose := vulkanRefpageMemberListPattern.ReplaceAllString(introHTML, "")
	return strings.TrimSpace(vulkanSpecIRDocHTMLToText(prose))
}

func vulkanRefpageMembersSectionParse(sectionHTML string, members map[string]VulkanRefpageMemberDoc) {
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
		memberDoc := vulkanRefpageMemberChunkParse(chunk)
		if !memberDoc.IsEmpty() {
			existing := members[memberName]
			members[memberName] = vulkanRefpageMemberDocMerge(existing, memberDoc)
		}
	}
}

func vulkanRefpageMembersListParse(listHTML string, members map[string]VulkanRefpageMemberDoc) {
	chunks := strings.Split(listHTML, "<li>")
	for _, chunk := range chunks[1:] {
		if end := strings.Index(chunk, "</li>"); end >= 0 {
			chunk = chunk[:end]
		}

		codeMatch := vulkanRefpageMemberCodePattern.FindStringSubmatch(chunk)
		if len(codeMatch) < 2 {
			continue
		}

		memberName := codeMatch[1]
		memberDoc := vulkanRefpageMemberChunkParse(chunk)
		if !memberDoc.IsEmpty() {
			existing := members[memberName]
			members[memberName] = vulkanRefpageMemberDocMerge(existing, memberDoc)
		}
	}
}

func vulkanRefpageMemberChunkParse(chunkHTML string) VulkanRefpageMemberDoc {
	blocks := strings.Split(chunkHTML, vulkanRefpageSidebarBlockMarker)
	proseHTML := blocks[0]

	memberName := ""
	if codeMatch := vulkanRefpageMemberCodePattern.FindStringSubmatch(proseHTML); len(codeMatch) >= 2 {
		memberName = codeMatch[1]
	}

	summary := vulkanSpecIRDocHTMLToText(proseHTML)
	if memberName != "" {
		summary = strings.TrimSpace(strings.TrimPrefix(summary, memberName))
		summary = strings.TrimLeft(summary, ":- \t")
	}

	doc := VulkanRefpageMemberDoc{Summary: summary}
	for _, block := range blocks[1:] {
		titleMatch := vulkanRefpageSidebarTitlePattern.FindStringSubmatch(block)
		if len(titleMatch) < 2 {
			continue
		}

		title := vulkanSpecIRDocHTMLToText(titleMatch[1])
		rules := vulkanRefpageSidebarListItemsParse(block)
		switch strings.TrimSpace(title) {
		case "Valid Usage":
			doc.ValidUsageExplicit = append(doc.ValidUsageExplicit, rules...)
		case "Valid Usage (Implicit)":
			doc.ValidUsageImplicit = append(doc.ValidUsageImplicit, rules...)
		}
	}
	return doc
}

func vulkanRefpageMemberValidUsageDistribute(doc *VulkanRefpageDoc) {
	if len(doc.Members) == 0 {
		return
	}

	memberNames := make([]string, 0, len(doc.Members))
	for name := range doc.Members {
		memberNames = append(memberNames, name)
	}

	doc.ValidUsageExplicit = vulkanRefpageMemberValidUsageRouteRules(doc.ValidUsageExplicit, memberNames, doc.Members, false)
	doc.ValidUsageImplicit = vulkanRefpageMemberValidUsageRouteRules(doc.ValidUsageImplicit, memberNames, doc.Members, true)
}

func vulkanRefpageMemberValidUsageRouteRules(
	rules []string,
	memberNames []string,
	members map[string]VulkanRefpageMemberDoc,
	implicit bool,
) []string {
	structRules := make([]string, 0, len(rules))
	for _, rule := range rules {
		targets := vulkanRefpageValidUsageMemberNames(rule, memberNames)
		if len(targets) == 0 {
			structRules = append(structRules, rule)
			continue
		}
		for _, memberName := range targets {
			member := members[memberName]
			if implicit {
				member.ValidUsageImplicit = append(member.ValidUsageImplicit, rule)
			} else {
				member.ValidUsageExplicit = append(member.ValidUsageExplicit, rule)
			}
			members[memberName] = member
		}
	}
	return structRules
}

func vulkanRefpageValidUsageMemberNames(rule string, memberNames []string) []string {
	seen := make(map[string]struct{})
	var targets []string

	add := func(name string) {
		for _, memberName := range memberNames {
			if name == memberName {
				if _, ok := seen[memberName]; !ok {
					seen[memberName] = struct{}{}
					targets = append(targets, memberName)
				}
				return
			}
		}
	}

	for _, match := range vulkanRefpageMemberCodePattern.FindAllStringSubmatch(rule, -1) {
		if len(match) >= 2 {
			add(match[1])
		}
	}

	if len(targets) == 0 {
		clean := vulkanRefpageValidUsageItemClean(vulkanSpecIRDocHTMLToText(rule))
		for _, match := range vulkanRefpageMemberCodeTextPattern.FindAllStringSubmatch(clean, -1) {
			if len(match) >= 2 {
				add(match[1])
			}
		}
	}

	if len(targets) == 0 {
		for _, memberName := range memberNames {
			if strings.Contains(rule, memberName) {
				add(memberName)
			}
		}
	}

	return targets
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
