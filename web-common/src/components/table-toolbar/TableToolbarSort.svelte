<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ArrowUpDown } from "lucide-svelte";
  import type { SortOption } from "./types";
  import type { RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  let {
    sortStore,
    sortOptions,
  }: {
    sortStore: RuneStore<string>;
    sortOptions: SortOption[];
  } = $props();

  let sortLabel = $derived(
    sortOptions.find((o) => o.value === sortStore.value)?.label ?? "None",
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
    {#each sortOptions as option (option.value)}
      <DropdownMenu.CheckboxItem
        closeOnSelect
        checked={sortStore.value === option.value}
        onCheckedChange={() => sortStore.setter(option.value)}
      >
        {option.label}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
