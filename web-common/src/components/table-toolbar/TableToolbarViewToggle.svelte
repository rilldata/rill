<script lang="ts">
  import List from "@rilldata/web-common/components/icons/List.svelte";
  import { LayoutGrid } from "lucide-svelte";
  import type { ViewMode } from "./types";

  let {
    viewMode = $bindable("list"),
  }: {
    viewMode: ViewMode;
  } = $props();
</script>

<div class="flex flex-row border overflow-hidden">
  <button
    type="button"
    class="flex items-center justify-center w-9 h-9 border-r {viewMode ===
    'grid'
      ? 'bg-surface-hover'
      : 'bg-white hover:bg-surface-hover'} rounded-l-[2px]"
    onclick={() => (viewMode = "grid")}
    aria-label="Grid view"
    aria-pressed={viewMode === "grid"}
  >
    <LayoutGrid size={16} />
  </button>
  <button
    type="button"
    class="flex items-center justify-center w-9 h-9 {viewMode === 'list'
      ? 'bg-surface-hover'
      : 'bg-white hover:bg-surface-hover'} rounded-r-[2px]"
    onclick={() => (viewMode = "list")}
    aria-label="List view"
    aria-pressed={viewMode === "list"}
  >
    <List size="16px" />
  </button>
</div>
