<script lang="ts">
  import { ArrowUpDown } from "lucide-svelte";
  import type { SortDirection } from "./types";

  let {
    sortDirection = "newest",
    onSortToggle,
  }: {
    sortDirection: SortDirection;
    onSortToggle?: () => void;
  } = $props();

  const sortLabel = $derived(sortDirection === "newest" ? "Newest" : "Oldest");
</script>

<button
  type="button"
  class="flex flex-row items-center gap-x-1.5 h-9 px-2 text-sm font-medium text-fg-primary hover:text-fg-secondary cursor-pointer"
  onclick={() => onSortToggle?.()}
  aria-label="Sort order: {sortLabel}"
>
  <span>{sortLabel}</span>
  <ArrowUpDown size={14} />
</button>
