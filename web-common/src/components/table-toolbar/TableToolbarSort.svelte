<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SortOption } from "./types";
  import { type RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

  type SortSize = "sm" | "lg";

  let {
    sortStore,
    sortOptions,
    size = "lg",
    noOutline = false,
  }: {
    sortStore: RuneStore<string>;
    sortOptions: SortOption[];
    size?: SortSize;
    noOutline?: boolean;
  } = $props();

  let open = $state(false);

  let sortLabel = $derived(
    "Sort by " +
      (sortOptions.find((o) => o.id === sortStore.value)?.label ?? "None"),
  );

  const ClassForSize: Record<SortSize, string> = {
    sm: "h-7 px-2",
    lg: "h-9 px-4",
  };
  let sizeClass = $derived(ClassForSize[size]);

  let outlineClass = $derived(
    noOutline ? "" : "border rounded-[2px] shadow-xs bg-white",
  );
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger
    class="flex flex-row items-center gap-x-1.5 text-sm font-medium text-fg-primary hover:bg-surface-hover cursor-pointer {sizeClass} {outlineClass}"
    aria-label={sortLabel}
  >
    <span>{sortLabel}</span>
  </DropdownMenu.Trigger>
  <DropdownMenu.Content align="start">
    {#each sortOptions as option (option.id)}
      <DropdownMenu.CheckboxItem
        closeOnSelect
        checked={sortStore.value === option.id}
        onCheckedChange={() => sortStore.setter(option.id)}
      >
        {option.label}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.Content>
</DropdownMenu.Root>
