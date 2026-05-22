package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"foundation/system"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const vulkanRefpageManifestFile = "manifest.json"

/*
VulkanRefpageManifest tracks which man pages are cached, known-absent on the registry, and remote cache validators.
*/
type VulkanRefpageManifest struct {
	SpecSHA256    string                             `json:"specSHA256"`
	SymbolSetHash string                             `json:"symbolSetHash"`
	NotFound      map[string]bool                    `json:"notFound,omitempty"`
	Entries       map[string]VulkanRefpageCacheEntry `json:"entries,omitempty"`
	UpdatedAt     time.Time                          `json:"updatedAt"`
}

type VulkanRefpageCacheEntry struct {
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"lastModified,omitempty"`
	CachedAt     time.Time `json:"cachedAt,omitempty"`
}

func vulkanRefpageManifestPath(refpageDir string) string {
	return system.PathJoin(refpageDir, vulkanRefpageManifestFile)
}

func vulkanRefpageManifestLoad(refpageDir string) VulkanRefpageManifest {
	path := vulkanRefpageManifestPath(refpageDir)
	content, err := system.FileReadAllBytes(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VulkanRefpageManifest{
				NotFound: make(map[string]bool),
				Entries:  make(map[string]VulkanRefpageCacheEntry),
			}
		}
		return VulkanRefpageManifest{
			NotFound: make(map[string]bool),
			Entries:  make(map[string]VulkanRefpageCacheEntry),
		}
	}

	var manifest VulkanRefpageManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return VulkanRefpageManifest{
			NotFound: make(map[string]bool),
			Entries:  make(map[string]VulkanRefpageCacheEntry),
		}
	}
	if manifest.NotFound == nil {
		manifest.NotFound = make(map[string]bool)
	}
	if manifest.Entries == nil {
		manifest.Entries = make(map[string]VulkanRefpageCacheEntry)
	}
	return manifest
}

func vulkanRefpageManifestSave(refpageDir string, manifest VulkanRefpageManifest) error {
	manifest.UpdatedAt = time.Now().UTC()
	content, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode refpage manifest: %w", err)
	}
	path := vulkanRefpageManifestPath(refpageDir)
	if err := system.FileWriteBytes(path, content); err != nil {
		return fmt.Errorf("write refpage manifest %s: %w", path, err)
	}
	return nil
}

func vulkanRefpageSpecSHA256(specPath string) (string, error) {
	content, err := system.FileReadAllBytes(specPath)
	if err != nil {
		return "", fmt.Errorf("read spec for refpage manifest: %w", err)
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func vulkanRefpageSymbolSetHash(names []string) string {
	if len(names) == 0 {
		return ""
	}
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	sum := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	return hex.EncodeToString(sum[:])
}

func (manifest VulkanRefpageManifest) Matches(specSHA256 string, symbolSetHash string) bool {
	return manifest.SpecSHA256 == specSHA256 &&
		manifest.SymbolSetHash == symbolSetHash &&
		manifest.SpecSHA256 != ""
}

func vulkanRefpageManifestMarkNotFound(manifest *VulkanRefpageManifest, name string) {
	if manifest.NotFound == nil {
		manifest.NotFound = make(map[string]bool)
	}
	manifest.NotFound[name] = true
	delete(manifest.Entries, name)
}

func vulkanRefpageManifestMarkCached(manifest *VulkanRefpageManifest, name string, etag string, lastModified string) {
	if manifest.Entries == nil {
		manifest.Entries = make(map[string]VulkanRefpageCacheEntry)
	}
	delete(manifest.NotFound, name)
	manifest.Entries[name] = VulkanRefpageCacheEntry{
		ETag:         etag,
		LastModified: lastModified,
		CachedAt:     time.Now().UTC(),
	}
}

func vulkanRefpageManifestPrune(manifest *VulkanRefpageManifest, required map[string]struct{}) {
	for name := range manifest.NotFound {
		if _, ok := required[name]; !ok {
			delete(manifest.NotFound, name)
		}
	}
	for name := range manifest.Entries {
		if _, ok := required[name]; !ok {
			delete(manifest.Entries, name)
		}
	}
}

func vulkanRefpageRequiredSet(names []string) map[string]struct{} {
	required := make(map[string]struct{}, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			required[name] = struct{}{}
		}
	}
	return required
}

func vulkanRefpageNamesMissing(refpageDir string, names []string, manifest VulkanRefpageManifest) []string {
	refpageDirAbs, err := filepath.Abs(refpageDir)
	if err != nil {
		return names
	}

	missing := make([]string, 0)
	for _, name := range names {
		if manifest.NotFound[name] {
			continue
		}
		if system.FileExists(vulkanRefpageCachePath(refpageDirAbs, name)) {
			continue
		}
		missing = append(missing, name)
	}
	sort.Strings(missing)
	return missing
}
