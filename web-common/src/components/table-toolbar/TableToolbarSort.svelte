<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import type { SortOption } from "./types";
  import { type RuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
    m.common_sort_by({
      label:
        sortOptions.find((o) => o.value === sortStore.value)?.label ??
        m.common_none(),
    }),
  );

  const ClassForSize: Record<SortSize, string> = {
    sm: "h-7 px-2",
    lg: "h-9 px-4",
  };
  let sizeClass = $derived(ClassForSize[size]);

  let outlineClass = $derived(
    noOutline ? "" : "border rounded-[2px] shadow-xs",
  );
</script>

<DropdownMenu.Root bind:open>
  <DropdownMenu.Trigger
    class="flex flex-row items-center gap-x-1.5 text-sm font-medium text-fg-primary hover:bg-surface-hover cursor-pointer {sizeClass} {outlineClass}"
    aria-label={sortLabel}
  >
    {sortLabel}
    <div class="caret transition-transform">
      <CaretDownIcon size="12px" className="fill-fg-secondary" />
    </div>
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

<style lang="postcss">
  .caret {
    @apply transition-transform text-fg-muted;
  }

  .caret:global([data-state="open"]) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
