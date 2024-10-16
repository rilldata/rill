<script context="module" lang="ts">
  import BreadcrumbItem from "./BreadcrumbItem.svelte";
  import Slash from "./Slash.svelte";
  import type { PathOptions } from "./types";
</script>

<script lang="ts">
  export let pathParts: (PathOptions | null)[];
  export let currentPath: (string | undefined)[] = [];

  $: currentPage = currentPath.findLastIndex(Boolean);
</script>

<nav class="flex gap-x-2 items-center">
  <slot name="icon" />
  <ol class="flex flex-row items-center">
    {#each pathParts as options, depth (depth)}
      {@const current = currentPath[depth]}
      {#if current && options?.size}
        {#if depth}
          <Slash />
        {/if}
        <BreadcrumbItem
          {depth}
          {options}
          {current}
          {currentPath}
          isCurrentPage={depth === currentPage}
        />
      {/if}
    {/each}
  </ol>
</nav>
