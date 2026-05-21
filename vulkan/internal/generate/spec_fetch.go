package main

import (
	"fmt"
	"foundation/system"
	"io"
	"net/http"
)

/*
VulkanRegistrySpecURL is the Khronos Vulkan-Docs registry XML used as the canonical API description.
*/
const VulkanRegistrySpecURL = "https://raw.githubusercontent.com/KhronosGroup/Vulkan-Docs/refs/heads/main/xml/vk.xml"

/*
vulkanRegistrySpecFetch downloads the Vulkan registry XML from url and returns the raw document bytes.

[Parameters]
url must be a reachable HTTP(S) location. When empty, VulkanRegistrySpecURL is used.

[Returns]
The full response body on success. An error when the request fails, the status is not 200, or the body cannot be read.

[Side Effects]
Performs a network GET. The caller owns the returned slice.
*/
func vulkanRegistrySpecFetch(url string) ([]byte, error) {
	if url == "" {
		url = VulkanRegistrySpecURL
	}

	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch Vulkan registry spec from %s: %w", url, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Vulkan registry spec from %s: HTTP %s", url, response.Status)
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read Vulkan registry spec body: %w", err)
	}

	return content, nil
}

/*
vulkanRegistrySpecWriteToFile downloads the Vulkan registry XML and writes it to outputPath.

[Parameters]
absOutputPath is the destination file path. Parent directories must already exist.

[Returns]
nil on success. An error when fetch or write fails.

[Side Effects]
Overwrites outputPath when it already exists. Performs a network GET.
*/
func vulkanRegistrySpecWriteToFile(absOutputPath string) error {
	content, err := vulkanRegistrySpecFetch(VulkanRegistrySpecURL)
	if err != nil {
		return err
	}

	if err := system.FileWriteBytes(absOutputPath, content); err != nil {
		return fmt.Errorf("write Vulkan registry spec to %s: %w", absOutputPath, err)
	}

	return nil
}
