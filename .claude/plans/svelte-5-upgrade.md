# Svelte 5 Migration Plan

## Context

The frontend is on Svelte 4.2.19 with SvelteKit 2.7.1 across ~1,142 `.svelte` files (web-common, web-admin, web-local). Svelte 5 has a **legacy compatibility mode** — Svelte 4 syntax works unchanged, and components only enter "runes mode" when they use runes. This means we can bump to Svelte 5 and keep all existing code working, then migrate incrementally.

**Strategy: Two PRs.**

- **PR 1** — Minimal Svelte 5 bump. Only fix what breaks.
- **PR 2** — Upgrade bits-ui from 0.22 to 2.x (dedicated effort).

---

# PR 1: Svelte 5 Bump (Minimal)

## What Must Change

Only two packages import from `svelte/internal` (removed in Svelte 5) and will **hard break**:

1. **`@tanstack/svelte-table` v8** → replace with [`tanstack-table-8-svelte-5`](https://github.com/dummdidumm/tanstack-table-8-svelte-5) (by dummdidumm/Simon Holthausen, Svelte maintainer)
2. **`@sveltejs/vite-plugin-svelte` v3** → upgrade to v5 (Svelte 5 requires v4+)

One package has **behavioral breakage** (doesn't crash, but filtering stops working):

3. **`cmdk-sv` 0.0.19** → DOM filtering breaks due to Svelte 5 rendering changes. Package is archived (May 2025). Only used in 3 files in web-admin.

## What Works in Legacy Mode (No Changes Needed for PR 1)

| Package                          | Why                                                                                                                                                                                                                                                                     |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `bits-ui` 0.22.0                 | No `svelte/internal` imports; melt-ui runtime is clean                                                                                                                                                                                                                  |
| `@melt-ui/svelte` (transitive)   | No `svelte/internal` in runtime                                                                                                                                                                                                                                         |
| `svelte-radix` 1.1.0             | Ships `.svelte` source files; compiler handles legacy syntax                                                                                                                                                                                                            |
| `lucide-svelte` 0.298.0          | Ships `.svelte` source; Vite resolves source files via export conditions                                                                                                                                                                                                |
| `@tanstack/svelte-virtual` 3.0.1 | Store-based adapter; works in legacy mode. Has known issues with runes (`$state` element binding, count reactivity) but these only surface when components are migrated to runes. No fix exists in any version — the Svelte adapter has never been rewritten for runes. |
| `@tanstack/svelte-query` 5.69.0  | Only imports from `svelte/store` (public API)                                                                                                                                                                                                                           |
| `sveltekit-superforms` 2.19.1    | Works in legacy mode                                                                                                                                                                                                                                                    |
| `@xyflow/svelte` 0.1.39          | Needs testing, but should work in legacy mode                                                                                                                                                                                                                           |
| `svelte-vega` 2.3.0              | Needs testing                                                                                                                                                                                                                                                           |
| `@storybook/svelte` 7.6.17       | Needs testing; may require upgrade to 8.x                                                                                                                                                                                                                               |

## Steps

### Step 1: Bump Svelte and Vite Plugin

**Files to modify:**

- `package.json` (root) — update `@sveltejs/vite-plugin-svelte` override from `^3.1.2` to `^5.0.0`
- `web-common/package.json` — `svelte` from `^4.2.19` to `^5.0.0`
- `web-admin/package.json` — `svelte` from `^4.2.19` to `^5.0.0`, `@sveltejs/vite-plugin-svelte` devDep if present
- `web-local/package.json` — `svelte` from `^4.2.19` to `^5.0.0`, `@sveltejs/vite-plugin-svelte` devDep if present
- Run `npm install`

**Also check/update:**

- `svelte-check` — currently `^4.0.4`, v4 already supports Svelte 5 (it was released for Svelte 5)
- `@sveltejs/kit` — `^2.7.1` already supports Svelte 5 (SvelteKit 2.12+ has Svelte 5 support)

### Step 2: Replace `@tanstack/svelte-table`

**Install:** `npm install tanstack-table-8-svelte-5 -w web-common` (and web-admin if listed there)
**Remove:** `npm uninstall @tanstack/svelte-table -w web-common -w web-admin`

**Code changes (34 files):**

- Change imports: `from "@tanstack/svelte-table"` → `from "tanstack-table-8-svelte-5"`
- API is near-identical: `createSvelteTable`, `flexRender` stay the same
- **One breaking change:** custom component rendering must use `renderComponent(SomeCell, props)` instead of passing the component directly. Need to audit column definitions for this pattern.
- Re-exports from `@tanstack/table-core` (types like `ColumnDef`, `TableOptions`, etc.) should work unchanged

**Key files to update:**

- `web-common/src/components/table/BasicTable.svelte`
- `web-common/src/components/table/VirtualizedTable.svelte`
- `web-common/src/components/table/InfiniteScrollTable.svelte`
- `web-common/src/features/dashboards/pivot/` (PivotTable, FlatTable, NestedTable)
- `web-common/src/components/table/tanstack-table-column-meta.ts`
- All files importing from `@tanstack/svelte-table`

### Step 3: Handle `cmdk-sv` Breakage

`cmdk-sv` (archived May 2025) has behavioral breakage in Svelte 5 — DOM filtering stops working. The footprint is small:

- **Wrapper layer:** 8 files + index.ts in `web-common/src/components/command/`
- **Consumers:** only 3 files in web-admin:
  - `web-admin/src/features/view-as-user/ViewAsUserPopover.svelte`
  - `web-admin/src/features/organizations/ShareOrganizationCTA.svelte`
  - `web-admin/src/features/ai/mcp/OAuthSection.svelte`

**Approach:** Test first after the Svelte 5 bump. If filtering is broken, rewrite the 8 wrapper files to use a simple custom implementation (basic input + filtered list) without pulling in a new dependency. The consumer components should need no changes since they use the wrapper API.

### Step 4: Verify

- `npm run build` in all three workspaces
- `npm run test -w web-common` (unit tests)
- `npm run quality` (lint/format)
- Manual smoke test:
  - Dashboard table rendering (TanStack Table)
  - Virtualized tables / pivot tables
  - Dialog, Select, Combobox interactions (bits-ui)
  - Command palette (cmdk-sv)
  - General navigation and page loads
- Playwright E2E: `npm run test -w web-admin` and `npm run test -w web-local`

---

# PR 2: bits-ui 0.22 → 2.x

## Why Upgrade

bits-ui 0.22 works in Svelte 5 legacy mode, so it's not blocking PR 1. However:

- bits-ui 2.x is the actively maintained version; 0.22 will stop getting fixes
- bits-ui 2.x replaces cmdk-sv with a built-in Command component (fixes the cmdk-sv breakage from PR 1)
- bits-ui 2.x uses Svelte 5 snippets instead of the Melt UI builder pattern, which is cleaner

## Breaking Changes (0.22 → 2.x)

| Pattern            | bits-ui 0.22                                        | bits-ui 2.x                     |
| ------------------ | --------------------------------------------------- | ------------------------------- |
| Composition        | `asChild let:builder` + `builderActions`/`getAttrs` | `child` snippet                 |
| Slot content       | `<slot>` / `let:` directives                        | Snippet callback props          |
| Transitions        | Transition props on components                      | Use Svelte transitions directly |
| Selection          | `Selected` type, `selected` binding                 | Native value binding            |
| Multiple selection | `multiple` prop                                     | `type="multiple"`               |
| Events             | `on:click`, `on:change`                             | Callback props                  |

Full migration guide: https://bits-ui.com/docs/migration-guide

## Scope

**Wrapper components to rewrite (15 component folders in `web-common/src/components/`):**

| Component            | Wrapper files | Consumer files | Risk                           |
| -------------------- | :-----------: | :------------: | ------------------------------ |
| Dialog               |      ~8       |      16+       | `asChild` on triggers          |
| AlertDialog          |      ~8       |      10+       | `asChild` on triggers          |
| DropdownMenu         |      ~10      |      25+       | `asChild`, builder composition |
| ContextMenu          |      ~10      |      10+       | Same patterns as DropdownMenu  |
| Popover              |      ~3       |      15+       | `asChild` on triggers          |
| Select               |      ~5       |      10+       | `Selected` type, `asChild`     |
| Tabs                 |      ~5       |      10+       | Minimal changes                |
| Tooltip / Tooltip-v2 |      ~3       |   ubiquitous   | `asChild` on triggers          |
| Collapsible          |      ~3       |       5+       | `asChild` + `builderActions`   |
| Checkbox             |       1       |       5+       | API rework                     |
| Switch               |       1       |       5+       | API rework                     |
| Avatar               |       1       |       3+       | Minimal                        |
| Progress             |       1       |       2+       | Minimal                        |
| Label                |       1       |       5+       | Minimal                        |

**Highest-risk patterns:**

- **126 `asChild` usages** across ~126 files (web-common: 92, web-admin: 33, web-local: 1) — each needs conversion to the `child` snippet pattern
- **9 files using `builderActions`/`getAttrs`** — these are the deepest builder integrations (Button, Chip, DropdownMenuItem, SelectorButton, ExpandableOption)
- **Direct bits-ui imports** in ~79 web-common files and ~2 web-admin files

**Strategy:** Since all bits-ui usage goes through the wrapper layer in `web-common/src/components/`, the wrapper rewrite absorbs most of the API changes. Consumer components mostly need:

- Remove `asChild` + `let:builder` and use the new composition pattern
- Update any `Selected` type usage
- Update event handler patterns

**Also in this PR:**

- Replace `cmdk-sv` with bits-ui 2.x `Command` component (8 wrapper files in `web-common/src/components/command/`, 3 consumers in web-admin)
- Remove `cmdk-sv` dependency

---

# Future Work (After PR 1 and PR 2)

Deferred upgrades that can be done incrementally:

- **lucide-svelte → @lucide/svelte** — package rename, not urgent
- **svelte-radix → @lucide/svelte** — consolidate icon libraries
- **@tanstack/svelte-query v5 → v6** — runes-native adapter
- **Storybook 7.x → 8.x** — Svelte 5 support
- **Component migration to runes** — `npx sv migrate svelte-5` can automate most of it. ~1,080 `$:` declarations, ~1,825 `export let` props, ~252 slots, ~477 event directives. Migrate workspace-by-workspace: web-local (25 components) → web-admin (375) → web-common (550).
- **`@tanstack/svelte-virtual`** — Svelte adapter needs a runes rewrite (or a custom adapter using `@tanstack/virtual-core`) before components using it can be migrated to runes. No upstream fix exists; open issues [#866](https://github.com/TanStack/virtual/issues/866), [#969](https://github.com/TanStack/virtual/issues/969).
- **`sveltekit-superforms` v3** — runes-native (when released)
- **`@xyflow/svelte` 0.x → 1.x** — Svelte 5 native (significant API changes)
