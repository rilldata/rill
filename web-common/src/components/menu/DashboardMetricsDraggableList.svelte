<script lang="ts">
  import * as Popover from "@rilldata/web-common/components/popover";
  import { Button } from "../button";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import EyeOffIcon from "@rilldata/web-common/components/icons/EyeInvisible.svelte";
  import EyeIcon from "@rilldata/web-common/components/icons/Eye.svelte";
  import Search from "../search/Search.svelte";
  import { Tooltip } from "bits-ui";
  import DraggableList from "../draggable-list/DraggableList.svelte";
  import type {
    MetricsViewSpecMeasure,
    MetricsViewSpecDimension,
  } from "@rilldata/web-common/runtime-client";

  const UPPER_BOUND = 12 + 28 + 25;
  const ITEM_HEIGHT = 28;
  const THROTTLE_MS = 16;

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

  $: filteredSelectedItems = searchText
    ? selectedItems.filter((id) => {
        const item = allItemsMap.get(id);
        return (
          item?.displayName?.toLowerCase().includes(searchText.toLowerCase()) ??
          false
        );
      })
    : selectedItems;

  $: filteredHiddenItems = searchText
    ? Array.from(allItemsMap.entries()).filter(
        ([id, item]) =>
          id &&
          !selectedItems.includes(id) &&
          (item.displayName?.toLowerCase().includes(searchText.toLowerCase()) ??
            false),
      )
    : Array.from(allItemsMap.entries()).filter(
        ([id]) => id && !selectedItems.includes(id),
      );

  function handleReorder(items: string[]) {
    onSelectedChange(items);
  }
</script>

<Popover.Root bind:open={active}>
  <Popover.Trigger asChild let:builder>
    <Button builders={[builder]} type="text" label={tooltipText} on:click>
      <div
        class="flex items-center gap-x-0.5 px-1 text-gray-700 hover:text-inherit"
      >
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

      <div class="shown-section flex flex-col flex-1 p-1.5 pt-0">
        <header
          class="flex-none flex w-full py-1.5 pb-1 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b z-10"
        >
          <h3 class="uppercase text-gray-500 font-semibold">
            Shown {type === "measure" ? "Measures" : "Dimensions"}
          </h3>
          {#if selectedItems.length > 1}
            <button
              class="text-primary-500 pointer-events-auto hover:text-primary-600 font-medium text-[11px]"
              on:click={() => {
                selectedItems = [selectedItems[0]];
                onSelectedChange(selectedItems);
              }}
            >
              Hide all
            </button>
          {/if}
        </header>

        {#if filteredSelectedItems.length === 0}
          <div class="px-2 py-2 text-xs text-gray-500">
            {searchText
              ? `No matching ${type === "measure" ? "measures" : "dimensions"} shown`
              : `No ${type === "measure" ? "measures" : "dimensions"} shown`}
          </div>
        {:else}
          <DraggableList
            items={filteredSelectedItems}
            onReorder={handleReorder}
            itemHeight={ITEM_HEIGHT}
            throttleMs={THROTTLE_MS}
            upperBound={UPPER_BOUND}
            disabled={selectedItems.length === 1}
          >
            {#each filteredSelectedItems as id, i (i)}
              {@const elementId = `visible-${type === "measure" ? "measures" : "dimensions"}-${id}`}
              {#if allItemsMap.get(id)?.description || selectedItems.length === 1}
                <Tooltip.Root openDelay={200} portal="body">
                  <Tooltip.Trigger>
                    <div
                      role="presentation"
                      data-testid={elementId}
                      id={elementId}
                      class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto items-center hover:bg-slate-50 rounded-sm"
                    >
                      <span
                        class="truncate flex-1 text-left pointer-events-none"
                        >{allItemsMap.get(id)?.displayName ??
                          `Unknown ${type === "measure" ? "measure" : "dimension"}`}</span
                      >

                      <button
                        class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                        on:click={() => {
                          selectedItems = selectedItems.filter(
                            (_, j) => j !== i,
                          );
                          onSelectedChange(selectedItems);
                        }}
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
                    </div>
                  </Tooltip.Trigger>

                  <Tooltip.Content
                    side="right"
                    sideOffset={12}
                    class="z-popover"
                  >
                    <div
                      class="bg-gray-800 text-gray-50 rounded p-2 pt-1 pb-1 shadow-md pointer-events-none z-50"
                    >
                      {#if selectedItems.length === 1}
                        Must show at least one {type === "measure"
                          ? "measure"
                          : "dimension"}
                      {:else}
                        {allItemsMap.get(id)?.description}
                      {/if}
                    </div>
                  </Tooltip.Content>
                </Tooltip.Root>
              {:else}
                <div
                  role="presentation"
                  data-testid={elementId}
                  id={elementId}
                  class="w-full flex gap-x-1 flex-none px-2 py-1 pointer-events-auto items-center hover:bg-slate-50 rounded-sm"
                >
                  <span class="truncate flex-1 text-left pointer-events-none"
                    >{allItemsMap.get(id)?.displayName ??
                      `Unknown ${type === "measure" ? "measure" : "dimension"}`}</span
                  >

                  <button
                    class="ml-auto hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                    on:click={() => {
                      selectedItems = selectedItems.filter((_, j) => j !== i);
                      onSelectedChange(selectedItems);
                    }}
                    on:mousedown|stopPropagation={() => {
                      // NO-OP
                    }}
                    disabled={selectedItems.length === 1}
                    class:pointer-events-none={selectedItems.length === 1}
                    class:opacity-50={selectedItems.length === 1}
                  >
                    <EyeIcon size="14px" color="#6b7280" />
                  </button>
                </div>
              {/if}
            {/each}
          </DraggableList>
        {/if}
      </div>

      {#if selectedItems.length < allItems.length}
        <span class="flex-none h-px bg-slate-200 w-full" />
        <div class="hidden-section flex flex-col flex-1 min-h-0 p-1.5 pt-0">
          <header
            class="flex-none flex py-1.5 justify-between px-2 sticky top-0 from-white from-80% to-transparent bg-gradient-to-b"
          >
            <h3
              class="uppercase text-xs text-gray-500 font-semibold from-white from-80% to-transparent bg-gradient-to-b"
            >
              Hidden {type === "measure" ? "Measures" : "Dimensions"}
            </h3>
            <button
              class="pointer-events-auto text-primary-500 text-[11px] font-medium"
              on:click={() => {
                selectedItems = allItems.map((item) => item.name ?? "");
                onSelectedChange(selectedItems);
              }}
            >
              Show all
            </button>
          </header>
          {#if filteredHiddenItems.length === 0}
            <div class="px-2 py-2 text-xs text-gray-500">
              {searchText
                ? `No matching hidden ${type === "measure" ? "measures" : "dimensions"}`
                : `No hidden ${type === "measure" ? "measures" : "dimensions"}`}
            </div>
          {:else}
            {#each filteredHiddenItems as [id = "", item], i (i)}
              {@const elementId = `hidden-${type === "measure" ? "measures" : "dimensions"}-${id}`}
              {#if item.description}
                <Tooltip.Root openDelay={200} portal="body">
                  <Tooltip.Trigger>
                    <div
                      id={elementId}
                      data-testid={elementId}
                      style:height="{ITEM_HEIGHT}px"
                      class="w-full flex gap-x-1 px-2 py-1 justify-between pointer-events-auto items-center p-1 rounded-sm hover:bg-slate-50 cursor-pointer"
                      on:click={() => {
                        selectedItems = [...selectedItems, id];
                        onSelectedChange(selectedItems);
                      }}
                      on:keydown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          selectedItems = [...selectedItems, id];
                          onSelectedChange(selectedItems);
                        }
                      }}
                      role="presentation"
                    >
                      <span
                        class="truncate flex-1 text-left pointer-events-none"
                        >{item.displayName}</span
                      >

                      <button
                        class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                        on:click|stopPropagation={() => {
                          selectedItems = [...selectedItems, id];
                          onSelectedChange(selectedItems);
                        }}
                        aria-label="Toggle visibility"
                        data-testid="toggle-visibility-button"
                      >
                        <EyeOffIcon size="14px" color="#9ca3af" />
                      </button>
                    </div>
                  </Tooltip.Trigger>

                  <Tooltip.Content
                    side="right"
                    sideOffset={12}
                    class="z-popover"
                  >
                    <div
                      class="bg-gray-800 text-gray-50 rounded p-2 pt-1 pb-1 shadow-md pointer-events-none z-50"
                    >
                      {item.description}
                    </div>
                  </Tooltip.Content>
                </Tooltip.Root>
              {:else}
                <div
                  id={elementId}
                  data-testid={elementId}
                  style:height="{ITEM_HEIGHT}px"
                  class="w-full flex gap-x-1 px-2 py-1 justify-between pointer-events-auto items-center p-1 rounded-sm hover:bg-slate-50 cursor-pointer"
                  on:click={() => {
                    selectedItems = [...selectedItems, id];
                    onSelectedChange(selectedItems);
                  }}
                  on:keydown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault();
                      selectedItems = [...selectedItems, id];
                      onSelectedChange(selectedItems);
                    }
                  }}
                  role="presentation"
                >
                  <span class="truncate flex-1 text-left pointer-events-none"
                    >{item.displayName}</span
                  >

                  <button
                    class="hover:bg-slate-200 p-1 rounded-sm active:bg-slate-300"
                    on:click|stopPropagation={() => {
                      selectedItems = [...selectedItems, id];
                      onSelectedChange(selectedItems);
                    }}
                    aria-label="Toggle visibility"
                    data-testid="toggle-visibility-button"
                  >
                    <EyeOffIcon size="14px" color="#9ca3af" />
                  </button>
                </div>
              {/if}
            {/each}
          {/if}
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
