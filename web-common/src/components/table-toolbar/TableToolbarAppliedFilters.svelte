<script lang="ts">
  import { X } from "lucide-svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
    onFilterChange,
    onClearAllFilters,
  }: {
    filterGroups: FilterGroup[];
    onFilterChange?: (key: string, selected: string | string[]) => void;
    onClearAllFilters?: () => void;
  } = $props();

  interface AppliedChip {
    key: string;
    value: string;
    label: string;
  }

  const appliedFilters = $derived(
    filterGroups.flatMap((g): AppliedChip[] => {
      if (g.multiSelect && Array.isArray(g.selected)) {
        return g.selected.map((val) => ({
          key: g.key,
          value: val,
          label: g.options.find((o) => o.value === val)?.label ?? val,
        }));
      }
      if (
        typeof g.selected === "string" &&
        g.selected &&
        g.selected !== g.defaultValue
      ) {
        return [
          {
            key: g.key,
            value: g.selected,
            label:
              g.options.find((o) => o.value === g.selected)?.label ??
              g.selected,
          },
        ];
      }
      return [];
    }),
  );

  let hasFilters = $derived(appliedFilters.length > 0);

  function handleDelete(key: string, value: string) {
    const group = filterGroups.find((g) => g.key === key);
    if (!group) return;
    if (group.multiSelect) {
      const current = Array.isArray(group.selected) ? group.selected : [];
      const next = current.filter((v) => v !== value);
      onFilterChange?.(group.key, next);
    } else {
      onFilterChange?.(group.key, group.defaultValue);
    }
  }
</script>

<div class="applied-filters-wrapper" class:open={hasFilters}>
  <div class="applied-filters-inner">
    {#if hasFilters}
      <hr class="border-t mt-2" />
      <div class="flex flex-row items-center justify-between gap-x-2 h-9">
        <div class="flex flex-row items-center gap-2 flex-wrap">
          {#each appliedFilters as filter (`${filter.key}:${filter.value}`)}
            <span
              class="inline-flex items-center gap-x-1 h-7 px-2 rounded-sm border bg-surface-background text-xs font-medium text-fg-primary"
            >
              {filter.label}
              <button
                type="button"
                class="text-fg-secondary hover:text-fg-primary shrink-0 cursor-pointer"
                onclick={() => handleDelete(filter.key, filter.value)}
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
  </div>
</div>

<style>
  .applied-filters-wrapper {
    display: grid;
    grid-template-rows: 0fr;
    transition: grid-template-rows 180ms ease-out;
  }

  .applied-filters-wrapper.open {
    grid-template-rows: 1fr;
  }

  .applied-filters-inner {
    overflow: hidden;
  }
</style>
