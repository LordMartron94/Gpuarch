/*
Package loader resolves Vulkan entry points from the ICD and exposes them on VulkanCommands.

Generated holders live in loader_commands_gen.go: VulkanCommands groups PFN fields by semantic
endpoint (Global, Instance, PhysicalDevice, Device, Queue, CommandBuffer, and others when the
registry assigns commands to those handles). VulkanModuleLoad opens the platform Vulkan library
(Linux, Windows, or macOS) and binds vkGetInstanceProcAddr and vkGetDeviceProcAddr.

Use VulkanCommandsLoadGlobal, VulkanCommandsLoadInstance, and VulkanCommandsLoadDevice to bind
every entry point in a loader tier into a VulkanCommands value. Or build a VulkanCommandManifest
at runtime (VulkanCommandManifestReset, VulkanCommandManifestAdd*, VulkanCommandsLoadManifest) to
bind only the commands you list. Typed manifest helpers are in loader_manifest_gen.go.

Unloaded PFN fields remain nil. Calling through a field that was not loaded panics with a nil
pointer dereference. With manifest loading, register every command you will call before use.
Holder fields include Khronos man page documentation for LSP hover on the entry points clients
use directly.
*/
package loader
