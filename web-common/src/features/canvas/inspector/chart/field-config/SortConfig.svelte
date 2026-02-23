<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import {
    ChartSortType,
    type ChartSortDirectionOptions,
    type FieldConfig,
  } from "@rilldata/web-common/features/components/charts/types";
  import { List } from "lucide-svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let sortConfig: ChartFieldInput["sortSelector"];

  let isCustomSortDropdownOpen = false;

  const sortOptions: { label: string; value: ChartSortDirectionOptions }[] = [
    { label: "X-axis ascending", value: ChartSortType.X_ASC },
    { label: "X-axis descending", value: ChartSortType.X_DESC },
    { label: "Y-axis ascending", value: ChartSortType.Y_ASC },
    { label: "Y-axis descending", value: ChartSortType.Y_DESC },
    { label: "Y-axis delta ascending", value: ChartSortType.Y_DELTA_ASC },
    { label: "Y-axis delta descending", value: ChartSortType.Y_DELTA_DESC },
    { label: "Color ascending", value: ChartSortType.COLOR_ASC },
    { label: "Color descending", value: ChartSortType.COLOR_DESC },
    { label: "Measure ascending", value: ChartSortType.MEASURE_ASC },
    { label: "Measure descending", value: ChartSortType.MEASURE_DESC },
    { label: "Custom", value: ChartSortType.CUSTOM },
  ];

  $: sortOptionsForChart = sortOptions.filter((option) =>
    sortConfig?.options?.includes(option.value),
  );

  $: sortValue = fieldConfig.sort
    ? typeof fieldConfig.sort === "string"
      ? fieldConfig.sort
      : ChartSortType.CUSTOM
    : (sortConfig?.defaultSort ?? ChartSortType.X_ASC);

  $: customSortDraggableItems = sortConfig?.customSortItems?.map((item) => ({
    id: sanitizeItemId(item),
    value: item,
  }));

  function sanitizeItemId(item: string | null) {
    if (item === null) return "null-item";
    if (item === "") return "<empty-string>";
    return item;
  }

  function handleReorder(data: {
    items: { id: string; value: string }[];
    fromIndex: number;
    toIndex: number;
  }) {
    const reorderedItems = data.items.map((item) => item.value);
    onChange("sort", reorderedItems);
  }

  function handleSortChange(newSortValue: string) {
    if (newSortValue === ChartSortType.CUSTOM) {
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
        options={sortOptionsForChart}
        value={sortValue}
        onChange={handleSortChange}
      />
      {#if sortValue === ChartSortType.CUSTOM}
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
              onReorder={handleReorder}
              minHeight="auto"
              maxHeight="300px"
            >
              <div slot="empty" class="px-2 py-2 text-xs text-fg-secondary">
                No sort item found
              </div>
              <div slot="item" let:item class="flex items-center gap-x-1">
                <DragHandle
                  size="16px"
                  className="text-fg-secondary pointer-events-none"
                />
                <span class="text-xs truncate">{item.value}</span>
              </div>
            </DraggableList>
          </Popover.Content>
        </Popover.Root>
      {/if}
    </div>
  </div>
{/if}
