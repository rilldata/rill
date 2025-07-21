<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { FieldConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import { List } from "lucide-svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let sortConfig: ChartFieldInput["sortSelector"];

  let isCustomSortDropdownOpen = false;

  const sortOptions = [
    { label: "Ascending", value: "x" },
    { label: "Descending", value: "-x" },
    { label: "Y-axis ascending", value: "y" },
    { label: "Y-axis descending", value: "-y" },
    { label: "Custom", value: "custom" },
  ];

  $: sortValue = fieldConfig.sort
    ? typeof fieldConfig.sort === "string"
      ? fieldConfig.sort
      : "custom"
    : (sortConfig?.defaultSort ?? "x");

  $: customSortDraggableItems = sortConfig?.customSortItems?.map((item) => ({
    id: item,
    value: item,
  }));

  function handleReorder(
    event: CustomEvent<{
      items: { id: string; value: string }[];
      fromIndex: number;
      toIndex: number;
    }>,
  ) {
    const reorderedItems = event.detail.items.map((item) => item.value);
    onChange("sort", reorderedItems);
  }

  function handleSortChange(e: CustomEvent<string>) {
    const newSortValue = e.detail;
    if (newSortValue === "custom") {
      onChange("sort", sortConfig?.customSortItems);
    } else {
      onChange("sort", newSortValue);
    }
  }
</script>

{#if sortConfig?.enable}
  <div class="py-1 flex items-center justify-between">
    <span class="text-xs">Sort</span>
    <div class="flex items-center gap-x-1">
      <Select
        size="sm"
        id="sort-select"
        width={190}
        options={sortOptions}
        value={sortValue}
        on:change={handleSortChange}
      />
      {#if sortValue === "custom"}
        <Popover.Root bind:open={isCustomSortDropdownOpen}>
          <Popover.Trigger>
            <IconButton rounded active={isCustomSortDropdownOpen}>
              <List size="14px" />
            </IconButton>
          </Popover.Trigger>
          <Popover.Content align="end" class="w-[240px] p-0">
            <div class="px-3 py-2 border-b border-gray-200">
              <span class="text-xs font-medium">Sort Order</span>
            </div>
            <DraggableList
              items={customSortDraggableItems || []}
              on:reorder={handleReorder}
              minHeight="auto"
              maxHeight="300px"
            >
              <div slot="empty" class="px-2 py-2 text-xs text-gray-500">
                No sort item found
              </div>
              <div slot="item" let:item class="flex items-center gap-x-1">
                <DragHandle size="16px" className="text-gray-400" />
                <span class="text-xs truncate">{item.value}</span>
              </div>
            </DraggableList>
          </Popover.Content>
        </Popover.Root>
      {/if}
    </div>
  </div>
{/if}
