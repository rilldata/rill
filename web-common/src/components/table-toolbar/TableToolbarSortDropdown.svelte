<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import type { SortDirection } from "./types";

  let {
    sortDirection = "newest",
    onSortChange,
  }: {
    sortDirection: SortDirection;
    onSortChange?: (direction: SortDirection) => void;
  } = $props();

  let isOpen = $state(false);

  const sortLabel = $derived(
    sortDirection === "newest" ? "Newest" : "Oldest",
  );
</script>

<DropdownMenu.Root bind:open={isOpen}>
  <DropdownMenu.Trigger
    class="flex flex-row gap-1.5 items-center rounded-sm border bg-input px-2.5 py-1.5 {isOpen
      ? 'bg-gray-200'
      : 'hover:bg-surface-hover'}"
  >
    <span class="text-sm text-fg-secondary font-medium">{sortLabel}</span>
    {#if isOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="end">
    <DropdownMenu.RadioGroup value={sortDirection}>
      <DropdownMenu.RadioItem
        value="newest"
        onclick={() => onSortChange?.("newest")}
      >
        Newest
      </DropdownMenu.RadioItem>
      <DropdownMenu.RadioItem
        value="oldest"
        onclick={() => onSortChange?.("oldest")}
      >
        Oldest
      </DropdownMenu.RadioItem>
    </DropdownMenu.RadioGroup>
  </DropdownMenu.Content>
</DropdownMenu.Root>
