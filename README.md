# gpuarch

Go library for GPU-facing architecture: Vulkan registry bindings, ICD loading, and helpers that connect generated types to manual memory.

## Overview

`gpuarch` is a low-level Vulkan layer. It is not a renderer and does not own windows, swapchains, or frame loops. It provides:

- Types and constants generated from the official Khronos `vk.xml` registry
- Runtime resolution of Vulkan entry points from the platform ICD
- Optional DTO helpers for `memstruct.Array` and `pNext` chains

The API stays generic and registry-faithful so higher-level engines can wrap or curate it without fighting domain assumptions baked into the bindings.

## Requirements

- Go 1.25+
- Vulkan loader and ICD installed on the host (platform-specific)
- This repository and its dependencies obtained **from source** (not published on a module proxy)

Generated bindings and loader sources are committed in this tree; you normally do not run the generator unless you are updating the registry snapshot.

## Obtaining the source

`gpuarch` is not distributed as a versioned Go module on a proxy. Clone the repository into your project, typically as a **git submodule** so the revision stays pinned with the rest of your tree:

```bash
git submodule add <gpuarch-repository-url> third_party/gpuarch
git submodule update --init third_party/gpuarch
```

Use whatever URL and path match your hosting layout. The module path in `go.mod` is `gpuarch`; your `go.work` or `replace` directives must point at the cloned directory.

The same applies to every dependency from this architecture: **syscore**, **memstruct**, **memcore**, **foundation**, **codegen**, and any others your import graph pulls in. Add each repository as its own submodule (or equivalent vendored clone), then wire them locally—do not expect `go get` to resolve them.

### Dependencies

| Module | Required when |
|--------|-------------|
| **syscore** | Always (loader ICD bind) |
| **memcore**, **memstruct** | Using `gpuarch/vulkan/dto` |
| **foundation**, **codegen** | Running `vulkan/internal/generate` only |

Clone each dependency from its own source repository alongside `gpuarch`.

### Wiring with `go.work`

In your application or monorepo root, create a `go.work` that includes every cloned module path:

```go
go 1.25

use (
    ./third_party/gpuarch
    ./third_party/syscore
    ./third_party/memcore
    ./third_party/memstruct
)
```

Adjust paths to match where your submodules live. With `use` entries in place, `import "gpuarch/vulkan/loader"` resolves to the local clone without proxy fetch.

Example submodule layout:

```
your-project/
├── go.work
├── go.mod
└── third_party/
    ├── gpuarch/
    ├── syscore/
    ├── memcore/      # if using dto
    └── memstruct/    # if using dto
```

After adding submodules, run `git submodule update --init --recursive` from the project root whenever you clone the parent repository fresh.

## Module layout

```
gpuarch/
├── go.mod
├── README.md
└── vulkan/
    ├── bindings/          # generated: types, PFNs, VkRegistry* tables
    ├── loader/            # generated holders + ICD module + manifest
    ├── dto/               # Array ↔ Vulkan, pNext chain helpers
    └── internal/
        ├── generate/      # vk.xml + refpage → bindings + loader
        └── spec/          # cached vulkan_spec.xml and man-page HTML
```

| Import path | Role |
|-------------|------|
| `gpuarch/vulkan/bindings` | Generated Vulkan types and registry tables |
| `gpuarch/vulkan/loader` | `VulkanModuleLoad`, command holders, manifest loading |
| `gpuarch/vulkan/dto` | `memstruct.Array` bind/load, `pNext` chain helpers |

Do not import `gpuarch/vulkan/internal/...` from application code.

## Boundaries

**In scope**

- Vulkan 1:1 bindings (structs, enums, flags, handles, PFN types, registry metadata)
- Loader holders and selective command loading via a runtime manifest
- Ergonomic bridges to `memstruct.Array` for counted lists and extension chains

**Out of scope (today)**

- Callable `vk*` wrapper layer (call through `loader` PFN fields on holders)
- `VK_*` C preprocessor defines from the registry
- WSI and platform surface setup beyond ICD `dlopen`
- Validation layers, render passes, or rendering policy

Additional GPU backends may appear under `gpuarch` later; the tree is Vulkan-centric today.

## Quick start

```go
import (
    "gpuarch/vulkan/bindings"
    "gpuarch/vulkan/loader"
)

func main() {
    module, err := loader.VulkanModuleLoad()
    if err != nil {
        panic(err)
    }

    var global loader.VulkanGlobalCommands
    if err := loader.VulkanGlobalCommandsLoad(module, &global); err != nil {
        panic(err)
    }

    // Call through holder fields, e.g. global.CreateInstance(...)
}
```

Selective loading (smaller proc resolve set, extension-friendly):

```go
var manifest loader.VulkanCommandManifest
loader.VulkanCommandManifestReset(&manifest)
loader.VulkanCommandManifestAddCreateInstance(&manifest, &global.CreateInstance)

// After vkCreateInstance: pass real instance and device handles (device may be 0).
err := loader.VulkanCommandsLoadManifest(module, instance, device, &manifest)
```

Typed manifest helpers (`VulkanCommandManifestAddCreateInstance`, etc.) are generated in `loader_manifest_gen.go`.

## Using bindings

Populate `bindings` structs (`VkInstanceCreateInfo`, `VkDeviceCreateInfo`, …) before each call. Every extensible struct uses `SType` and `PNext` as its first fields; set `SType` to the matching `VK_STRUCTURE_TYPE_*` constant.

PFN typedefs for all registry commands live in `bindings_commands_gen.go`. The loader binds the subset you load onto holder structs in `loader_commands_gen.go`.

## Using dto (optional)

Import `gpuarch/vulkan/dto` when extension lists and `pNext` chains are backed by `memstruct.Array`.

**Counted lists** (Vulkan `count` + pointer to first element):

```go
import "gpuarch/vulkan/dto"

// Before the call: array capacity becomes count; element 0 supplies the pointer.
dto.VulkanArrayBindU32(&info.EnabledExtensionCount, &info.PPEnabledExtensionNames, names)

// After the call: copy the driver-filled run into your array.
dto.VulkanArrayLoadU32(count, first, into)
```

**pNext chains**

```go
dto.VulkanStructPNextPrependTyped(&createInfo, &ext, bindings.VK_STRUCTURE_TYPE_…)

// Multiple extensions staged in memstruct.Array[unsafe.Pointer]
dto.VulkanStructPNextChainBindFromPointerArray(unsafe.Pointer(&createInfo), nodes)

// After a query: capture node pointers, then find by sType
dto.VulkanStructPNextChainLoadIntoPointerArray(unsafe.Pointer(&root), captured, true)
node := dto.VulkanStructPNextFindInChain(unsafe.Pointer(&root), wantedSType)
```

## Registry metadata

`bindings` ships tables parsed from vk.xml for capability negotiation:

- `VkRegistryExtensions` — extension enable strings and feature references
- `VkRegistryApiVersions` — core version feature blocks
- `VkRegistryFeatures` — feature struct members and flags

Use these to decide what to enable before filling create-info structs and extension chains.

## Regenerating from vk.xml

Run from the `gpuarch` module root (directory containing `go.mod`):

```bash
# Fetch vk.xml only
go run ./vulkan/internal/generate \
  -specOutputDir=./vulkan/internal/spec

# Full generation (bindings + loader + manifest); offline refpage cache
go run ./vulkan/internal/generate \
  -specOutputDir=./vulkan/internal/spec \
  -bindingOutputDir=./vulkan/bindings \
  -refpageSkipFetch
```

| Flag | Purpose |
|------|---------|
| `-specOutputDir` | Directory for `vulkan_spec.xml` and refpage cache |
| `-bindingOutputDir` | Output for `bindings/*.go`; loader is written to `vulkan/loader/` |
| `-refpageDir` | Man-page HTML cache (default: `<specOutputDir>/refpages`) |
| `-refpageSkipFetch` | Use cache only (no network) |
| `-refpageUpdate` | Re-fetch stale man pages |

Generated documentation merges vk.xml comments with cached Khronos man pages (descriptions, valid usage, member-level VU where available). Do not edit `*_gen.go` by hand.

## Generated artifacts

| Output | Contents |
|--------|----------|
| `bindings_structs_gen.go` | Structs and unions |
| `bindings_enums_gen.go` | Enums and `VkStructureType` |
| `bindings_commands_gen.go` | `PFN_vk*` types |
| `bindings_registry_gen.go` | Extension / API / feature registry |
| `loader_commands_gen.go` | `Vulkan*Commands` holders and tier load functions |
| `loader_manifest_gen.go` | Manifest field constants and typed `VulkanCommandManifestAdd*` helpers |

## Platforms

ICD loading is implemented for Linux, Windows, and macOS. Other GOOS values compile with a stub that returns a clear error from `VulkanModuleLoad`.

## Design

- **C-style API** — Structs and holders hold data; package functions operate on them. Holders expose PFN fields, not methods.
- **Explicit control** — Full tier load vs runtime manifest; bind vs load; prepend vs append on `pNext` chains.
- **LSP-oriented docs** — Generated symbols carry Khronos text so editors can show valid usage in hover without leaving the IDE.

