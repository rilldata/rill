<script lang="ts">
  import { ArrowUpDown } from "lucide-svelte";
  import type { SortDirection } from "./types";
  import type { RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let {
    sortDirectionStore,
  }: {
    sortDirectionStore: RuneStore<SortDirection>;
  } = $props();

  const sortLabel = $derived(
    sortDirectionStore.value === "newest" ? "Newest" : "Oldest",
  );

  function toggleSortDirection() {
    sortDirectionStore.setter(sortDirectionStore.value ? "oldest" : "newest");
  }
</script>

<button
  type="button"
  class="flex flex-row items-center gap-x-1.5 h-9 px-2 text-sm font-medium text-fg-primary hover:text-fg-secondary cursor-pointer"
  onclick={toggleSortDirection}
  aria-label="Sort order: {sortLabel}"
>
  <span>{sortLabel}</span>
  <ArrowUpDown size={14} />
</button>
