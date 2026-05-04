<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FilterOutlined from "@rilldata/web-common/components/icons/FilterOutlined.svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
    onFilterChange,
  }: {
    filterGroups: FilterGroup[];
    onFilterChange?: (key: string, selected: string | string[]) => void;
  } = $props();

  let hasActiveFilters = $derived(
    filterGroups.some((g) => {
      if (g.multiSelect && Array.isArray(g.selected)) {
        return g.selected.length > 0;
      }
      return (
        typeof g.selected === "string" &&
        g.selected !== "" &&
        g.selected !== g.defaultValue
      );
    }),
  );

  function handleClick(group: FilterGroup, value: string) {
    if (group.multiSelect) {
      const current = Array.isArray(group.selected) ? group.selected : [];
      const next = current.includes(value)
        ? current.filter((v) => v !== value)
        : [...current, value];
      onFilterChange?.(group.key, next);
    } else {
      onFilterChange?.(group.key, value);
    }
  }
</script>

{#if filterGroups.length > 0}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger
      class="flex flex-row items-center gap-x-2 h-9 px-4 border rounded-[2px] shadow-xs text-sm font-medium cursor-pointer {hasActiveFilters
        ? 'bg-surface-hover border-border text-fg-accent'
        : 'bg-white border-border text-fg-primary hover:bg-surface-hover'}"
      aria-label="Filter options"
    >
      <FilterOutlined size="16" />
      <span>Filter</span>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#each filterGroups as group, i (group.key)}
        <DropdownMenu.Group>
          <DropdownMenu.Label class="uppercase"
            >{group.label}</DropdownMenu.Label
          >
          {#each group.options as option}
            <DropdownMenu.CheckboxItem
              closeOnSelect={!group.multiSelect}
              checked={group.multiSelect
                ? Array.isArray(group.selected) &&
                  group.selected.includes(option.value)
                : group.selected === option.value}
              onclick={() => handleClick(group, option.value)}
            >
              {option.label}
            </DropdownMenu.CheckboxItem>
          {/each}
        </DropdownMenu.Group>
        {#if i < filterGroups.length - 1}
          <DropdownMenu.Separator />
        {/if}
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
