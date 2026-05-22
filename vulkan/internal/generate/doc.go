/*
Package main generates Vulkan registry bindings from vk.xml and Khronos man page documentation.

Binding generation uses three documentation tiers:

  - Tier 1: <comment> attributes and child elements in vk.xml (short registry hints).
  - Tier 2: [Reference] links to https://registry.khronos.org/vulkan/specs/latest/man/html/.
  - Tier 3: cached Khronos man page HTML under -refpageDir (default <specOutputDir>/refpages). Struct and union
    types include [Description], [Valid Usage], and [Valid Usage (Implicit)]; struct fields include member-level
    valid usage when the refpage attributes rules to a member.
  - Registry metadata: bindings_registry_gen.go exposes VkRegistryExtension, VkRegistryApiVersion, and
    VkRegistryFeature tables parsed from vk.xml for extension enable strings and capability negotiation.

Run with -bindingOutputDir and -specOutputDir. Man pages are fetched only when the spec or symbol set changes,
pages are missing from the cache, or -refpageUpdate is set. Known registry 404s are recorded in refpages/manifest.json
and are not re-fetched. Pass -refpageSkipFetch to use the cache only with no network access.

Example:

	go run ./libs/gpuarch/vulkan/internal/generate \
	  -specOutputDir=./libs/gpuarch/vulkan/internal/spec \
	  -bindingOutputDir=./libs/gpuarch/vulkan/bindings
*/
package main
