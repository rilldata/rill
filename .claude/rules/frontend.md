---
paths: *.svelte, *.ts
---

# Frontend Style Guide

This is a living document that captures frontend conventions and best practices for the Rill web applications.

## TypeScript

- File names use kebab-case
- Boolean variables: `isX` (e.g., `isConversationLoading`)
- Functions use `function` keyword, not arrow functions
- Prefer options objects over multiple parameters for optional configuration
- Avoid too many layers of abstraction
- Prefer `null` over empty string `""` when representing "no value" — it's more semantically clear

## Naming Conventions

| Type                | Convention       | Example                   |
| ------------------- | ---------------- | ------------------------- |
| TypeScript files    | kebab-case       | `user-management.ts`      |
| Svelte components   | PascalCase       | `ProjectCard.svelte`      |
| Directories         | kebab-case       | `user-management/`        |
| Variables/functions | camelCase        | `getProjectPermissions()` |
| True constants      | UPPER_SNAKE_CASE | `MAX_FILE_SIZE`           |
| Interfaces/types    | PascalCase       | `ProjectPermissions`      |

## Svelte

- Keep components small and focused
- Do not use `createEventDispatcher` (deprecated in Svelte 5) — use callback props instead
- Use SuperForms for form handling
- Prefer idiomatic Svelte over patterns from other frameworks
- Use semantic HTML
- Lean into the existing design system — don't custom-build modals, popovers, dropdowns
- In script blocks, place function declarations at the bottom
- Break long reactive statement sequences into logical groups with comments

### Styles

- Default to component-scoped `<style lang="postcss">` blocks
- Use semantic classes with Tailwind v3 via `@apply`
- Break long `@apply` statements into multiple lines, grouped logically
- Use Tailwind theme colors; rarely use custom colors
- Prefer global styles over inline overrides — check `app.css` first

## TanStack Query

### Core Principles

- Lean into TanStack Query paradigms
- Put Query Observers in the component that needs the data, not in TypeScript files
- Use `select` to transform data rather than wrapping with `derived` stores
- Components must handle `isLoading` and `isError` states
- Prefer direct cache updates over query invalidation when complexity allows

### Query Key Pattern

Create query keys using Orval-generated functions:

```typescript
getRuntimeServiceGetResourceQueryKey(...)
```

### Observer Naming

Name observers like `queryNameQuery`:

```typescript
const getConversationQuery = createQuery(...)
```

### Recommended Pattern

```svelte
<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";
  import { derived } from "svelte/store";
  import { getRuntimeServiceGetConversationQueryOptions } from ".../runtime-client";
  import { runtime } from ".../runtime-client/runtime-store";

  // Reactive QueryOptions store
  const getConversationQueryOptionsStore = derived(
    [runtime, currentConversation],
    ([$runtime, $currentConversation]) =>
      getRuntimeServiceGetConversationQueryOptions(
        $runtime.instanceId,
        $currentConversation?.id || "",
        undefined, // Use undefined for unused optional params, not {}
        {
          query: {
            enabled: !!$currentConversation?.id,
          },
        },
      ),
  );

  const getConversationQuery = createQuery(getConversationQueryOptionsStore);
</script>
```

## State Management Patterns

| Pattern                  | Use When                                           | Example                  |
| ------------------------ | -------------------------------------------------- | ------------------------ |
| Direct TanStack Query    | Single component, simple data (1-2 queries)        | Query in component       |
| TypeScript Query Factory | Complex query logic, reusable across components    | `useFilteredTableData()` |
| ES6 Classes              | Client state + server coordination, business logic | `ChatStateManager`       |

**Key principle**: TanStack Query owns server state. Classes own client state and coordinate between them.

### URL as Source of Truth

When state can be represented in the URL, it should be:

- Shareable links that restore exact state
- Browser back/forward works intuitively
- Bookmarkable views

**URL-appropriate**: filters, time ranges, tabs, search queries, pagination
**Not URL-appropriate**: loading states, hover/focus, sensitive data, rapidly changing state

## File Organization

- Organize by semantic features (user flows), not file types
- Each sub-feature contains all layers (UI, logic, state) for that flow
- Use `shared/` for cross-cutting concerns
- Co-locate documentation with code (`README.md` in each sub-feature)
- Target ~8 files per directory maximum
- `web-common/src/lib/` is for low-level, domain-agnostic utilities only

```
features/my-feature/
├── user-flow-1/         # Everything for flow 1
├── user-flow-2/         # Everything for flow 2
└── shared/              # Shared utilities

# NOT this:
features/my-feature/
├── components/          # UI layer
├── lib/                 # Logic layer
└── utils/               # Utility layer
```

## General

- Comments communicate to other developers, not to AI assistants
- Our app is an SPA — no need for `if (browser)` blocks in `+page.ts`
