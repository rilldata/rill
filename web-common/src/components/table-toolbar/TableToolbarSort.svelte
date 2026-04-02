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

  const sortLabel = $derived(
    sortDirection === "newest" ? "Newest" : "Oldest",
  );
</script>

<button
  class="flex flex-row items-center gap-x-1.5 text-sm font-medium text-fg-secondary hover:text-fg-primary cursor-pointer"
  onclick={onSortToggle}
>
  <span>{sortLabel}</span>
  <ArrowUpDown size={14} />
</button>
