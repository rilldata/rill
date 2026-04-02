<script lang="ts">
  import { X } from "lucide-svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
  }: {
    filterGroups: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    onClearAllFilters?: () => void;
  } = $props();

  const appliedFilters = $derived(
    filterGroups
      .filter((g) => g.selected !== g.defaultValue)
      .map((g) => ({
        key: g.key,
        defaultValue: g.defaultValue,
        label:
          g.options.find((o) => o.value === g.selected)?.label ?? g.selected,
      })),
  );
</script>

{#if appliedFilters.length > 0}
  <div
    class="flex flex-row items-center justify-between gap-x-2 h-9 border-t"
  >
    <div class="flex flex-row items-center gap-2 flex-wrap">
      {#each appliedFilters as filter (filter.key)}
        <span
          class="inline-flex items-center gap-x-1 h-7 px-2 rounded-sm border bg-white text-xs font-medium text-fg-primary"
        >
          {filter.label}
          <button
            class="text-fg-secondary hover:text-fg-primary shrink-0 cursor-pointer"
            onclick={() => onFilterChange?.(filter.key, filter.defaultValue)}
            aria-label="Remove filter {filter.label}"
          >
            <X size={12} />
          </button>
        </span>
      {/each}
    </div>
    <button
      class="text-sm text-fg-secondary hover:text-fg-primary whitespace-nowrap cursor-pointer"
      onclick={onClearAllFilters}
    >
      Clear all
    </button>
  </div>
{/if}
