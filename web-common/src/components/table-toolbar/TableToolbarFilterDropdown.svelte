<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FilterOutlined from "@rilldata/web-common/components/icons/FilterOutlined.svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
    onFilterChange,
    disabled = false,
  }: {
    filterGroups: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
    disabled?: boolean;
  } = $props();
</script>

{#if filterGroups.length > 0}
  <DropdownMenu.Root>
    <DropdownMenu.Trigger
      class="flex flex-row items-center gap-x-1.5 h-9 px-2 text-sm font-medium text-fg-primary cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
      aria-label="Filter options"
      {disabled}
    >
      <FilterOutlined size="14" />
      <span>Filter</span>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#each filterGroups as group, i}
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
              onclick={() => onFilterChange?.(group.key, option.value)}
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
