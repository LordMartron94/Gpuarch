/*
Package loader resolves Vulkan entry points from the ICD and exposes them on typed command holders.

Generated command tables live in loader_commands_gen.go (VulkanGlobalCommands, VulkanInstanceCommands,
VulkanDeviceCommands). VulkanModuleLoad opens the platform Vulkan library (Linux, Windows, or macOS) and binds
vkGetInstanceProcAddr and vkGetDeviceProcAddr. Use VulkanGlobalCommandsLoad, VulkanInstanceCommandsLoad, and
VulkanDeviceCommandsLoad to populate holders before calling through holder fields. Holder fields include Khronos
man page documentation for LSP hover on the command entry points clients use directly.
*/
package loader
