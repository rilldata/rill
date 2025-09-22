<script lang="ts">
  import DraggableList from "@rilldata/web-common/components/draggable-list";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import * as Popover from "@rilldata/web-common/components/popover";
  import type {
    MetricsViewSpecDimension,
    MetricsViewSpecMeasure,
  } from "@rilldata/web-common/runtime-client";
  import { Tooltip } from "bits-ui";
  import { Button } from "../button";
  import Search from "../search/Search.svelte";

  type SelectableItem = MetricsViewSpecMeasure | MetricsViewSpecDimension;

  export let selectedItems: string[];
  export let allItems: SelectableItem[] = [];
  export let type: "measure" | "dimension" = "measure";
  export let onSelectedChange: (items: string[]) => void;

  let searchText = "";
  let active = false;

  $: allItemsMap = new Map(allItems.map((item) => [item.name, item]));
  $: numAvailable = allItems?.length ?? 0;
  $: numShown = selectedItems?.filter((x) => x).length ?? 0;
  $: numShownString =
    numAvailable === numShown ? "All" : `${numShown} of ${numAvailable}`;
  $: tooltipText = `Choose ${type === "measure" ? "measures" : "dimensions"} to display`;

  // Convert selectedItems to DraggableItem format
  $: selectedDraggableItems = selectedItems
    .filter((id) => id)
    .map((id) => ({ id }));

  // Convert hidden items to DraggableItem format
  $: hiddenDraggableItems = Array.from(allItemsMap.entries())
    .filter(([id]) => id && !selectedItems.includes(id))
    .map(([id]) => ({ id: id! }));

  function handleSelectedReorder(data: {
    items: Array<{ id: string }>;
    fromIndex: number;
    toIndex: number;
  }) {
    const newSelectedItems = data.items.map((item) => item.id);
    onSelectedChange(newSelectedItems);
  }

  function handleHiddenItemClick(data: {
    item: { id: string };
    index: number;
  }) {
    const newSelectedItems = [...selectedItems, data.item.id];
    onSelectedChange(newSelectedItems);
  }

  function removeSelectedItem(id: string) {
    const newSelectedItems = selectedItems.filter((i) => i !== id);
    onSelectedChange(newSelectedItems);
  }

  function showAllItems() {
    const newSelectedItems = allItems.map((item) => item.name ?? "");
    onSelectedChange(newSelectedItems);
  }

  function hideAllItems() {
    const newSelectedItems = [selectedItems[0]];
    onSelectedChange(newSelectedItems);
  }
</script>

<Popover.Root bind:open={active}>
  <Popover.Trigger asChild let:builder>
    <Button builders={[builder]} type="text" theme label={tooltipText}>
      <div class="flex items-center gap-x-0.5 px-1">
        <strong
          >{`${numShownString} ${type === "measure" ? "Measures" : "Dimensions"}`}</strong
        >
        <span class="transition-transform" class:-rotate-180={active}>
          <CaretDownIcon />
        </span>
      </div>
    </Button>
  </Popover.Trigger>
  <Popover.Content
    class="p-0 z-popover"
    align="start"
    strategy="absolute"
    fitViewport={true}
    overflowY="auto"
    overflowX="hidden"
    minHeight="100px"
  >
    <div class="flex flex-col relative">
      <div class="px-3 pt-3 pb-0">
        <Search
          bind:value={searchText}
          label="Search list"
          showBorderOnFocus={false}
        />
      </div>

      <!-- Shown Items Section -->
      <div class="shown-section flex-1 p-1.5 pt-0" data-testid="shown-section">
        <DraggableList
          items={selectedDraggableItems}
          bind:searchValue={searchText}
          minHeight="auto"
          maxHeight="300px"
          onReorder={handleSelectedReorder}
        >
          <div
            slot="header"
            class="flex-none flex w-full py-1.5 pb-1 justify-between px-2 sticky top-0 from-popover from-80% to-transparent bg-gradient-to-b z-10"
          >
            <h3 class="uppercase text-gray-500 font-semibold text-[11px]">
              Shown {type === "measure" ? "Measures" : "Dimensions"}
            </h3>
            {#if selectedItems.length > 1}
              <button
                class="text-theme-500 pointer-events-auto hover:text-theme-600 font-medium text-[11px]"
                on:click={hideAllItems}
              >
                Hide all
              </button>
            {/if}
          </div>

          <div slot="empty" class="px-2 py-2 text-xs text-gray-500">
            {searchText
              ? `No matching ${type === "measure" ? "measures" : "dimensions"} shown`
              : `No ${type === "measure" ? "measures" : "dimensions"} shown`}
          </div>

          <div slot="item" let:item class="w-full flex gap-x-1 items-center">
            {@const itemData = allItemsMap.get(item.id)}
            {#if itemData?.description || selectedItems.length === 1}
              <Tooltip.Root openDelay={200} portal="body">
                <Tooltip.Trigger class="w-full flex gap-x-1 items-center">
                  <DragHandle
                    size="16px"
                    className="text-gray-400 pointer-events-none"
                  />
                  <span class="truncate flex-1 text-left pointer-events-none">
                    {itemData?.displayName ??
                      `Unknown ${type === "measure" ? "measure" : "dimension"}`}
                  </span>
                  <button
                    class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                    on:click|stopPropagation={() => removeSelectedItem(item.id)}
                    on:mousedown|stopPropagation={() => {
                      // NO-OP
                    }}
                    disabled={selectedItems.length === 1}
                    class:pointer-events-none={selectedItems.length === 1}
                    class:opacity-50={selectedItems.length === 1}
                    aria-label="Toggle visibility"
                    data-testid="toggle-visibility-button"
                  >
                    <EyeIcon size="14px" color="#6b7280" />
                  </button>
                </Tooltip.Trigger>
                <Tooltip.Content side="right" sideOffset={12} class="z-popover">
                  <div
                    class="bg-gray-800 text-gray-50 rounded p-2 pt-1 pb-1 shadow-md pointer-events-none z-50"
                  >
                    {#if selectedItems.length === 1}
                      Must show at least one {type === "measure"
                        ? "measure"
                        : "dimension"}
                    {:else}
                      {itemData?.description}
                    {/if}
                  </div>
                </Tooltip.Content>
              </Tooltip.Root>
            {:else}
              <DragHandle
                size="16px"
                className="text-gray-400 pointer-events-none"
              />
              <span class="truncate flex-1 text-left pointer-events-none">
                {itemData?.displayName ??
                  `Unknown ${type === "measure" ? "measure" : "dimension"}`}
              </span>
              <button
                class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                on:click|stopPropagation={() => removeSelectedItem(item.id)}
                on:mousedown|stopPropagation={() => {
                  // NO-OP
                }}
                disabled={selectedItems.length === 1}
                class:pointer-events-none={selectedItems.length === 1}
                class:opacity-50={selectedItems.length === 1}
              >
                <EyeIcon size="14px" color="#6b7280" />
              </button>
            {/if}
          </div>
        </DraggableList>
      </div>

      <!-- Hidden Items Section -->
      {#if selectedItems.length < allItems.length}
        <span class="flex-none h-px bg-slate-200 w-full" />
        <div class="hidden-section flex flex-col flex-1 min-h-0 p-1.5 pt-0">
          <DraggableList
            items={hiddenDraggableItems}
            bind:searchValue={searchText}
            minHeight="auto"
            maxHeight="200px"
            onItemClick={handleHiddenItemClick}
          >
            <div
              slot="header"
              class="flex-none flex py-1.5 justify-between px-2 sticky top-0 from-popover from-80% to-transparent bg-gradient-to-b"
            >
              <h3
                class="uppercase text-xs text-gray-500 font-semibold from-popover from-80% to-transparent bg-gradient-to-b"
              >
                Hidden {type === "measure" ? "Measures" : "Dimensions"}
              </h3>
              <button
                class="pointer-events-auto text-theme-500 text-[11px] font-medium"
                on:click={showAllItems}
              >
                Show all
              </button>
            </div>

            <div slot="empty" class="px-2 py-2 text-xs text-gray-500">
              {searchText
                ? `No matching hidden ${type === "measure" ? "measures" : "dimensions"}`
                : `No hidden ${type === "measure" ? "measures" : "dimensions"}`}
            </div>

            <div
              slot="item"
              let:item
              let:index
              class="w-full flex gap-x-1 justify-between items-center cursor-pointer"
            >
              {@const itemData = allItemsMap.get(item.id)}
              {#if itemData?.description}
                <Tooltip.Root openDelay={200} portal="body">
                  <Tooltip.Trigger
                    class="w-full flex gap-x-1 justify-between items-center"
                  >
                    <span class="truncate flex-1 text-left pointer-events-none">
                      {itemData.displayName}
                    </span>
                    <button
                      class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                      on:click|stopPropagation={() =>
                        handleHiddenItemClick({ item, index })}
                      aria-label="Toggle visibility"
                      data-testid="toggle-visibility-button"
                    >
                      <EyeOffIcon size="14px" color="#9ca3af" />
                    </button>
                  </Tooltip.Trigger>
                  <Tooltip.Content
                    side="right"
                    sideOffset={12}
                    class="z-popover"
                  >
                    <div
                      class="bg-gray-800 text-gray-50 rounded p-2 pt-1 pb-1 shadow-md pointer-events-none z-50"
                    >
                      {itemData.description}
                    </div>
                  </Tooltip.Content>
                </Tooltip.Root>
              {:else}
                <span class="truncate flex-1 text-left pointer-events-none">
                  {itemData?.displayName}
                </span>
                <button
                  class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                  on:click|stopPropagation={() =>
                    handleHiddenItemClick({ item, index })}
                  aria-label="Toggle visibility"
                  data-testid="toggle-visibility-button"
                >
                  <EyeOffIcon size="14px" color="#9ca3af" />
                </button>
              {/if}
            </div>
          </DraggableList>
        </div>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>

<style lang="postcss">
  h3 {
    @apply text-[11px] text-gray-500;
  }
</style>
