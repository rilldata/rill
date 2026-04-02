<script lang="ts">
  import { ArrowDownNarrowWide, ArrowUpNarrowWide } from "lucide-svelte";
  import type { SortDirection } from "./types";

  let {
    sortDirection = "newest",
    onSortToggle,
  }: {
    sortDirection: SortDirection;
    onSortToggle?: () => void;
  } = $props();

  const sortLabel = $derived(sortDirection === "newest" ? "Newest first" : "Oldest first");
</script>

<button
  class="flex items-center justify-center h-9 w-9 text-fg-primary hover:text-fg-secondary cursor-pointer"
  onclick={() => onSortToggle?.()}
  aria-label={sortLabel}
  title={sortLabel}
>
  {#if sortDirection === "newest"}
    <ArrowDownNarrowWide size={16} />
  {:else}
    <ArrowUpNarrowWide size={16} />
  {/if}
</button>
