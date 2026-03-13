# Svelte 5 Migration Plan

## Context

The frontend is on Svelte 4.2.19 with SvelteKit 2.7.1 across ~1,142 `.svelte` files (web-common, web-admin, web-local). Svelte 5 has a **legacy compatibility mode** — Svelte 4 syntax works unchanged, and components only enter "runes mode" when they use runes. This means we can bump to Svelte 5 and keep all existing code working, then migrate incrementally.

**Strategy: Single PR** — Svelte 5 bump + bits-ui 2.x upgrade together.

bits-ui 2.x requires Svelte 5 (peer dep `^5.33.0`), and bits-ui 0.22 is not reliably compatible with Svelte 5: its dependency `@melt-ui/svelte` 0.76.2 explicitly excludes Svelte 5 (`>=3 <5`) and has known runtime bugs on Svelte 5. The bits-ui maintainer has confirmed that 0.x will not be updated for Svelte 5 compatibility. These upgrades must happen atomically.

**Commit structure** (within the single PR):

1. Svelte 5 bump + vite plugin + tanstack table replacement
2. bits-ui 0.22 → 2.x wrapper rewrites
3. bits-ui consumer component updates
4. cmdk-sv → bits-ui Command replacement + cmdk-sv removal

---

## Step 1: Bump Svelte, Vite Plugin, and Replace TanStack Table

### Svelte and Vite Plugin

**Files to modify:**

- `package.json` (root) — update `@sveltejs/vite-plugin-svelte` override from `^3.1.2` to `^5.0.0`
- `web-common/package.json` — `svelte` from `^4.2.19` to `^5.0.0`
- `web-admin/package.json` — `svelte` from `^4.2.19` to `^5.0.0`, `@sveltejs/vite-plugin-svelte` devDep if present
- `web-local/package.json` — `svelte` from `^4.2.19` to `^5.0.0`, `@sveltejs/vite-plugin-svelte` devDep if present
- Run `npm install`

**Also check/update:**

- `svelte-check` — currently `^4.0.4`, v4 already supports Svelte 5 (it was released for Svelte 5)
- `@sveltejs/kit` — `^2.7.1` already supports Svelte 5 (SvelteKit 2.12+ has Svelte 5 support)

### Replace `@tanstack/svelte-table`

`@tanstack/svelte-table` v8 imports from `svelte/internal` (removed in Svelte 5) and will hard break.

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

### Other Packages (Legacy Mode — No Changes Needed)

| Package                          | Why it works                                                                                                                                                                                                                                                            |
| -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `svelte-radix` 1.1.0             | Ships `.svelte` source files; compiler handles legacy syntax                                                                                                                                                                                                            |
| `lucide-svelte` 0.298.0          | Ships `.svelte` source; Vite resolves source files via export conditions                                                                                                                                                                                                |
| `@tanstack/svelte-virtual` 3.0.1 | Store-based adapter; works in legacy mode. Has known issues with runes (`$state` element binding, count reactivity) but these only surface when components are migrated to runes. No fix exists in any version — the Svelte adapter has never been rewritten for runes. |
| `@tanstack/svelte-query` 5.69.0  | Only imports from `svelte/store` (public API)                                                                                                                                                                                                                           |
| `sveltekit-superforms` 2.19.1    | Works in legacy mode                                                                                                                                                                                                                                                    |
| `@xyflow/svelte` 0.1.39          | Needs testing, but should work in legacy mode                                                                                                                                                                                                                           |
| `svelte-vega` 2.3.0              | Needs testing                                                                                                                                                                                                                                                           |
| `@storybook/svelte` 7.6.17       | Needs testing; may require upgrade to 8.x                                                                                                                                                                                                                               |

---

## Step 2: Upgrade bits-ui 0.22 → 2.x

### Why This Can't Be Deferred

- bits-ui 0.22's dependency `@melt-ui/svelte` 0.76.2 declares `svelte: ">=3 <5"` — explicitly excludes Svelte 5
- Known melt-ui runtime bugs on Svelte 5: broken component actions (issue #749), incorrect PIN input updates (issue #1263)
- bits-ui maintainer (issue #1023): *"We aren't going to be bumping Melt in `bits-ui@0.x`. If you're using Svelte 5, it's recommended to use `bits-ui@next`."*
- bits-ui 2.x requires Svelte 5 (`^5.33.0`); it cannot run on Svelte 4
- Therefore the two upgrades must happen atomically

### Breaking Changes (0.22 → 2.x)

| Pattern            | bits-ui 0.22                                        | bits-ui 2.x                     |
| ------------------ | --------------------------------------------------- | ------------------------------- |
| Composition        | `asChild let:builder` + `builderActions`/`getAttrs` | `child` snippet                 |
| Slot content       | `<slot>` / `let:` directives                        | Snippet callback props          |
| Transitions        | Transition props on components                      | Use Svelte transitions directly |
| Selection          | `Selected` type, `selected` binding                 | Native value binding            |
| Multiple selection | `multiple` prop                                     | `type="multiple"`               |
| Events             | `on:click`, `on:change`                             | Callback props                  |

Full migration guide: https://bits-ui.com/docs/migration-guide

### Scope

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

**Strategy:** All bits-ui usage goes through the wrapper layer in `web-common/src/components/`. The wrapper rewrite absorbs most of the API changes. Consumer components mostly need:

- Remove `asChild` + `let:builder` and use the new composition pattern
- Update any `Selected` type usage
- Update event handler patterns

---

## Step 3: Replace `cmdk-sv` with bits-ui Command

`cmdk-sv` (archived May 2025) has behavioral breakage in Svelte 5 and is no longer maintained. bits-ui 2.x includes a built-in `Command` component that replaces it.

- **Wrapper layer:** 8 files + index.ts in `web-common/src/components/command/`
- **Consumers:** 3 files in web-admin:
  - `web-admin/src/features/view-as-user/ViewAsUserPopover.svelte`
  - `web-admin/src/features/organizations/ShareOrganizationCTA.svelte`
  - `web-admin/src/features/ai/mcp/OAuthSection.svelte`

**Approach:** Rewrite the 8 wrapper files to use bits-ui 2.x `Command` component. Remove `cmdk-sv` dependency.

---

## Step 4: Verify

- `npm run build` in all three workspaces
- `npm run test -w web-common` (unit tests)
- `npm run quality` (lint/format)
- Manual smoke test:
  - Dashboard table rendering (TanStack Table)
  - Virtualized tables / pivot tables
  - Dialog, Select, Combobox interactions (bits-ui 2.x)
  - Command palette (bits-ui Command)
  - General navigation and page loads
- Playwright E2E: `npm run test -w web-admin` and `npm run test -w web-local`

---

# Future Work (After This PR)

Deferred upgrades that can be done incrementally:

- **lucide-svelte → @lucide/svelte** — package rename, not urgent
- **svelte-radix → @lucide/svelte** — consolidate icon libraries
- **@tanstack/svelte-query v5 → v6** — runes-native adapter
- **Storybook 7.x → 8.x** — Svelte 5 support
- **Component migration to runes** — `npx sv migrate svelte-5` can automate most of it. ~1,080 `$:` declarations, ~1,825 `export let` props, ~252 slots, ~477 event directives. Migrate workspace-by-workspace: web-local (25 components) → web-admin (375) → web-common (550).
- **`@tanstack/svelte-virtual`** — Svelte adapter needs a runes rewrite (or a custom adapter using `@tanstack/virtual-core`) before components using it can be migrated to runes. No upstream fix exists; open issues [#866](https://github.com/TanStack/virtual/issues/866), [#969](https://github.com/TanStack/virtual/issues/969).
- **`sveltekit-superforms` v3** — runes-native (when released)
- **`@xyflow/svelte` 0.x → 1.x** — Svelte 5 native (significant API changes)
