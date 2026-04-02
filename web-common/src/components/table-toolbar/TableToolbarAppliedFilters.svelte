<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
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
    class="flex flex-row items-center justify-between gap-x-2 border-t pt-2"
  >
    <div class="flex flex-row items-center gap-2 flex-wrap">
      {#each appliedFilters as filter (filter.key)}
        <Chip
          removable
          gray
          compact
          slideDuration={0}
          onRemove={() => onFilterChange?.(filter.key, filter.defaultValue)}
        >
          <span slot="body" class="text-xs">{filter.label}</span>
        </Chip>
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
