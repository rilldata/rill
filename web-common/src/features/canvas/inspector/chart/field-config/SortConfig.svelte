<script lang="ts">
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
  import type {
    ChartSortDirectionOptions,
    FieldConfig,
  } from "@rilldata/web-common/features/components/charts/types";
  import { List } from "lucide-svelte";

  export let fieldConfig: FieldConfig;
  export let onChange: (property: keyof FieldConfig, value: any) => void;
  export let sortConfig: ChartFieldInput["sortSelector"];

  let isCustomSortDropdownOpen = false;

  const sortOptions: { label: string; value: ChartSortDirectionOptions }[] = [
    { label: "X-axis ascending", value: "x" },
    { label: "X-axis descending", value: "-x" },
    { label: "Y-axis ascending", value: "y" },
    { label: "Y-axis descending", value: "-y" },
    { label: "Color ascending", value: "color" },
    { label: "Color descending", value: "-color" },
    { label: "Measure ascending", value: "measure" },
    { label: "Measure descending", value: "-measure" },
    { label: "Custom", value: "custom" },
  ];

  $: sortOptionsForChart = sortOptions.filter((option) =>
    sortConfig?.options?.includes(option.value),
  );

  $: sortValue = fieldConfig.sort
    ? typeof fieldConfig.sort === "string"
      ? fieldConfig.sort
      : "custom"
    : (sortConfig?.defaultSort ?? "x");

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
        options={sortOptionsForChart}
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
              onReorder={handleReorder}
              minHeight="auto"
              maxHeight="300px"
            >
              <div slot="empty" class="px-2 py-2 text-xs text-gray-500">
                No sort item found
              </div>
              <div slot="item" let:item class="flex items-center gap-x-1">
                <DragHandle
                  size="16px"
                  className="text-gray-400 pointer-events-none"
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
