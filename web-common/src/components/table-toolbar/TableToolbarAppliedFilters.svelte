<script lang="ts">
  import { X } from "lucide-svelte";
  import type { FilterGroup } from "./types";
  import { slide } from "svelte/transition";

  let {
    filterGroups = [],
    onClearAllFilters,
  }: {
    filterGroups: FilterGroup[];
    onClearAllFilters?: () => void;
  } = $props();

  interface AppliedChip {
    key: string;
    value: string;
    label: string;
  }

  let appliedFilters = $derived(
    filterGroups.flatMap((g): AppliedChip[] => {
      if (g.multiSelect) {
        return g.selectedStore.value.map((val) => ({
          key: g.key,
          value: val,
          label: g.options.find((o) => o.value === val)?.label ?? val,
        }));
      }
      const selected = g.selectedStore.value;
      if (
        typeof selected === "string" &&
        selected &&
        selected !== g.defaultValue
      ) {
        return [
          {
            key: g.key,
            value: selected,
            label:
              g.options.find((o) => o.value === selected)?.label ?? selected,
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
      group.selectedStore.delete(value);
    } else {
      group.selectedStore.setter(group.defaultValue);
    }
  }
</script>

{#if hasFilters}
  <div class="overflow-hidden" in:slide out:slide>
    <div
      class="flex flex-row items-center justify-between gap-x-2.5 bg-surface-subtle rounded-b-sm p-2"
    >
      <div class="flex flex-row items-center gap-2.5 flex-1 min-w-0 flex-wrap">
        {#each appliedFilters as filter (`${filter.key}:${filter.value}`)}
          <span
            class="inline-flex items-center gap-x-1 h-7 px-2 rounded-sm border bg-surface-background shadow-xs text-sm font-medium text-fg-primary"
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
        class="text-sm font-medium text-fg-accent hover:text-fg-primary whitespace-nowrap cursor-pointer px-4 py-2 shrink-0"
        onclick={onClearAllFilters}
      >
        Clear all
      </button>
    </div>
  </div>
{/if}
