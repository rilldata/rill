<!-- Used by dashboard list/grid view toggle -->
<script lang="ts">
  import List from "@rilldata/web-common/components/icons/List.svelte";
  import { LayoutGrid } from "lucide-svelte";
  import type { ViewMode } from "./types";

  let {
    viewMode = "list",
    onViewModeChange,
  }: {
    viewMode: ViewMode;
    onViewModeChange?: (mode: ViewMode) => void;
  } = $props();
</script>

<div class="flex flex-row border rounded-sm overflow-hidden">
  <button
    type="button"
    class="flex items-center justify-center w-8 h-8 {viewMode === 'grid'
      ? 'bg-surface-hover'
      : 'bg-input hover:bg-surface-hover'}"
    onclick={() => onViewModeChange?.("grid")}
    aria-label="Grid view"
    aria-pressed={viewMode === "grid"}
  >
    <LayoutGrid size={16} />
  </button>
  <button
    type="button"
    class="flex items-center justify-center w-8 h-8 border-l {viewMode ===
    'list'
      ? 'bg-surface-hover'
      : 'bg-input hover:bg-surface-hover'}"
    onclick={() => onViewModeChange?.("list")}
    aria-label="List view"
    aria-pressed={viewMode === "list"}
  >
    <List size="16px" />
  </button>
</div>
