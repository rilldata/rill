<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
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
  export let showClearOption = false;
  export let onClear: (() => void) | undefined = undefined;

  let open = false;
  let searchValue = "";

  $: ({ instanceId } = $runtime);

  $: ctx = getCanvasStore(canvasName, instanceId);
  $: ({ getTimeDimensionForMetricView } = ctx.canvasEntity.metricsView);

  $: timeDimension = getTimeDimensionForMetricView(metricName);

  $: isTimeSelected = $timeDimension && selectedItem === $timeDimension;
  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    [type],
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
        type={isTimeSelected ? "time" : type}
        builders={[builder]}
      >
        <span class="font-bold truncate" slot="body">
          {#if isTimeSelected}
            Time
          {:else if selectedItem}
            {$fieldData.displayMap[selectedItem]?.label || selectedItem}
          {:else}
            Select a {type} field
          {/if}
        </span>
      </Chip>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content sameWidth class="p-0">
      <div class="p-3 pb-1">
        <Search bind:value={searchValue} autofocus={false} />
      </div>
      <div class="max-h-64 overflow-y-auto">
        {#if selectedItem && showClearOption && onClear && $fieldData.displayMap[selectedItem]}
          <DropdownMenu.Item
            class="pl-8 mx-1"
            on:click={(e) => {
              e.stopPropagation();
            }}
          >
            <div class="flex items-center justify-between w-full">
              <span
                >{$fieldData.displayMap[selectedItem]?.label ||
                  selectedItem}</span
              >
              <button
                class="ml-2 text-gray-500 hover:text-gray-700"
                aria-label="Clear"
                on:click|stopPropagation={() => {
                  onClear();
                  open = false;
                }}
                type="button"
              >
                <CancelCircle size="16px" />
              </button>
            </div>
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
        {/if}
        {#if includeTime && $timeDimension}
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
        {#each $fieldData.filteredItems as item (item)}
          {#if item !== selectedItem}
            <DropdownMenu.Item
              class="pl-8 mx-1"
              on:click={() => {
                onSelect(item, $fieldData.displayMap[item]?.label || item);
                open = false;
              }}
            >
              <slot {item}>
                {$fieldData.displayMap[item]?.label || item}
              </slot>
            </DropdownMenu.Item>
          {/if}
        {:else}
          {#if searchValue}
            <div class="ui-copy-disabled text-center p-2 w-full">
              no results
            </div>
          {/if}
        {/each}
      </div>
    </DropdownMenu.Content>
  </DropdownMenu.Root>
</div>
