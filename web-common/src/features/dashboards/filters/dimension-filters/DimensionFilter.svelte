<script lang="ts">
  import { Button, IconButton } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import { Tag } from "@rilldata/web-common/components/tag";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { XIcon } from "lucide-svelte";
  import { fly } from "svelte/transition";
  import {
    useBulkSearchMatchedCount,
    useBulkSearchResults,
    useDimensionSearch,
  } from "./dimensionFilterValues";

  export let name: string;
  export let metricsViewNames: string[];
  export let label: string;
  export let selectedValues: string[];
  export let searchText: string | undefined;
  export let isMatchList: boolean | undefined;
  export let excludeMode: boolean;
  export let openOnMount: boolean = true;
  export let readOnly: boolean = false;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let timeControlsReady: boolean | undefined;
  export let smallChip = false;
  export let onRemove: () => void;
  export let onBulkSelect: (values: string[]) => void;
  export let onSelect: (value: string) => void;
  export let onSearch: (searchText: string) => void = () => {};
  export let onToggleFilterMode: () => void;

  let open = openOnMount && !selectedValues.length && !searchText;
  $: sanitisedSearchText = searchText?.replace(/^%/, "").replace(/%$/, "");
  let curSearchText = isMatchList
    ? selectedValues.join(",")
    : (sanitisedSearchText ?? "");

  $: ({ instanceId } = $runtime);

  enum SearchMode {
    Select,
    Search,
    Bulk,
  }
  let mode: SearchMode = isMatchList
    ? SearchMode.Bulk
    : searchText?.length
      ? SearchMode.Search
      : SearchMode.Select;

  function updateBasedOnMatchList(isMatchList: boolean | undefined) {
    if (isMatchList) {
      mode = SearchMode.Bulk;
      curSearchText = selectedValues.join(",");
    } else if (mode === SearchMode.Bulk) {
      mode = SearchMode.Select;
      curSearchText = "";
    }
  }
  $: updateBasedOnMatchList(isMatchList);

  let searchedBulkValues: string[] = isMatchList ? selectedValues : [];
  $: searchValues = useDimensionSearch(
    instanceId,
    metricsViewNames,
    name,
    curSearchText,
    timeStart,
    timeEnd,
    Boolean(timeControlsReady && open) && mode !== SearchMode.Bulk,
  );
  $: ({
    data: dataFromSearch,
    error: errorFromSearch,
    isFetching: isFetchingFromSearch,
  } = $searchValues);
  $: bulkValues = useBulkSearchResults(
    instanceId,
    metricsViewNames,
    name,
    searchedBulkValues,
    timeStart,
    timeEnd,
    Boolean(timeControlsReady && open) && mode === SearchMode.Bulk,
  );
  $: ({
    data: dataFromBulk,
    error: errorFromBulk,
    isFetching: isFetchingFromBulk,
  } = $bulkValues);
  $: bulkMatchedCount = useBulkSearchMatchedCount(
    instanceId,
    metricsViewNames,
    name,
    searchedBulkValues,
    timeStart,
    timeEnd,
    Boolean(timeControlsReady) && mode === SearchMode.Bulk,
  );
  $: ({
    data: dataFromBulkMatchedCount,
    error: errorFromBulkMatchedCount,
    isFetching: isFetchingFromBulkMatchedCount,
  } = $bulkMatchedCount);

  $: data = mode === SearchMode.Bulk ? dataFromBulk : dataFromSearch;
  $: error = errorFromSearch ?? errorFromBulk ?? errorFromBulkMatchedCount;
  $: isFetching =
    isFetchingFromSearch ??
    isFetchingFromBulk ??
    isFetchingFromBulkMatchedCount;

  $: showExtraInfo = mode !== SearchMode.Select || curSearchText.length > 0;

  $: if (curSearchText.length > 0) {
    const values = curSearchText.split(/\s*,\s*/);
    if (values.length > 1) {
      searchedBulkValues = values;
      mode = SearchMode.Bulk;
    } else if (mode === SearchMode.Bulk) {
      searchedBulkValues = [];
      mode = SearchMode.Select;
    }
  } else if (mode === SearchMode.Bulk) {
    searchedBulkValues = [];
    mode = SearchMode.Select;
  }

  $: allSelected = Boolean(
    selectedValues.length && data?.length === selectedValues.length,
  );
  $: effectiveSelectedValues =
    mode !== SearchMode.Bulk ? selectedValues : (data ?? []);

  function onToggleSelectAll() {
    data?.forEach((dimensionValue) => {
      if (!allSelected && selectedValues.includes(dimensionValue)) return;

      onSelect(dimensionValue);
    });
  }

  function onApply() {
    if (mode === SearchMode.Bulk) {
      onBulkSelect(searchedBulkValues);
      isMatchList = true;
      open = false;
    } else {
      onSearch(curSearchText);
      searchText = curSearchText;
      open = false;
    }
  }
</script>

<DropdownMenu.Root
  bind:open
  typeahead={false}
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open) {
      curSearchText = isMatchList
        ? selectedValues.join(",")
        : (sanitisedSearchText ?? "");
    } else {
      if (selectedValues.length === 0 && !searchText) {
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
          values={mode === SearchMode.Bulk
            ? searchedBulkValues
            : effectiveSelectedValues}
          matchedCount={dataFromBulkMatchedCount}
          loading={isFetchingFromBulkMatchedCount}
          search={sanitisedSearchText}
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
    class="flex flex-col max-h-96 w-[400px] overflow-hidden p-0"
  >
    <div class="flex flex-col px-3 pt-3 pb-1">
      <Search
        bind:value={curSearchText}
        label="Search list"
        showBorderOnFocus={false}
        placeholder="Enter search term or paste list of values"
      />
      {#if showExtraInfo}
        <div class="flex flex-row items-center justify-between pt-3">
          {#if mode === SearchMode.Bulk}
            <Tag noSpan>
              <span>Match list</span>
              <IconButton
                disableHover
                size={12}
                on:click={() => (mode = SearchMode.Select)}
              >
                <XIcon size="12px" class="text-gray-500 cursor-pointer" />
              </IconButton>
            </Tag>
          {:else if mode === SearchMode.Search}
            <Tag noSpan>
              <span>Search</span>
              <IconButton
                disableHover
                size={12}
                on:click={() => (mode = SearchMode.Select)}
              >
                <XIcon size="12px" class="text-gray-500 cursor-pointer" />
              </IconButton>
            </Tag>
          {:else if curSearchText.length}
            <Button
              type="subtle"
              small
              on:click={() => (mode = SearchMode.Search)}
            >
              Convert to search filter
            </Button>
          {/if}
          <a
            href="https://docs.rilldata.com/"
            target="_blank"
            class="text-primary-600 font-medium justify-end"
          >
            Learn more
          </a>
        </div>
      {/if}
    </div>

    <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
      {#if isFetching}
        <div class="min-h-9 flex flex-row items-center mx-auto">
          <LoadingSpinner />
        </div>
      {:else if error}
        <div class="min-h-9 p-3 text-center text-red-600 text-xs">
          {error}
        </div>
      {:else if data}
        <DropdownMenu.Group class="px-1">
          {#each data as name (name)}
            {@const selected = effectiveSelectedValues.includes(name)}
            {@const label = name ?? "null"}

            <DropdownMenu.CheckboxItem
              class="text-xs cursor-pointer"
              role="menuitem"
              checked={selected}
              showXForSelected={excludeMode}
              on:click={() => onSelect(name)}
            >
              <span>
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
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
      <div class="flex items-center gap-x-1.5">
        <Switch
          checked={excludeMode}
          id="include-exclude"
          small
          on:click={onToggleFilterMode}
        />
        <Label class="font-normal text-xs" for="include-exclude">Exclude</Label>
      </div>
      {#if mode === SearchMode.Select}
        <Button on:click={onToggleSelectAll} type="plain" class="justify-end">
          {#if allSelected}
            Deselect all
          {:else}
            Select all
          {/if}
        </Button>
      {:else}
        <Button on:click={onApply} type="plain" class="justify-end">
          Apply
        </Button>
      {/if}
    </footer>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<style lang="postcss">
  footer {
    height: 42px;
    @apply border-t border-slate-300;
    @apply bg-slate-100;
    @apply flex flex-row flex-none items-center justify-between;
    @apply gap-x-2 p-2 px-3.5;
  }

  footer:is(.dark) {
    @apply bg-gray-800;
    @apply border-gray-700;
  }
</style>
