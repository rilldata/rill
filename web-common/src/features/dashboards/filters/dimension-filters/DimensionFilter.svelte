<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fly } from "svelte/transition";
  import { useDimensionSearch } from "./dimensionFilterValues";

  export let name: string;
  export let metricsViewNames: string[];
  export let label: string;
  export let selectedValues: string[];
  export let excludeMode: boolean;
  export let openOnMount: boolean = true;
  export let readOnly: boolean = false;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let timeControlsReady: boolean | undefined;
  export let smallChip = false;
  export let onRemove: () => void;
  export let onSelect: (value: string) => void;
  export let onToggleFilterMode: () => void;

  let open = openOnMount && !selectedValues.length;
  let searchText = "";

  $: ({ instanceId } = $runtime);

  $: searchValues = useDimensionSearch(
    instanceId,
    metricsViewNames,
    name,
    searchText,
    timeStart,
    timeEnd,
    Boolean(timeControlsReady && open),
  );
  $: ({ data, error, isFetching } = $searchValues);

  $: allSelected = Boolean(
    selectedValues.length && data?.length === selectedValues.length,
  );

  function onToggleSelectAll() {
    data?.forEach((dimensionValue) => {
      if (!allSelected && selectedValues.includes(dimensionValue)) return;

      onSelect(dimensionValue);
    });
  }
</script>

<DropdownMenu.Root
  bind:open
  typeahead={false}
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open) {
      searchText = "";
    } else {
      if (selectedValues.length === 0) {
        onRemove();
      }
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={open || readOnly}
    >
      <Chip
        builders={[builder]}
        type="dimension"
        active={open}
        exclude={excludeMode}
        label="View filter"
        on:remove={onRemove}
        removable={!readOnly}
        {readOnly}
      >
        <svelte:fragment slot="remove-tooltip">
          <slot name="remove-tooltip-content">
            remove {selectedValues.length}
            value{#if selectedValues.length !== 1}s{/if} for {name}</slot
          >
        </svelte:fragment>

        <RemovableListBody
          slot="body"
          label={excludeMode ? `Exclude ${label}` : label}
          show={1}
          {smallChip}
          values={selectedValues}
        />
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description">dimension</svelte:fragment>
          </TooltipTitle>
          Click to edit the the filters in this dimension
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <!-- There will be some custom controls for this. Until we have the full design have a custom dropdown here. -->
  <DropdownMenu.Content
    align="start"
    class="flex flex-col max-h-96 w-72 overflow-hidden p-0"
  >
    <div class="px-3 pt-3 pb-1">
      <Search
        bind:value={searchText}
        label="Search list"
        showBorderOnFocus={false}
      />
    </div>

    <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
      {#if isFetching}
        <div class="min-h-9 flex flex-row items-center mx-auto">
          <LoadingSpinner />
        </div>
      {:else if error}
        <div class="min-h-9 p-3 text-center text-red-600 text-xs">
          {error.response?.data?.message}
        </div>
      {:else if data}
        <DropdownMenu.Group class="px-1">
          {#each data as name (name)}
            {@const selected = selectedValues.includes(name)}

            <DropdownMenu.CheckboxItem
              class="text-xs cursor-pointer"
              role="menuitem"
              checked={selected}
              showXForSelected={excludeMode}
              on:click={() => onSelect(name)}
            >
              <span>
                {#if name.length > 240}
                  {name.slice(0, 240)}...
                {:else}
                  {name}
                {/if}
              </span>
            </DropdownMenu.CheckboxItem>
          {:else}
            <div class="ui-copy-disabled text-center p-2 w-full">
              no results
            </div>
          {/each}
        </DropdownMenu.Group>
      {/if}
    </div>

    <footer>
      <Button on:click={onToggleSelectAll} type="plain">
        {#if allSelected}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>
      <Button on:click={onToggleFilterMode} type="secondary">
        {#if excludeMode}
          Include
        {:else}
          Exclude
        {/if}
      </Button>
    </footer>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-end;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
