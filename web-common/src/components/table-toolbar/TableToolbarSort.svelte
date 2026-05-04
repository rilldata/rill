<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ArrowUpDown } from "lucide-svelte";
  import { SORT_OPTIONS, type SortDirection } from "./types";

  let {
    sortDirection = $bindable("newest"),
  }: {
    sortDirection: SortDirection;
  } = $props();

  const sortLabel = $derived(
    SORT_OPTIONS.find((o) => o.value === sortDirection)?.label ?? "Newest",
  );
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="flex flex-row items-center gap-x-1.5 h-9 px-4 border rounded-[2px] shadow-xs bg-white text-sm font-medium text-fg-primary hover:bg-surface-hover cursor-pointer"
    aria-label="Sort order: {sortLabel}"
  >
    <ArrowUpDown size={16} />
    <span>Sort by {sortLabel}</span>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start" sameWidth>
    {#each SORT_OPTIONS as option (option.value)}
      <DropdownMenu.CheckboxItem
        closeOnSelect
        checked={sortDirection === option.value}
        onclick={() => (sortDirection = option.value)}
      >
        {option.label}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
