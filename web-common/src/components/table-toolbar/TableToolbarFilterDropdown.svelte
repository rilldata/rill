<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FilterOutlined from "@rilldata/web-common/components/icons/FilterOutlined.svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
  }: {
    filterGroups: FilterGroup[];
  } = $props();

  let hasActiveFilters = $derived(
    filterGroups.some((g) => {
      if (g.multiSelect) {
        return g.selectedStore.value.length > 0;
      }
      return (
        g.selectedStore.value !== "" && g.selectedStore.value !== g.defaultValue
      );
    }),
  );

  function handleClick(group: FilterGroup, value: string) {
    if (group.multiSelect) {
      group.selectedStore.toggle(value);
    } else {
      group.selectedStore.setter(value);
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
    <DropdownMenu.Content align="start" class="max-h-72 overflow-y-auto">
      {#each filterGroups as group, i (group.key)}
        <DropdownMenu.Group>
          <DropdownMenu.Label class="uppercase">
            {group.label}
          </DropdownMenu.Label>
          {#each group.options as option (option.value)}
            <DropdownMenu.CheckboxItem
              closeOnSelect={!group.multiSelect}
              checked={group.multiSelect
                ? group.selectedStore.value.includes(option.value)
                : group.selectedStore.value === option.value}
              onCheckedChange={() => handleClick(group, option.value)}
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
