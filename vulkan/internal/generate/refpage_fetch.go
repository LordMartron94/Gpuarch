package main

import (
	"fmt"
	"foundation/system"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	vulkanRegistryRefpageBaseURL = "https://registry.khronos.org/vulkan/specs/latest/man/html/"
	vulkanRefpageFetchWorkers    = 8
)

type vulkanRefpageFetchOptions struct {
	SkipNetwork bool
	UpdateStale bool
}

/*
vulkanRefpageCorpusEnsure loads cached refpages and fetches only when the spec or symbol set changed, pages are missing, or UpdateStale requests a remote refresh.
*/
func vulkanRefpageCorpusEnsure(
	refpageDir string,
	specPath string,
	names []string,
	options vulkanRefpageFetchOptions,
) (VulkanRefpageCorpus, error) {
	refpageDirAbs, err := filepath.Abs(refpageDir)
	if err != nil {
		return nil, fmt.Errorf("resolve refpage directory: %w", err)
	}

	system.DirCreate(refpageDirAbs, true)

	corpus, err := vulkanRefpageCorpusLoad(refpageDirAbs)
	if err != nil {
		return nil, err
	}

	if options.SkipNetwork {
		return corpus, nil
	}

	specSHA256, err := vulkanRefpageSpecSHA256(specPath)
	if err != nil {
		return nil, err
	}
	symbolSetHash := vulkanRefpageSymbolSetHash(names)
	required := vulkanRefpageRequiredSet(names)

	manifest := vulkanRefpageManifestLoad(refpageDirAbs)
	cacheComplete := manifest.Matches(specSHA256, symbolSetHash)

	missing := vulkanRefpageNamesMissing(refpageDirAbs, names, manifest)
	stale := make([]string, 0)
	if options.UpdateStale {
		stale = vulkanRefpageNamesStale(refpageDirAbs, names, manifest)
	}

	if len(missing) > 0 && vulkanRefpageCachedCount(refpageDirAbs) > 0 && manifest.SpecSHA256 == "" {
		missing = vulkanRefpageProbeMissing(refpageDirAbs, missing, &manifest)
	}

	manifest.SpecSHA256 = specSHA256
	manifest.SymbolSetHash = symbolSetHash
	vulkanRefpageManifestPrune(&manifest, required)

	if len(missing) == 0 && len(stale) == 0 && cacheComplete {
		cached := vulkanRefpageCachedCount(refpageDirAbs)
		fmt.Printf(
			"Refpage cache up to date (%d cached, %d not on registry, %d symbols)\n",
			cached,
			len(manifest.NotFound),
			len(required),
		)
		if err := vulkanRefpageManifestSave(refpageDirAbs, manifest); err != nil {
			return nil, err
		}
		return corpus, nil
	}

	if len(missing) > 0 {
		fmt.Printf("Fetching %d missing Vulkan man pages into %s\n", len(missing), refpageDirAbs)
		if err := vulkanRefpageFetchNames(refpageDirAbs, missing, &manifest); err != nil {
			return nil, err
		}
	}

	if len(stale) > 0 {
		fmt.Printf("Updating %d stale Vulkan man pages in %s\n", len(stale), refpageDirAbs)
		if err := vulkanRefpageFetchNames(refpageDirAbs, stale, &manifest); err != nil {
			return nil, err
		}
	}

	if err := vulkanRefpageManifestSave(refpageDirAbs, manifest); err != nil {
		return nil, err
	}

	reloaded, err := vulkanRefpageCorpusLoad(refpageDirAbs)
	if err != nil {
		return nil, err
	}

	return reloaded, nil
}

func vulkanRefpageCachedCount(refpageDir string) int {
	entries, err := os.ReadDir(refpageDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), vulkanRefpageFileSuffix) {
			continue
		}
		count++
	}
	return count
}

func vulkanRefpageNamesStale(refpageDir string, names []string, manifest VulkanRefpageManifest) []string {
	client := &http.Client{Timeout: 30 * time.Second}
	stale := make([]string, 0)

	for _, name := range names {
		if manifest.NotFound[name] {
			continue
		}
		path := vulkanRefpageCachePath(refpageDir, name)
		if !system.FileExists(path) {
			continue
		}
		entry, ok := manifest.Entries[name]
		if !ok {
			continue
		}
		if vulkanRefpageRemoteChanged(client, name, entry) {
			stale = append(stale, name)
		}
	}

	sort.Strings(stale)
	return stale
}

func vulkanRefpageProbeMissing(refpageDir string, names []string, manifest *VulkanRefpageManifest) []string {
	fmt.Printf("Probing %d uncached Vulkan man page names (HEAD)\n", len(names))

	workers := vulkanRefpageFetchWorkers
	if workers > len(names) {
		workers = len(names)
	}

	nameCh := make(chan string)
	stillMissingCh := make(chan string, len(names))
	var wg sync.WaitGroup

	var manifestMu sync.Mutex
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 30 * time.Second}
			for name := range nameCh {
				url := vulkanRegistryRefpageBaseURL + name + ".html"
				request, err := http.NewRequest(http.MethodHead, url, nil)
				if err != nil {
					stillMissingCh <- name
					continue
				}

				response, err := client.Do(request)
				if err != nil {
					stillMissingCh <- name
					continue
				}
				response.Body.Close()

				switch response.StatusCode {
				case http.StatusNotFound:
					manifestMu.Lock()
					vulkanRefpageManifestMarkNotFound(manifest, name)
					manifestMu.Unlock()
				case http.StatusOK:
					stillMissingCh <- name
				default:
					stillMissingCh <- name
				}
			}
		}()
	}

	for _, name := range names {
		nameCh <- name
	}
	close(nameCh)
	wg.Wait()
	close(stillMissingCh)

	stillMissing := make([]string, 0)
	for name := range stillMissingCh {
		stillMissing = append(stillMissing, name)
	}

	sort.Strings(stillMissing)
	return stillMissing
}

func vulkanRefpageRemoteChanged(client *http.Client, name string, entry VulkanRefpageCacheEntry) bool {
	url := vulkanRegistryRefpageBaseURL + name + ".html"
	request, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return false
	}
	if entry.ETag != "" {
		request.Header.Set("If-None-Match", entry.ETag)
	}
	if entry.LastModified != "" {
		request.Header.Set("If-Modified-Since", entry.LastModified)
	}

	response, err := client.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()

	return response.StatusCode != http.StatusNotModified
}

func vulkanRefpageFetchNames(refpageDir string, names []string, manifest *VulkanRefpageManifest) error {
	if len(names) == 0 {
		return nil
	}

	workers := vulkanRefpageFetchWorkers
	if workers > len(names) {
		workers = len(names)
	}

	nameCh := make(chan string)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 60 * time.Second}
			for name := range nameCh {
				etag, lastModified, status, err := vulkanRefpageFetchOne(client, refpageDir, name)
				if err != nil {
					fmt.Printf("  refpage warning %s: %v\n", name, err)
					continue
				}

				mu.Lock()
				switch status {
				case http.StatusNotFound:
					vulkanRefpageManifestMarkNotFound(manifest, name)
				case http.StatusOK:
					vulkanRefpageManifestMarkCached(manifest, name, etag, lastModified)
				}
				mu.Unlock()
			}
		}()
	}

	for _, name := range names {
		nameCh <- name
	}
	close(nameCh)
	wg.Wait()

	return nil
}

func vulkanRefpageCachePath(refpageDir string, name string) string {
	return system.PathJoin(refpageDir, name+vulkanRefpageFileSuffix)
}

func vulkanRefpageFetchOne(client *http.Client, refpageDir string, name string) (etag string, lastModified string, status int, err error) {
	url := vulkanRegistryRefpageBaseURL + name + ".html"
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("fetch refpage %s: %w", name, err)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", "", 0, fmt.Errorf("fetch refpage %s: %w", name, err)
	}
	defer response.Body.Close()

	etag = response.Header.Get("ETag")
	lastModified = response.Header.Get("Last-Modified")
	status = response.StatusCode

	if status == http.StatusNotFound {
		return etag, lastModified, status, nil
	}

	if status != http.StatusOK {
		return etag, lastModified, status, fmt.Errorf("fetch refpage %s: HTTP %s", name, response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return etag, lastModified, status, fmt.Errorf("read refpage %s: %w", name, err)
	}

	path := vulkanRefpageCachePath(refpageDir, name)
	if err := system.FileWriteBytes(path, body); err != nil {
		return etag, lastModified, status, fmt.Errorf("write refpage %s: %w", path, err)
	}

	return etag, lastModified, status, nil
}
