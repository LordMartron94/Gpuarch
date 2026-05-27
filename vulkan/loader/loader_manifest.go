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
helper) to register commands, then VulkanCommandsLoadManifest with the handles you already have and a
VulkanCommands holder. Loader tier (global, instance, or device) is chosen from the registry catalog;
callers do not pass tiers.

Unloaded PFN fields on VulkanCommands remain nil. Calling through a field that was not loaded (manifest
omission or partial tier load) panics with a nil pointer dereference. Register every command you will call.
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
*/
func VulkanCommandManifestAdd(manifest *VulkanCommandManifest, field VulkanCommandManifestField) error {
	if manifest == nil {
		return fmt.Errorf("vulkan loader: nil manifest")
	}
	if field == VulkanCommandManifestFieldInvalid {
		return fmt.Errorf("vulkan loader: invalid manifest field")
	}
	if _, ok := vulkanCommandCatalogLookup(field); !ok {
		return fmt.Errorf("vulkan loader: unknown manifest field %v", field)
	}
	manifest.entries = append(manifest.entries, vulkanCommandManifestEntry{
		field: field,
	})
	return nil
}

func vulkanCommandCatalogLookup(field VulkanCommandManifestField) (vulkanCommandCatalogEntry, bool) {
	if field == VulkanCommandManifestFieldInvalid || field > vulkanCommandManifestFieldMax {
		return vulkanCommandCatalogEntry{}, false
	}
	return vulkanCommandCatalog[field], true
}

func vulkanCommandsManifestTarget(commands *VulkanCommands, field VulkanCommandManifestField) (any, error) {
	if commands == nil {
		return nil, fmt.Errorf("vulkan loader: nil commands")
	}
	if field == VulkanCommandManifestFieldInvalid || int(field) >= len(vulkanCommandsManifestTargetByField) {
		return nil, fmt.Errorf("vulkan loader: unknown manifest field %v", field)
	}
	resolve := vulkanCommandsManifestTargetByField[field]
	if resolve == nil {
		return nil, fmt.Errorf("vulkan loader: unknown manifest field %v", field)
	}
	return resolve(commands), nil
}

func vulkanCommandManifestEntryBind(
	module *VulkanModule,
	instance bindings.VkInstance,
	device bindings.VkDevice,
	entry *vulkanCommandManifestEntry,
	catalog vulkanCommandCatalogEntry,
	target any,
) error {
	if catalog.tier == vulkanCommandLoaderTierInstance && instance == 0 {
		return vulkanLoaderEntryPointResolveError(catalog.vkName, catalog.tier, instance, device)
	}
	if catalog.tier == vulkanCommandLoaderTierDevice && device == 0 {
		return vulkanLoaderEntryPointResolveError(catalog.vkName, catalog.tier, instance, device)
	}

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
		return vulkanLoaderEntryPointResolveError(catalog.vkName, catalog.tier, instance, device)
	}
	if err := syscore.SYSCORE_Pure_FunctionBindAddress(target, addr); err != nil {
		return fmt.Errorf("vulkan loader: bind %q: %w", catalog.vkName, err)
	}

	return nil
}

/*
VulkanCommandsLoadManifest resolves and binds every command registered on manifest into commands.

commands must be non-nil. Unloaded PFN fields on commands stay nil; calling them panics.

instance and device may be zero when the manifest only lists global-tier commands.
*/
func VulkanCommandsLoadManifest(
	module *VulkanModule,
	instance bindings.VkInstance,
	device bindings.VkDevice,
	manifest *VulkanCommandManifest,
	commands *VulkanCommands,
) error {
	if manifest == nil {
		return fmt.Errorf("vulkan loader: nil manifest")
	}
	if commands == nil {
		return fmt.Errorf("vulkan loader: nil commands")
	}
	for entryIndex := range manifest.entries {
		entry := &manifest.entries[entryIndex]
		catalog, ok := vulkanCommandCatalogLookup(entry.field)
		if !ok {
			return fmt.Errorf("vulkan loader: invalid manifest field %v", entry.field)
		}
		target, err := vulkanCommandsManifestTarget(commands, entry.field)
		if err != nil {
			return err
		}
		if err := vulkanCommandManifestEntryBind(module, instance, device, entry, catalog, target); err != nil {
			return err
		}
	}
	return nil
}
