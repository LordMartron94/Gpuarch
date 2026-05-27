# gpuarch

Go library for GPU-facing architecture: Vulkan registry bindings, ICD loading, and helpers that connect generated types to manual memory.

## Overview

`gpuarch` is a low-level Vulkan layer. It is not a renderer and does not own windows, swapchains, or frame loops. It provides:

- Types and constants generated from the official Khronos `vk.xml` registry
- Runtime resolution of Vulkan entry points from the platform ICD
- Optional DTO helpers for `memstruct.Array` and `pNext` chains
- Optional VkResult assertion helpers and Go error conversion

The API stays generic and registry-faithful so higher-level engines can wrap or curate it without fighting domain assumptions baked into the bindings.

## Disclaimer

While this library is maintained by its developer and used in real projects, **most of `gpuarch` is generated code produced with Cursor AI assistance** (Vulkan bindings, loader holders, registry tables, and related `*_gen.go` artifacts). It is **not guaranteed to be correct or complete**, even though bugs and crashes encountered in use are typically fixed and pushed when found.

Treat generated surfaces as best-effort registry mirrors: validate behavior against the Vulkan spec, validation layers, and your target hardware. Hand-written packages (`dto`, `result`, and generator tooling) receive the same scrutiny but are still evolving.

## Requirements

- Go 1.25+
- Vulkan loader and ICD installed on the host (platform-specific)
- All modules in the import graph available to the Go toolchain (see [Dependencies](#dependencies))

Generated bindings and loader sources are committed in this tree; you normally do not run the generator unless you are updating the registry snapshot.

## Obtaining the source

You can consume `gpuarch` in two ways: **clone into the tree** (submodule or vendor) or **import by module path** when the repository is reachable from your module proxy or VCS (for example GitHub). Both approaches require the same companion libraries; `gpuarch` does not stand alone.

### Submodule or local clone (recommended for pinning)

Clone the repository into your project, typically as a **git submodule** so the revision stays pinned with the rest of your tree:

```bash
git submodule add <gpuarch-repository-url> third_party/gpuarch
git submodule update --init third_party/gpuarch
```

Use whatever URL and path match your hosting layout. The module path in `go.mod` is `gpuarch`; your `go.work` or `replace` directives must point at the cloned directory.

The same applies to every dependency from this architecture (see [Dependencies](#dependencies)): add each repository as its own submodule (or equivalent vendored clone), then wire them locally.

### Standard Go imports (GitHub or other VCS)

When the published `go.mod` uses a fully qualified module path, you can import packages with normal Go import statements, for example:

```go
import (
    "github.com/LordMartron94/gpuarch/vulkan/bindings"
    "github.com/LordMartron94/gpuarch/vulkan/loader"
    "github.com/LordMartron94/gpuarch/vulkan/result"
)
```

Add the module to your project:

```bash
go get github.com/LordMartron94/gpuarch@<version>
```

Replace the org, repository name, and version with whatever matches the remote you use. If the upstream `go.mod` still declares `module gpuarch` (short path), either import with `gpuarch/vulkan/...` and satisfy that path via `go.work` / `replace`, or align the remote module path with the GitHub URL before relying on `go get` alone.

**You must also make every dependency module resolvable.** Importing `gpuarch/vulkan/loader` alone is not enough: the build will pull in **syscore** (always), and **memcore** / **memstruct** if you use `dto`. Each of those modules must be present in your `go.work`, vendored tree, or available at its own import path with matching `go get` / `replace` entries. Missing a sibling module produces unresolved-import errors even when `gpuarch` itself resolves.

Example `go.mod` fragment when everything is on GitHub under the same org:

```go
require (
    github.com/LordMartron94/gpuarch v0.0.0
    github.com/LordMartron94/syscore v0.0.0
    // github.com/LordMartron94/memcore, memstruct — when using dto
)
```

Use `replace` directives during local development if you clone siblings side by side instead of tagging releases.

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

Adjust paths to match where your submodules live. With `use` entries in place, `import "gpuarch/vulkan/loader"` (or the fully qualified GitHub path, if that is what the remote `go.mod` declares) resolves to the local clone without proxy fetch.

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
    ├── result/            # VkResult assert + error conversion
    └── internal/
        ├── generate/      # vk.xml + refpage → bindings + loader
        └── spec/          # cached vulkan_spec.xml and man-page HTML
```

| Import path | Role |
|-------------|------|
| `gpuarch/vulkan/bindings` | Generated Vulkan types and registry tables |
| `gpuarch/vulkan/loader` | `VulkanModuleLoad`, command holders, manifest loading |
| `gpuarch/vulkan/dto` | `memstruct.Array` bind/load, `pNext` chain helpers |
| `gpuarch/vulkan/result` | VkResult classification, assert helpers, `error` conversion |

When the remote module path is fully qualified, prefix the same suffixes with your VCS root (for example `github.com/LordMartron94/gpuarch/vulkan/loader`).

Do not import `gpuarch/vulkan/internal/...` from application code.

## Boundaries

**In scope**

- Vulkan 1:1 bindings (structs, enums, flags, handles, PFN types, registry metadata)
- Loader holders and selective command loading via a runtime manifest
- Ergonomic bridges to `memstruct.Array` for counted lists and extension chains
- VkResult helpers: strict success checks, allowed status codes, failure-only errors

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

    var commands loader.VulkanCommands
    if err := loader.VulkanCommandsLoadGlobal(module, &commands); err != nil {
        panic(err)
    }

    // Call through endpoint fields, e.g. commands.Global.CreateInstance(...)
}
```

Selective loading (smaller proc resolve set, extension-friendly):

```go
var commands loader.VulkanCommands
var manifest loader.VulkanCommandManifest
loader.VulkanCommandManifestReset(&manifest)
loader.VulkanCommandManifestAddCreateInstance(&manifest)

// After vkCreateInstance: pass real instance and device handles (device may be 0).
err := loader.VulkanCommandsLoadManifest(module, instance, device, &manifest, &commands)
```

Register every command you will call on the manifest. Unloaded PFN fields on `VulkanCommands` stay nil; calling them panics with a nil pointer dereference.

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

## Using result (optional)

Import `gpuarch/vulkan/result` after Vulkan calls that return `bindings.VkResult`:

```go
import (
    "gpuarch/vulkan/bindings"
    "gpuarch/vulkan/result"
)

if err := result.VulkanResultAssertSuccess(vkCreateInstance(...)); err != nil {
    return err
}

// Enumeration may return VK_INCOMPLETE:
if err := result.VulkanResultAssertSuccessOr(r, bindings.VK_INCOMPLETE); err != nil {
    return err
}

// Map to standard Go errors:
if err := result.VulkanResultToErrorIfFailure(r); err != nil {
    return err
}
```

`VulkanResultToError` treats any non-`VK_SUCCESS` value as an error (including `VK_NOT_READY`). `VulkanResultToErrorIfFailure` only errors on negative failure codes.

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

# Full generation (bindings + result names + loader + manifest); offline refpage cache
go run ./vulkan/internal/generate \
  -specOutputDir=./vulkan/internal/spec \
  -bindingOutputDir=./vulkan/bindings \
  -refpageSkipFetch
```

| Flag | Purpose |
|------|---------|
| `-specOutputDir` | Directory for `vulkan_spec.xml` and refpage cache |
| `-bindingOutputDir` | Output for `bindings/*.go`; `result/` and `loader/` are written beside it under `vulkan/` |
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

