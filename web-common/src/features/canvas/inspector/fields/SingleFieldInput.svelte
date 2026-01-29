<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useMetricFieldData } from "../selectors";

  export let metricName: string;
  export let label: string | undefined = undefined;
  export let id: string;
  export let selectedItem: string | undefined = undefined;
  export let type: "measure" | "dimension";
  export let includeTime = false;
  export let canvasName: string;
  export let searchableItems: string[] | undefined = undefined;
  export let excludedValues: string[] | undefined = undefined;
  export let onSelect: (item: string, displayName: string) => void = () => {};

  let open = false;
  let searchValue = "";

  $: hasSelection = isTimeSelected || !!selectedItem;

  function handleClear() {
    // Notify parent that the selection has been cleared.
    // Consumers that only care about the field name can treat an empty string as \"no field\".
    onSelect("", "");
  }

  $: ({ instanceId } = $runtime);

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: ({ getTimeDimensionForMetricView } = ctx.canvasEntity.metricsView);

  $: timeDimension = getTimeDimensionForMetricView(metricName);

  $: isTimeSelected = $timeDimension && selectedItem === $timeDimension;
  // Show all available field types in the dropdown; axis-specific filtering
  // is handled by higher-level inspector logic (e.g. excludedValues).
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    ["measure", "dimension", "time"],
    searchableItems,
    searchValue,
    excludedValues,
  );
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <div class="flex items-center gap-x-2">
    {#if label}
      <InputLabel small {label} {id} />
    {/if}
  </div>

  <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={false}>
    <DropdownMenu.Trigger asChild let:builder>
      <Chip
        fullWidth
        caret
        type={hasSelection ? (isTimeSelected ? "time" : type) : "measure"}
        gray={!hasSelection}
        removable={!!selectedItem}
        removeTooltipText={selectedItem ? "Clear selected field" : undefined}
        onRemove={handleClear}
        builders={[builder]}
      >
        <span class="font-bold truncate" slot="body">
          {#if isTimeSelected}
            Time
          {:else if selectedItem}
            {$fieldData.displayMap[selectedItem]?.label || selectedItem}
          {:else}
            Select a field
          {/if}
        </span>
      </Chip>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content class="p-0" sameWidth>
      <div class="p-3 pb-1">
        <Search bind:value={searchValue} autofocus={false} />
      </div>
      <div class="max-h-64 overflow-y-auto pb-2">
        {#if type == "dimension" && includeTime && $timeDimension}
          <DropdownMenu.Item
            class="pl-8 mx-1"
            on:click={() => {
              onSelect($timeDimension, "Time");
              open = false;
            }}
          >
            Time
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
        {/if}

        {#if $fieldData.filteredItems.length === 0 && searchValue}
          <div class="text-fg-disabled text-center p-2 w-full">
            no results
          </div>
        {:else}
          {#if $fieldData.filteredItems.some(
            (item) => $fieldData.displayMap[item]?.type === "measure",
          )}
            <div class="px-3 pt-2 pb-1 text-[10px] font-semibold text-fg-disabled">
              MEASURES
            </div>
            {#each $fieldData.filteredItems.filter(
              (item) => $fieldData.displayMap[item]?.type === "measure",
            ) as item (item)}
              {#if item !== selectedItem}
                <DropdownMenu.Item
                  class="pl-8 mx-1"
                  on:click={() => {
                    onSelect(
                      item,
                      $fieldData.displayMap[item]?.label || item,
                    );
                    open = false;
                  }}
                >
                  <slot {item}>
                    {$fieldData.displayMap[item]?.label || item}
                  </slot>
                </DropdownMenu.Item>
              {/if}
            {/each}
          {/if}

          {#if $fieldData.filteredItems.some(
            (item) => $fieldData.displayMap[item]?.type === "time",
          )}
            <div class="px-3 pt-2 pb-1 text-[10px] font-semibold text-fg-disabled">
              TIME
            </div>
            {#each $fieldData.filteredItems.filter(
              (item) => $fieldData.displayMap[item]?.type === "time",
            ) as item (item)}
              {#if item !== selectedItem}
                <DropdownMenu.Item
                  class="pl-8 mx-1"
                  on:click={() => {
                    onSelect(
                      item,
                      $fieldData.displayMap[item]?.label || item,
                    );
                    open = false;
                  }}
                >
                  <slot {item}>
                    {$fieldData.displayMap[item]?.label || item}
                  </slot>
                </DropdownMenu.Item>
              {/if}
            {/each}
          {/if}

          {#if $fieldData.filteredItems.some(
            (item) => $fieldData.displayMap[item]?.type === "dimension",
          )}
            <div class="px-3 pt-2 pb-1 text-[10px] font-semibold text-fg-disabled">
              DIMENSIONS
            </div>
            {#each $fieldData.filteredItems.filter(
              (item) => $fieldData.displayMap[item]?.type === "dimension",
            ) as item (item)}
              {#if item !== selectedItem}
                <DropdownMenu.Item
                  class="pl-8 mx-1"
                  on:click={() => {
                    onSelect(
                      item,
                      $fieldData.displayMap[item]?.label || item,
                    );
                    open = false;
                  }}
                >
                  <slot {item}>
                    {$fieldData.displayMap[item]?.label || item}
                  </slot>
                </DropdownMenu.Item>
              {/if}
            {/each}
          {/if}
        {/if}
      </div>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
