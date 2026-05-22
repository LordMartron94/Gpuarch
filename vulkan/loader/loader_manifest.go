package loader

import (
	"fmt"

	"gpuarch/vulkan/bindings"
	"syscore"
)

const (
	vulkanCommandLoaderTierGlobal   = 0
	vulkanCommandLoaderTierInstance = 1
	vulkanCommandLoaderTierDevice   = 2
)

/*
VulkanCommandManifest lists Vulkan entry points to resolve at runtime.

Use VulkanCommandManifestReset and VulkanCommandManifestAdd (or a generated VulkanCommandManifestAdd*
helper) to register PFN holder fields, then VulkanCommandsLoadManifest with the handles you already have.
Loader tier (global, instance, or device) is chosen from the registry catalog; callers do not pass tiers.
*/
type VulkanCommandManifest struct {
	entries []vulkanCommandManifestEntry
}

/*
VulkanCommandManifestReset clears all entries from manifest.
*/
func VulkanCommandManifestReset(manifest *VulkanCommandManifest) {
	if manifest == nil {
		return
	}
	manifest.entries = manifest.entries[:0]
}

/*
VulkanCommandManifestLen returns how many commands are listed in manifest.
*/
func VulkanCommandManifestLen(manifest *VulkanCommandManifest) int {
	if manifest == nil {
		return 0
	}
	return len(manifest.entries)
}

/*
VulkanCommandManifestAdd registers one command to load.

field is a VulkanCommandManifestField* constant from loader_manifest_gen.go.
target must be a pointer to the matching PFN field on a command holder.
*/
func VulkanCommandManifestAdd(manifest *VulkanCommandManifest, field VulkanCommandManifestField, target any) error {
	if manifest == nil {
		return fmt.Errorf("vulkan loader: nil manifest")
	}
	if field == VulkanCommandManifestFieldInvalid {
		return fmt.Errorf("vulkan loader: invalid manifest field")
	}
	if target == nil {
		return fmt.Errorf("vulkan loader: nil target for manifest field %v", field)
	}
	if _, ok := vulkanCommandCatalogLookup(field); !ok {
		return fmt.Errorf("vulkan loader: unknown manifest field %v", field)
	}
	manifest.entries = append(manifest.entries, vulkanCommandManifestEntry{
		field:  field,
		target: target,
	})
	return nil
}

func vulkanCommandCatalogLookup(field VulkanCommandManifestField) (vulkanCommandCatalogEntry, bool) {
	if field == VulkanCommandManifestFieldInvalid || field > vulkanCommandManifestFieldMax {
		return vulkanCommandCatalogEntry{}, false
	}
	return vulkanCommandCatalog[field], true
}

func vulkanCommandManifestEntryBind(
	module *VulkanModule,
	instance bindings.VkInstance,
	device bindings.VkDevice,
	entry *vulkanCommandManifestEntry,
	catalog vulkanCommandCatalogEntry,
) error {
	cName := syscore.SYSCORE_C_StringToCStringFirstByte(catalog.vkName)

	var addr uintptr
	switch catalog.tier {
	case vulkanCommandLoaderTierGlobal:
		addr = module.GetInstanceProcAddr(0, cName)
	case vulkanCommandLoaderTierInstance:
		addr = module.GetInstanceProcAddr(instance, cName)
	case vulkanCommandLoaderTierDevice:
		addr = module.GetDeviceProcAddr(device, cName)
	default:
		return fmt.Errorf("vulkan loader: unknown loader tier for %q", catalog.vkName)
	}
	if addr == 0 {
		return fmt.Errorf("vulkan loader: missing entry point %q", catalog.vkName)
	}
	if err := syscore.SYSCORE_Pure_FunctionBindAddress(entry.target, addr); err != nil {
		return fmt.Errorf("vulkan loader: bind %q: %w", catalog.vkName, err)
	}
	return nil
}

/*
VulkanCommandsLoadManifest resolves and binds every command registered on manifest.

instance and device may be zero when the manifest only lists global-tier commands.
*/
func VulkanCommandsLoadManifest(
	module *VulkanModule,
	instance bindings.VkInstance,
	device bindings.VkDevice,
	manifest *VulkanCommandManifest,
) error {
	if manifest == nil {
		return fmt.Errorf("vulkan loader: nil manifest")
	}
	for entryIndex := range manifest.entries {
		entry := &manifest.entries[entryIndex]
		catalog, ok := vulkanCommandCatalogLookup(entry.field)
		if !ok {
			return fmt.Errorf("vulkan loader: invalid manifest field %v", entry.field)
		}
		if err := vulkanCommandManifestEntryBind(module, instance, device, entry, catalog); err != nil {
			return err
		}
	}
	return nil
}
