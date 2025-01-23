<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { MinusIcon, PlusIcon } from "lucide-svelte";
  import { useMetricFieldData } from "./selectors";

  export let metricName: string;
  export let label: string;
  export let id: string;
  export let selectedItems: string[] = [];
  export let type: "measure" | "dimension";
  export let searchableItems: string[] | undefined = undefined;
  export let onMultiSelect: (items: string[]) => void = () => {};

  let open = false;
  let searchValue = "";

  const ctx = getCanvasStateManagers();

  $: fieldData = useMetricFieldData(
    ctx,
    metricName,
    type,
    searchableItems,
    searchValue,
  );

  // This Set holds in-progress selection
  let selectedProxy = new Set(selectedItems);

  // Keep track when dropdown opens/closes
  $: if (!open) {
    // reset the proxy on close
    selectedProxy = new Set(selectedItems);
  }
</script>

<div class="flex flex-col gap-y-2 pt-1">
  <DropdownMenu.Root bind:open typeahead={false} closeOnItemClick={true}>
    <DropdownMenu.Trigger asChild let:builder>
      <div class="flex justify-between gap-x-2">
        <InputLabel small {label} {id} />
        <button use:builder.action {...builder} class="text-sm px-2 h-6">
          <PlusIcon size="14px" />
        </button>
      </div>
    </DropdownMenu.Trigger>

    <DropdownMenu.Content class="p-0 w-[300px]">
      <div class="p-3 pb-1">
        <Search bind:value={searchValue} autofocus={false} />
      </div>
      <div class="max-h-64 overflow-y-auto">
        {#each $fieldData.filteredItems as item (item)}
          <DropdownMenu.CheckboxItem
            checked={selectedProxy.has(item)}
            class="pl-8 mx-1"
            on:click={() => {
              if (selectedProxy.has(item)) {
                selectedProxy.delete(item);
              } else {
                selectedProxy.add(item);
              }
              onMultiSelect(Array.from(selectedProxy));
            }}
          >
            <slot {item}>
              {$fieldData.displayMap[item] || item}
            </slot>
          </DropdownMenu.CheckboxItem>
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

  {#if selectedItems?.length > 0}
    <div class="flex flex-col gap-1 mt-2">
      {#each selectedItems as item}
        <div class="flex items-center justify-between gap-x-2">
          <div class="flex-1">
            <Chip fullWidth {type}>
              <span class="font-bold truncate" slot="body">
                {$fieldData.displayMap[item] || item}
              </span>
            </Chip>
          </div>
          <button
            class="px-2 py-1 text-xs"
            on:click={() => {
              selectedProxy.delete(item);
              onMultiSelect(Array.from(selectedProxy));
            }}
          >
            <MinusIcon size="14px" />
          </button>
        </div>
      {/each}
    </div>
  {/if}
</div>
