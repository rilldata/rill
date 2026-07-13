<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { ArrowDownWideNarrow, ArrowUpWideNarrow } from "lucide-svelte";
  import type { SortOption } from "./types";
  import { RecordRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  type SortSize = "sm" | "lg";

  let {
    sortStore,
    sortOptions,
    size = "lg",
    noOutline = false,
  }: {
    sortStore: RecordRuneStore<boolean>;
    sortOptions: SortOption[];
    size?: SortSize;
    noOutline?: boolean;
  } = $props();

  let sortLabel = $derived(
    sortOptions.find((o) => o.value in sortStore.value)?.label ?? "None",
  );

  const ClassForSize: Record<SortSize, string> = {
    sm: "h-7 px-2",
    lg: "h-9 px-4",
  };
  let sizeClass = $derived(ClassForSize[size]);

  let outlineClass = $derived(
    noOutline ? "" : "border rounded-[2px] shadow-xs bg-white",
  );

  function toggleSortDirection(e: MouseEvent, key: string, curDir: boolean) {
    e.preventDefault();
    e.stopPropagation();
    sortStore.set(key, !curDir);
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="flex flex-row items-center gap-x-1.5 text-sm font-medium text-fg-primary hover:bg-surface-hover cursor-pointer {sizeClass} {outlineClass}"
    aria-label="Sort order: {sortLabel}"
  >
    <span>Sort by {sortLabel}</span>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#each sortOptions as option (option.value)}
      {@const checked = Object.hasOwn(sortStore.value, option.value)}
      {@const sortDirection = sortStore.value[option.value] ?? true}
      <DropdownMenu.CheckboxItem
        closeOnSelect
        {checked}
        onCheckedChange={() =>
          sortStore.set(option.value, checked ? null : sortDirection)}
      >
        <span class="grow">{option.label}</span>
        <button
          onclick={(e) => toggleSortDirection(e, option.value, sortDirection)}
        >
          {#if sortDirection}
            <ArrowDownWideNarrow size="16px" />
          {:else}
            <ArrowUpWideNarrow size="16px" />
          {/if}
        </button>
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
