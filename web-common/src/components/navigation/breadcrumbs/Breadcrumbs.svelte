<script context="module" lang="ts">
  import Slash from "./Slash.svelte";
  import BreadcrumbItem from "./BreadcrumbItem.svelte";

  export type Entry = {
    label: string;
    href: string;
  };

  export type Level = Map<string, Entry>;
</script>

<script lang="ts">
  export let levels: Level[];
  export let selections: string[] = [];

  $: currentPage = selections.findLastIndex((level) => level !== null);
</script>

<nav class="flex gap-x-0 pl-1.5 items-center">
  <slot name="icon" />
  <ol class="flex flex-row items-center">
    {#each levels as options, i (i)}
      {#if selections[i] && options.size}
        {#if i}
          <Slash />
        {/if}
        <BreadcrumbItem
          depth={i}
          {options}
          current={selections[i]}
          isCurrentPage={i === currentPage}
        />
      {/if}
    {/each}
  </ol>
</nav>
