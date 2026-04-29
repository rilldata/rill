<script context="module" lang="ts">
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import Slash from "./Slash.svelte";
  import type { PathOptions } from "./types";
</script>

<script lang="ts">
  export let pathParts: (PathOptions | null)[];
  export let currentPath: (string | undefined)[] = [];

  $: currentPage = currentPath.findLastIndex(Boolean);
  // The leading `/` separator should only appear between two rendered
  // breadcrumbs. When earlier depths are skipped (e.g. cloud editor
  // hides the org breadcrumb), suppress the slash that would otherwise
  // appear before the first visible item.
  $: firstRenderedDepth = pathParts.findIndex(
    (p, i) => !!currentPath[i] && !!p?.options,
  );
</script>

<nav class="flex gap-x-2 items-center">
  <slot name="icon" />
  <ol class="flex flex-row items-center">
    {#each pathParts as pathOptions, depth (depth)}
      {@const current = currentPath[depth]}
      {#if current && pathOptions?.options}
        {#if depth > firstRenderedDepth}
          <Slash />
        {/if}
        <BreadcrumbItem
          {depth}
          {pathOptions}
          {current}
          {currentPath}
          isCurrentPage={depth === currentPage}
        />
        {#if depth === 1}<!-- depth 0 = org, depth 1 = project -->
          <slot name="after-project" />
        {/if}
      {/if}
    {/each}
  </ol>
</nav>
