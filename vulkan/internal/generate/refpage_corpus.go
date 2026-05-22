package main

import (
	"fmt"
	"foundation/system"
	"os"
	"path/filepath"
	"strings"
)

const vulkanRefpageFileSuffix = ".html"

/*
VulkanRefpageMemberDoc is parsed Khronos man page prose for one struct or union member.
*/
type VulkanRefpageMemberDoc struct {
	Summary            string
	ValidUsageExplicit []string
	ValidUsageImplicit []string
}

func (doc VulkanRefpageMemberDoc) IsEmpty() bool {
	return strings.TrimSpace(doc.Summary) == "" &&
		len(doc.ValidUsageExplicit) == 0 &&
		len(doc.ValidUsageImplicit) == 0
}

/*
VulkanRefpageDoc is parsed Khronos registry man page prose for one API name.
*/
type VulkanRefpageDoc struct {
	Summary            string
	Description        string
	ValidUsageExplicit []string
	ValidUsageImplicit []string
	Members            map[string]VulkanRefpageMemberDoc
}

/*
VulkanRefpageCorpus maps Vulkan API names to parsed man page documentation.
*/
type VulkanRefpageCorpus map[string]VulkanRefpageDoc

/*
Lookup returns refpage documentation for name, falling back to aliasOf when the primary name is absent.
*/
func (corpus VulkanRefpageCorpus) Lookup(name string, aliasOf string) VulkanRefpageDoc {
	if corpus == nil {
		return VulkanRefpageDoc{Members: make(map[string]VulkanRefpageMemberDoc)}
	}
	if doc, ok := corpus[name]; ok && !doc.IsEmpty() {
		return doc
	}
	if aliasOf != "" {
		if doc, ok := corpus[aliasOf]; ok {
			return doc
		}
	}
	return VulkanRefpageDoc{Members: make(map[string]VulkanRefpageMemberDoc)}
}

func (doc VulkanRefpageDoc) IsEmpty() bool {
	if strings.TrimSpace(doc.Summary) != "" ||
		strings.TrimSpace(doc.Description) != "" ||
		len(doc.ValidUsageExplicit) != 0 ||
		len(doc.ValidUsageImplicit) != 0 {
		return false
	}
	for _, member := range doc.Members {
		if !member.IsEmpty() {
			return false
		}
	}
	return true
}

/*
vulkanRefpageCorpusNamesFromIR collects Vulkan symbol names that may have registry man pages.
*/
func vulkanRefpageCorpusNamesFromIR(ir VulkanSpecIR) []string {
	seen := make(map[string]struct{})
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		seen[name] = struct{}{}
	}

	for _, basetype := range ir.Basetypes {
		add(basetype.Name)
	}
	for _, handle := range ir.Handles {
		add(handle.Name)
	}
	for _, enumType := range ir.Enums {
		add(enumType.Name)
		for _, value := range enumType.Values {
			add(value.Key)
		}
	}
	for _, flagType := range ir.Flags {
		add(flagType.Name)
		add(flagType.AggregateName)
		for _, value := range flagType.Values {
			add(value.Key)
		}
	}
	for _, aggregate := range ir.Structs {
		add(aggregate.Name)
	}
	for _, fn := range ir.Funcpointers {
		add(fn.Name)
	}
	for _, command := range ir.Commands {
		add(command.Name)
	}
	if ir.Constants.Name != "" {
		for _, value := range ir.Constants.Values {
			add(value.Key)
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names
}

/*
vulkanRefpageCorpusLoad reads cached man page HTML files from refpageDir keyed by Vulkan API name.
*/
func vulkanRefpageCorpusLoad(refpageDir string) (VulkanRefpageCorpus, error) {
	refpageDirAbs, err := filepath.Abs(refpageDir)
	if err != nil {
		return nil, fmt.Errorf("resolve refpage directory: %w", err)
	}

	if !system.FileExists(refpageDirAbs) {
		return VulkanRefpageCorpus{}, nil
	}

	entries, err := os.ReadDir(refpageDirAbs)
	if err != nil {
		return nil, fmt.Errorf("read refpage directory %s: %w", refpageDirAbs, err)
	}

	corpus := make(VulkanRefpageCorpus)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), vulkanRefpageFileSuffix) {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), vulkanRefpageFileSuffix)
		if name == "" {
			continue
		}

		path := system.PathJoin(refpageDirAbs, entry.Name())
		content, err := system.FileReadAllBytes(path)
		if err != nil {
			return nil, fmt.Errorf("read refpage %s: %w", path, err)
		}

		corpus[name] = vulkanRefpageHTMLParse(string(content))
	}

	return corpus, nil
}

func vulkanRefpageMemberDocMerge(existing VulkanRefpageMemberDoc, add VulkanRefpageMemberDoc) VulkanRefpageMemberDoc {
	if strings.TrimSpace(existing.Summary) == "" {
		existing.Summary = add.Summary
	} else if add.Summary != "" && !strings.Contains(existing.Summary, add.Summary) {
		existing.Summary = vulkanSpecIRDocJoin(existing.Summary, add.Summary)
	}
	existing.ValidUsageExplicit = append(existing.ValidUsageExplicit, add.ValidUsageExplicit...)
	existing.ValidUsageImplicit = append(existing.ValidUsageImplicit, add.ValidUsageImplicit...)
	return existing
}
