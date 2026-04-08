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

  interface AppliedChip {
    key: string;
    resetValue: string;
    label: string;
  }

  const appliedFilters = $derived(
    filterGroups.flatMap((g): AppliedChip[] => {
      if (g.multiSelect && Array.isArray(g.selected)) {
        return g.selected.map((val) => ({
          key: g.key,
          resetValue: val,
          label: g.options.find((o) => o.value === val)?.label ?? val,
        }));
      }
      if (typeof g.selected === "string" && g.selected !== g.defaultValue) {
        return [
          {
            key: g.key,
            resetValue: g.defaultValue as string,
            label:
              g.options.find((o) => o.value === g.selected)?.label ??
              (g.selected as string),
          },
        ];
      }
      return [];
    }),
  );
</script>

{#if appliedFilters.length > 0}
  <hr class="border-t" />
  <div class="flex flex-row items-center justify-between gap-x-2 h-9">
    <div class="flex flex-row items-center gap-2 flex-wrap">
      {#each appliedFilters as filter (`${filter.key}:${filter.resetValue}`)}
        <span
          class="inline-flex items-center gap-x-1 h-7 px-2 rounded-sm border bg-surface-background text-xs font-medium text-fg-primary"
        >
          {filter.label}
          <button
            type="button"
            class="text-fg-secondary hover:text-fg-primary shrink-0 cursor-pointer"
            onclick={() => onFilterChange?.(filter.key, filter.resetValue)}
            aria-label="Remove filter {filter.label}"
          >
            <X size={12} />
          </button>
        </span>
      {/each}
    </div>
    <button
      type="button"
      class="text-sm text-fg-secondary hover:text-fg-primary whitespace-nowrap cursor-pointer"
      onclick={onClearAllFilters}
    >
      Clear all
    </button>
  </div>
{/if}
