<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    ArrowDownWideNarrow,
    ArrowUpWideNarrow,
    CheckIcon,
  } from "lucide-svelte";
  import type { SortOption } from "./types";
  import { ArrayRuneStore } from "@rilldata/web-common/lib/store-utils/types.svelte.ts";
  import { DraggableList } from "@rilldata/web-common/components/draggable-list";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import type { ColumnSort } from "tanstack-table-8-svelte-5";

  type SortSize = "sm" | "lg";

  let {
    sortStore,
    sortOptions,
    size = "lg",
    noOutline = false,
  }: {
    sortStore: ArrayRuneStore<ColumnSort>;
    sortOptions: SortOption[];
    size?: SortSize;
    noOutline?: boolean;
  } = $props();

  let open = $state(false);

  let selectedOptions = $derived(
    sortStore.value.map((s) => ({
      id: s.id,
      desc: s.desc,
      label: sortOptions.find((o) => o.id === s.id)?.label ?? s.id,
    })),
  );
  let unselectedOptions = $derived(
    sortOptions.filter((o) => !sortStore.value.find((s) => s.id === o.id)),
  );

  let sortLabel = $derived.by(() => {
    if (selectedOptions.length === 0) return "Select sort";

    const firstSortDir = getSortDirectionLabel(
      selectedOptions[0],
      sortStore.value,
    );
    const firstSortLabel = `Sort ${firstSortDir} by ${selectedOptions[0].label}`;
    if (selectedOptions.length === 1) return firstSortLabel;

    return `${firstSortLabel} +${selectedOptions.length - 1} others`;
  });

  const ClassForSize: Record<SortSize, string> = {
    sm: "h-7 px-2",
    lg: "h-9 px-4",
  };
  let sizeClass = $derived(ClassForSize[size]);

  let outlineClass = $derived(
    noOutline ? "" : "border rounded-[2px] shadow-xs bg-white",
  );

  function getSortDirectionLabel(option: SortOption, sorts: ColumnSort[]) {
    return sorts.find((s) => s.id === option.id)?.desc ? "desc" : "asc";
  }

  function toggleSortSelection(e: MouseEvent, item: any) {
    e.stopPropagation();
    sortStore.toggle({
      id: item.id,
      desc: item.desc ?? true,
    });
  }

  function toggleSortDirection(e: MouseEvent, key: string) {
    e.stopPropagation();

    const index = sortStore.value.findIndex((s) => s.id === key);
    if (index === -1) return;

    const newValue = [...sortStore.value];
    newValue[index].desc = !newValue[index].desc;
    sortStore.setter(newValue);
  }

  function handleReorder(newItems: any[]) {
    sortStore.setter(
      (newItems as ColumnSort[]).map((i) => ({
        id: i.id,
        desc: i.desc,
      })),
    );
  }
</script>

<Popover.Root bind:open>
  <Tooltip distance={8} suppress={open || selectedOptions.length <= 1}>
    <Popover.Trigger
      class="flex flex-row items-center gap-x-1.5 text-sm font-medium text-fg-primary hover:bg-surface-hover cursor-pointer {sizeClass} {outlineClass}"
      aria-label={sortLabel}
    >
      <span>{sortLabel}</span>
    </Popover.Trigger>
    <TooltipContent slot="tooltip-content">
      {#each selectedOptions as option (option.id)}
        {@const sortDir = getSortDirectionLabel(option, sortStore.value)}
        <div>{sortDir} by {option.label}</div>
      {/each}
    </TooltipContent>
  </Tooltip>
  <Popover.Content align="start">
    {#if selectedOptions.length}
      <DraggableList
        items={selectedOptions}
        minHeight="0px"
        draggable
        onReorder={({ items }) => handleReorder(items)}
      >
        {#snippet item({ item })}
          {@const sortDirection =
            sortStore.value.find((s) => s.id === item.id)?.desc ?? true}

          <DragHandle size="16px" className="fill-icon pointer-events-none" />

          <button
            class="grow text-left"
            onclick={(e) => toggleSortSelection(e, item)}
          >
            {item.label}
          </button>
          <button onclick={(e) => toggleSortDirection(e, item.id)}>
            {#if sortDirection}
              <ArrowDownWideNarrow size="16px" />
            {:else}
              <ArrowUpWideNarrow size="16px" />
            {/if}
          </button>
        {/snippet}
      </DraggableList>
    {/if}

    {#if selectedOptions.length && unselectedOptions.length}
      <div class="border-t"></div>
    {/if}

    {#if unselectedOptions.length}
      <div class="flex flex-col p-1.5 gap-y-2.5">
        {#each unselectedOptions as option (option.id)}
          <button
            onclick={(e) => toggleSortSelection(e, option)}
            class="text-left ml-5"
          >
            {option.label}
          </button>
        {/each}
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
