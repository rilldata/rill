<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import type { FilterGroup } from "./types";

  let {
    filterGroups = [],
    onFilterChange,
  }: {
    filterGroups: FilterGroup[];
    onFilterChange?: (key: string, value: string) => void;
  } = $props();

  let isOpen = $state(false);
</script>

{#if filterGroups.length > 0}
  <DropdownMenu.Root bind:open={isOpen}>
    <DropdownMenu.Trigger
      class="flex flex-row gap-1.5 items-center rounded-sm border bg-input px-2.5 py-1.5 {isOpen
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'}"
    >
      <Filter size="14px" />
      <span class="text-sm text-fg-secondary font-medium">Filter</span>
      {#if isOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content align="start">
      {#each filterGroups as group}
        <DropdownMenu.Label class="uppercase">{group.label}</DropdownMenu.Label>
        {#each group.options as option}
          <DropdownMenu.CheckboxItem
            checked={group.selected === option.value}
            onclick={() => onFilterChange?.(group.key, option.value)}
          >
            {option.label}
          </DropdownMenu.CheckboxItem>
        {/each}
        {#if filterGroups.indexOf(group) < filterGroups.length - 1}
          <DropdownMenu.Separator />
        {/if}
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
