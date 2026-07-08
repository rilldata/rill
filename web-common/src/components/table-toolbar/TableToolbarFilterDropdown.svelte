<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import FilterOutlined from "@rilldata/web-common/components/icons/FilterOutlined.svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
  }: {
    filterGroups: FilterGroup[];
  } = $props();

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
      class="flex flex-row items-center gap-x-1.5 h-9 px-2 text-sm font-medium text-fg-primary cursor-pointer"
      aria-label="Filter options"
    >
      <FilterOutlined size="14" />
      <span>Filter</span>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#each filterGroups as group, i (group.key)}
        <DropdownMenu.Group>
          <DropdownMenu.Label class="uppercase"
            >{group.label}</DropdownMenu.Label
          >
          {#each group.options as option (option.value)}
            <DropdownMenu.CheckboxItem
              closeOnSelect={!group.multiSelect}
              checked={group.multiSelect
                ? group.selected.includes(option.value)
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
