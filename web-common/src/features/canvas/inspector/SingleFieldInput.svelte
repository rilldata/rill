<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useMetricFieldData } from "./selectors";

  export let metricName: string;
  export let label: string | undefined = undefined;
  export let id: string;
  export let selectedItem: string | undefined = undefined;
  export let type: "measure" | "dimension";
  export let searchableItems: string[] | undefined = undefined;
  export let onSelect: (item: string, displayName: string) => void = () => {};

  let open = false;
  let searchValue = "";

  const { instanceId } = $runtime;

  $: fieldData = useMetricFieldData(
    instanceId,
    metricName,
    type,
    searchableItems,
    searchValue,
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
      <Chip fullWidth caret {type} builders={[builder]}>
        <span class="font-bold truncate" slot="body">
          {#if selectedItem}
            {$fieldData.displayMap[selectedItem] || selectedItem}
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
        {#each $fieldData.filteredItems as item (item)}
          <!-- Hide item if it's the already selected one -->
          {#if item !== selectedItem}
            <DropdownMenu.Item
              class="pl-8 mx-1"
              on:click={() => {
                onSelect(item, $fieldData.displayMap[item] || item);
                open = false;
              }}
            >
              <slot {item}>
                {$fieldData.displayMap[item] || item}
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
