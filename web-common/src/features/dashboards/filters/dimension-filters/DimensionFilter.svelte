<script context="module">
  const BulkValueSplitRegex = /\s*[,\n]\s*/;
</script>

<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import RemovableListBody from "@rilldata/web-common/components/chip/removable-list-chip/RemovableListBody.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fly } from "svelte/transition";
  import {
    useDimensionSearch,
    useAllSearchResultsCount,
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
  let curSearchText = "";

  $: ({ instanceId } = $runtime);

  enum SearchMode {
    Select = "Select",
    Search = "Search",
    Bulk = "Bulk",
  }
  let mode: SearchMode = isMatchList
    ? SearchMode.Bulk
    : searchText?.length
      ? SearchMode.Search
      : SearchMode.Select;

  function updateBasedOnFilterSettings(
    isMatchList: boolean | undefined,
    sanitisedSearchText: string | undefined,
  ) {
    if (isMatchList) {
      mode = SearchMode.Bulk;
      curSearchText = selectedValues.join(",");
    } else if (sanitisedSearchText) {
      mode = SearchMode.Search;
      curSearchText = sanitisedSearchText ?? "";
    } else {
      mode = SearchMode.Select;
      curSearchText = "";
    }
  }
  $: updateBasedOnFilterSettings(isMatchList, sanitisedSearchText);

  function checkSearchText(searchText: string) {
    let values = searchText.split(BulkValueSplitRegex);
    if (values.length > 0 && values[values.length - 1] === "") {
      values = values.slice(0, values.length - 1);
    }

    if (values.length <= 1) {
      if (mode === SearchMode.Bulk) {
        searchedBulkValues = searchText === "" ? [] : values;
      }
      return;
    }
    searchedBulkValues = values;
    mode = SearchMode.Bulk;
  }
  $: checkSearchText(curSearchText);

  let searchedBulkValues: string[] = isMatchList ? selectedValues : [];
  $: enableSearchQuery =
    Boolean(timeControlsReady && open) &&
    (mode === SearchMode.Select ||
      (mode === SearchMode.Search && curSearchText.length > 0) ||
      (mode === SearchMode.Bulk && searchedBulkValues.length > 0));
  $: searchResultsQuery = useDimensionSearch(
    instanceId,
    metricsViewNames,
    name,
    {
      ...(mode === SearchMode.Bulk
        ? {
            values: searchedBulkValues,
          }
        : {
            searchText: curSearchText,
          }),
      timeStart,
      timeEnd,
      enabled: enableSearchQuery,
    },
  );
  $: ({
    data: searchResults,
    error: errorFromSearchResults,
    isFetching: isFetchingFromSearchResults,
  } = $searchResultsQuery);
  $: correctedSearchResults = enableSearchQuery ? searchResults : [];
  $: enableSearchCountQuery =
    Boolean(timeControlsReady) &&
    ((mode === SearchMode.Search && curSearchText.length > 0) ||
      (mode === SearchMode.Bulk && searchedBulkValues.length > 0));
  $: allSearchResultsCountQuery = useAllSearchResultsCount(
    instanceId,
    metricsViewNames,
    name,
    {
      ...(mode === SearchMode.Bulk
        ? {
            values: searchedBulkValues,
          }
        : {
            searchText: curSearchText,
          }),
      timeStart,
      timeEnd,
      enabled: enableSearchCountQuery,
    },
  );
  $: ({
    data: allSearchResultsCount,
    error: errorFromAllSearchResultsCount,
    isFetching: isFetchingFromAllSearchResultsCount,
  } = $allSearchResultsCountQuery);
  $: searchResultCountText = enableSearchCountQuery
    ? mode === SearchMode.Search
      ? `${allSearchResultsCount} results`
      : `${allSearchResultsCount} of ${searchedBulkValues.length} matched`
    : "0 results";

  $: error = errorFromSearchResults ?? errorFromAllSearchResultsCount;
  $: isFetching =
    isFetchingFromSearchResults ?? isFetchingFromAllSearchResultsCount;

  $: showExtraInfo = mode !== SearchMode.Select || curSearchText.length > 0;

  $: allSelected = Boolean(
    selectedValues.length &&
      correctedSearchResults?.length === selectedValues.length,
  );
  $: effectiveSelectedValues =
    mode !== SearchMode.Bulk ? selectedValues : (correctedSearchResults ?? []);

  function handleModeChange(newMode: SearchMode) {
    if (newMode !== SearchMode.Bulk) {
      searchedBulkValues = [];
    } else {
      checkSearchText(curSearchText);
    }
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      curSearchText = isMatchList
        ? selectedValues.join(",")
        : (sanitisedSearchText ?? "");
    } else {
      if (selectedValues.length === 0 && !searchText) {
        // filter was cleared. so remove the filter
        onRemove();
      } else {
        // reset the settings on unmount
        updateBasedOnFilterSettings(isMatchList, sanitisedSearchText);
      }
    }
  }

  function onToggleSelectAll() {
    searchResults?.forEach((dimensionValue) => {
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
  onOpenChange={handleOpenChange}
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
        label={`${name} view filter`}
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
          matchedCount={allSearchResultsCount}
          loading={isFetchingFromAllSearchResultsCount}
          search={mode === SearchMode.Search ? curSearchText : undefined}
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
    <div class="flex flex-col px-3 pt-3">
      <div class="flex flex-row">
        <!-- min-w-[82px] -->
        <Select
          id="search-mode"
          bind:value={mode}
          options={[
            {
              value: SearchMode.Select,
              label: "Select",
              description: "Manually select values for this filter",
            },
            {
              value: SearchMode.Search,
              label: "Contains",
              description: "Create a dynamic filter based on a search term",
            },
            {
              value: SearchMode.Bulk,
              label: "In List",
              description: "Create a filter based on a list of values",
            },
          ]}
          onChange={handleModeChange}
          size="md"
          minWidth={82}
        />
        <Search
          bind:value={curSearchText}
          label={`${name} search list`}
          showBorderOnFocus={false}
          retailValueOnMount
          placeholder="Enter search term or paste list of values"
        />
      </div>
      {#if showExtraInfo}
        <div class="flex flex-row items-center justify-between pt-2 pb-1">
          {#if mode !== SearchMode.Select}
            <DropdownMenu.Label
              class="pb-0 uppercase text-[10px] text-gray-500"
              aria-label={`${name} result count`}
            >
              {searchResultCountText}
            </DropdownMenu.Label>
          {:else}
            <div class="grow" />
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

    {#if showExtraInfo}
      <DropdownMenu.Separator class="bg-slate-200" />
    {/if}

    <div
      class="flex flex-col flex-1 overflow-y-auto w-full h-fit min-h-24 pb-1"
    >
      {#if isFetching}
        <div class="min-h-9 flex flex-row items-center mx-auto">
          <LoadingSpinner />
        </div>
      {:else if error}
        <div class="min-h-9 p-3 text-center text-red-600 text-xs">
          {error}
        </div>
      {:else if correctedSearchResults}
        <DropdownMenu.Group class="px-1" aria-label={`${name} results`}>
          {#each correctedSearchResults as name (name)}
            {@const selected = effectiveSelectedValues.includes(name)}
            {@const label = name ?? "null"}

            <DropdownMenu.CheckboxItem
              class="text-xs cursor-pointer {mode !== SearchMode.Select
                ? 'pl-3'
                : ''}"
              role="menuitem"
              checked={mode === SearchMode.Select && selected}
              showXForSelected={excludeMode}
              disabled={mode !== SearchMode.Select}
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
        <Button
          on:click={onApply}
          type="primary"
          class="justify-end"
          disabled={!enableSearchCountQuery}
        >
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
