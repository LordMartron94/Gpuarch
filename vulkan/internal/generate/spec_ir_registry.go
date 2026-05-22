package main

import (
	"sort"
	"strconv"
	"strings"
)

/*
VulkanSpecIRApiVersion is one <feature name="VK_VERSION_..."> block from vk.xml.
*/
type VulkanSpecIRApiVersion struct {
	Name     string
	Number   string
	Depends  string
	Comment  string
	Internal bool
	Features []VulkanSpecIRFeatureRef
}

/*
VulkanSpecIRExtension is one <extension name="VK_..."> block from vk.xml.
*/
type VulkanSpecIRExtension struct {
	Name               string
	Number             string
	ExtensionType      string
	EnableName         string
	ExtensionNameConst string
	SpecVersion        string
	Depends            string
	Supported          string
	Ratified           string
	Platform           string
	Author             string
	Contact            string
	NoFeatures         bool
	Features           []VulkanSpecIRFeatureRef
}

/*
VulkanSpecIRFeatureRef links a capability flag to the struct that exposes it in vkGetPhysicalDeviceFeatures*.
*/
type VulkanSpecIRFeatureRef struct {
	Name           string
	StructType     string
	RequireDepends string
}

/*
VulkanSpecIRFeature is one registry capability entry with its providing extension or API version.
*/
type VulkanSpecIRFeature struct {
	Name           string
	StructType     string
	Provider       string
	RequireDepends string
}

func vulkanSpecIRApiVersionsCollect(root *xmlSpecNode) []VulkanSpecIRApiVersion {
	if root == nil {
		return nil
	}

	var versions []VulkanSpecIRApiVersion
	vulkanSpecIRApiVersionsCollectWalk(root, &versions)
	sort.Slice(versions, func(i, j int) bool { return versions[i].Name < versions[j].Name })
	return versions
}

func vulkanSpecIRApiVersionsCollectWalk(node *xmlSpecNode, versions *[]VulkanSpecIRApiVersion) {
	if node.Name == "feature" && vulkanSpecIRIsApiVersionFeature(node) {
		version := vulkanSpecIRApiVersionBuild(node)
		if version.Name != "" {
			*versions = append(*versions, version)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRApiVersionsCollectWalk(child, versions)
	}
}

func vulkanSpecIRApiVersionBuild(node *xmlSpecNode) VulkanSpecIRApiVersion {
	version := VulkanSpecIRApiVersion{
		Name:     strings.TrimSpace(xmlSpecNodeAttrValue(node, "name")),
		Number:   strings.TrimSpace(xmlSpecNodeAttrValue(node, "number")),
		Depends:  strings.TrimSpace(xmlSpecNodeAttrValue(node, "depends")),
		Comment:  strings.TrimSpace(xmlSpecNodeCommentAttr(node)),
		Internal: strings.TrimSpace(xmlSpecNodeAttrValue(node, "apitype")) == "internal",
	}

	for _, child := range node.Children {
		if child.Name == "require" {
			vulkanSpecIRRegistryRequireFeaturesCollect(child, "", version.Name, &version.Features, nil)
		}
	}

	return version
}

func vulkanSpecIRExtensionsCollect(root *xmlSpecNode) []VulkanSpecIRExtension {
	if root == nil {
		return nil
	}

	var extensions []VulkanSpecIRExtension
	vulkanSpecIRExtensionsCollectWalk(root, &extensions)
	sort.Slice(extensions, func(i, j int) bool {
		ni, _ := strconv.ParseUint(extensions[i].Number, 10, 32)
		nj, _ := strconv.ParseUint(extensions[j].Number, 10, 32)
		if ni != nj {
			return ni < nj
		}
		return extensions[i].Name < extensions[j].Name
	})
	return extensions
}

func vulkanSpecIRExtensionsCollectWalk(node *xmlSpecNode, extensions *[]VulkanSpecIRExtension) {
	if node.Name == "extension" {
		extension := vulkanSpecIRExtensionBuild(node)
		if extension.Name != "" {
			*extensions = append(*extensions, extension)
		}
	}

	for _, child := range node.Children {
		vulkanSpecIRExtensionsCollectWalk(child, extensions)
	}
}

func vulkanSpecIRExtensionBuild(node *xmlSpecNode) VulkanSpecIRExtension {
	extension := VulkanSpecIRExtension{
		Name:          strings.TrimSpace(xmlSpecNodeAttrValue(node, "name")),
		Number:        strings.TrimSpace(xmlSpecNodeAttrValue(node, "number")),
		ExtensionType: strings.TrimSpace(xmlSpecNodeAttrValue(node, "type")),
		Depends:       strings.TrimSpace(xmlSpecNodeAttrValue(node, "depends")),
		Supported:     strings.TrimSpace(xmlSpecNodeAttrValue(node, "supported")),
		Ratified:      strings.TrimSpace(xmlSpecNodeAttrValue(node, "ratified")),
		Platform:      strings.TrimSpace(xmlSpecNodeAttrValue(node, "platform")),
		Author:        strings.TrimSpace(xmlSpecNodeAttrValue(node, "author")),
		Contact:       strings.TrimSpace(xmlSpecNodeAttrValue(node, "contact")),
		NoFeatures:    strings.TrimSpace(xmlSpecNodeAttrValue(node, "nofeatures")) == "true",
	}

	for _, child := range node.Children {
		if child.Name == "require" {
			vulkanSpecIRRegistryRequireApply(child, &extension)
		}
	}

	return extension
}

func vulkanSpecIRRegistryRequireApply(requireNode *xmlSpecNode, extension *VulkanSpecIRExtension) {
	requireDepends := strings.TrimSpace(xmlSpecNodeAttrValue(requireNode, "depends"))

	for _, child := range requireNode.Children {
		switch child.Name {
		case "enum":
			vulkanSpecIRRegistryExtensionEnumApply(child, extension)
		case "feature":
			extension.Features = append(
				extension.Features,
				vulkanSpecIRRegistryFeatureRefsExpand(
					xmlSpecNodeAttrValue(child, "name"),
					xmlSpecNodeAttrValue(child, "struct"),
					requireDepends,
				)...,
			)
		}
	}
}

func vulkanSpecIRRegistryRequireFeaturesCollect(
	requireNode *xmlSpecNode,
	requireDepends string,
	provider string,
	featureRefs *[]VulkanSpecIRFeatureRef,
	features *[]VulkanSpecIRFeature,
) {
	if requireDepends == "" {
		requireDepends = strings.TrimSpace(xmlSpecNodeAttrValue(requireNode, "depends"))
	}

	for _, child := range requireNode.Children {
		switch child.Name {
		case "feature":
			refs := vulkanSpecIRRegistryFeatureRefsExpand(
				xmlSpecNodeAttrValue(child, "name"),
				xmlSpecNodeAttrValue(child, "struct"),
				requireDepends,
			)
			if featureRefs != nil {
				*featureRefs = append(*featureRefs, refs...)
			}
			if features != nil {
				for _, ref := range refs {
					*features = append(*features, VulkanSpecIRFeature{
						Name:           ref.Name,
						StructType:     ref.StructType,
						Provider:       provider,
						RequireDepends: requireDepends,
					})
				}
			}
		case "require":
			vulkanSpecIRRegistryRequireFeaturesCollect(child, requireDepends, provider, featureRefs, features)
		}
	}
}

func vulkanSpecIRRegistryExtensionEnumApply(enumNode *xmlSpecNode, extension *VulkanSpecIRExtension) {
	name := strings.TrimSpace(xmlSpecNodeAttrValue(enumNode, "name"))
	value := vulkanSpecIRRegistryEnumValue(enumNode)
	if name == "" || value == "" {
		return
	}

	switch {
	case strings.HasSuffix(name, "_EXTENSION_NAME"):
		extension.ExtensionNameConst = name
		extension.EnableName = value
	case strings.HasSuffix(name, "_SPEC_VERSION"):
		extension.SpecVersion = value
	}
}

func vulkanSpecIRRegistryEnumValue(enumNode *xmlSpecNode) string {
	raw := strings.TrimSpace(xmlSpecNodeAttrValue(enumNode, "value"))
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		return raw[1 : len(raw)-1]
	}
	return raw
}

func vulkanSpecIRRegistryFeaturesCollect(
	apiVersions []VulkanSpecIRApiVersion,
	extensions []VulkanSpecIRExtension,
) []VulkanSpecIRFeature {
	features := make([]VulkanSpecIRFeature, 0, 1024)

	for _, version := range apiVersions {
		for _, ref := range version.Features {
			features = append(features, VulkanSpecIRFeature{
				Name:           ref.Name,
				StructType:     ref.StructType,
				Provider:       version.Name,
				RequireDepends: ref.RequireDepends,
			})
		}
	}

	for _, extension := range extensions {
		for _, ref := range extension.Features {
			features = append(features, VulkanSpecIRFeature{
				Name:           ref.Name,
				StructType:     ref.StructType,
				Provider:       extension.Name,
				RequireDepends: ref.RequireDepends,
			})
		}
	}

	sort.Slice(features, func(i, j int) bool {
		if features[i].Name != features[j].Name {
			return features[i].Name < features[j].Name
		}
		if features[i].StructType != features[j].StructType {
			return features[i].StructType < features[j].StructType
		}
		return features[i].Provider < features[j].Provider
	})
	return features
}

func vulkanSpecIRRegistryFeatureRefsExpand(name, structType, requireDepends string) []VulkanSpecIRFeatureRef {
	name = strings.TrimSpace(name)
	structType = strings.TrimSpace(structType)
	requireDepends = strings.TrimSpace(requireDepends)
	if name == "" || structType == "" {
		return nil
	}

	names := strings.Split(name, ",")
	refs := make([]VulkanSpecIRFeatureRef, 0, len(names))
	for _, featureName := range names {
		featureName = strings.TrimSpace(featureName)
		if featureName == "" {
			continue
		}
		refs = append(refs, VulkanSpecIRFeatureRef{
			Name:           featureName,
			StructType:     structType,
			RequireDepends: requireDepends,
		})
	}
	return refs
}

func vulkanSpecIRIsApiVersionFeature(node *xmlSpecNode) bool {
	name := strings.TrimSpace(xmlSpecNodeAttrValue(node, "name"))
	return strings.HasPrefix(name, "VK_VERSION_") ||
		strings.HasPrefix(name, "VK_BASE_VERSION_") ||
		strings.HasPrefix(name, "VK_GRAPHICS_VERSION_") ||
		strings.HasPrefix(name, "VK_COMPUTE_VERSION_")
}
