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
  import {
    DimensionFilterMode,
    DimensionFilterModeOptions,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-mode";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fly } from "svelte/transition";
  import {
    useDimensionSearch,
    useAllSearchResultsCount,
  } from "./dimensionFilterValues";

  export let name: string;
  export let metricsViewNames: string[];
  export let label: string;
  export let mode: DimensionFilterMode;
  export let selectedValues: string[];
  export let inputText: string | undefined;
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
  export let onSearch: (inputText: string) => void = () => {};
  export let onToggleFilterMode: () => void;

  let open = openOnMount && !selectedValues.length && !inputText;
  $: sanitisedSearchText = inputText?.replace(/^%/, "").replace(/%$/, "");
  let curMode = mode;
  let curSearchText = "";

  $: ({ instanceId } = $runtime);

  $: resetFilterSettings(mode, sanitisedSearchText);

  $: checkSearchText(curSearchText);

  let searchedBulkValues: string[] =
    mode === DimensionFilterMode.InList ? selectedValues : [];
  $: enableSearchQuery =
    Boolean(timeControlsReady && open) &&
    (curMode === DimensionFilterMode.Select ||
      (curMode === DimensionFilterMode.Contains && curSearchText.length > 0) ||
      (curMode === DimensionFilterMode.InList &&
        searchedBulkValues.length > 0));
  $: searchResultsQuery = useDimensionSearch(
    instanceId,
    metricsViewNames,
    name,
    {
      mode: curMode,
      values: searchedBulkValues,
      searchText: curSearchText,
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
    ((curMode === DimensionFilterMode.Contains && curSearchText.length > 0) ||
      (curMode === DimensionFilterMode.InList &&
        searchedBulkValues.length > 0));
  $: allSearchResultsCountQuery = useAllSearchResultsCount(
    instanceId,
    metricsViewNames,
    name,
    {
      mode: curMode,
      values: searchedBulkValues,
      searchText: curSearchText,
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
    ? curMode === DimensionFilterMode.Contains
      ? `${allSearchResultsCount} results`
      : `${allSearchResultsCount} of ${searchedBulkValues.length} matched`
    : "0 results";

  $: searchPlaceholder =
    curMode === DimensionFilterMode.Select
      ? "Enter search term or paste list of values"
      : curMode === DimensionFilterMode.InList
        ? "Paste a list separated by commas or \\n"
        : "Enter a search term";

  $: error = errorFromSearchResults ?? errorFromAllSearchResultsCount;
  $: isFetching =
    isFetchingFromSearchResults ?? isFetchingFromAllSearchResultsCount;

  $: showExtraInfo =
    curMode !== DimensionFilterMode.Select || curSearchText.length > 0;

  $: allSelected = Boolean(
    selectedValues.length &&
      correctedSearchResults?.length === selectedValues.length,
  );
  $: effectiveSelectedValues =
    curMode !== DimensionFilterMode.InList
      ? selectedValues
      : (correctedSearchResults ?? []);

  /**
   * Reset filter settings based on params to the component.
   */
  function resetFilterSettings(
    mode: DimensionFilterMode,
    sanitisedSearchText: string | undefined,
  ) {
    switch (mode) {
      case DimensionFilterMode.Select:
        curMode = DimensionFilterMode.Select;
        curSearchText = "";
        break;

      case DimensionFilterMode.InList:
        curMode = DimensionFilterMode.InList;
        curSearchText = selectedValues.join(",");
        break;

      case DimensionFilterMode.Contains:
        curMode = DimensionFilterMode.Contains;
        curSearchText = sanitisedSearchText ?? "";
        break;
    }
  }

  function checkSearchText(inputText: string) {
    let values = inputText.split(BulkValueSplitRegex);
    if (values.length > 0 && values[values.length - 1] === "") {
      // Remove the last empty value when the last character is a comma/newline
      values = values.slice(0, values.length - 1);
    }

    if (values.length <= 1) {
      if (curMode === DimensionFilterMode.InList) {
        searchedBulkValues = inputText === "" ? [] : values;
      }
      return;
    }
    searchedBulkValues = values;
    curMode = DimensionFilterMode.InList;
  }

  function handleModeChange(newMode: DimensionFilterMode) {
    if (newMode !== DimensionFilterMode.InList) {
      searchedBulkValues = [];
    } else {
      checkSearchText(curSearchText);
    }
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      curSearchText =
        mode === DimensionFilterMode.InList
          ? selectedValues.join(",")
          : (sanitisedSearchText ?? "");
    } else {
      if (selectedValues.length === 0 && !inputText) {
        // filter was cleared. so remove the filter
        onRemove();
      } else {
        // reset the settings on unmount
        resetFilterSettings(mode, sanitisedSearchText);
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
    if (curMode === DimensionFilterMode.InList) {
      onBulkSelect(searchedBulkValues);
      // mode = DimensionFilterMode.InList;
      open = false;
    } else if (curMode === DimensionFilterMode.Contains) {
      onSearch(curSearchText);
      // inputText = curSearchText;
      open = false;
    }
  }

  // Pasting a text with new line is not supported in input element.
  // So we need to manually replace newlines to commas.
  function onPaste(e: ClipboardEvent) {
    e.stopPropagation();
    e.preventDefault();

    const pastedData = e.clipboardData?.getData("Text");
    if (!pastedData) return;

    curSearchText = pastedData.replace(/[\n\r]/g, ",");
  }
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Enter") {
      onApply();
    }
  }}
/>

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
        label={`${name} filter`}
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
          values={curMode === DimensionFilterMode.InList
            ? searchedBulkValues
            : effectiveSelectedValues}
          matchedCount={allSearchResultsCount}
          loading={isFetchingFromAllSearchResultsCount}
          search={curMode === DimensionFilterMode.Contains
            ? curSearchText
            : undefined}
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

  <!-- This has significant differences with SearchableMenuContent with how search text is handled.
       So we have a custom implementation here to not overload SearchableMenuContent unnecessarily. -->
  <DropdownMenu.Content
    align="start"
    class="flex flex-col max-h-96 w-[400px] overflow-hidden p-0"
  >
    <div class="flex flex-col px-3 pt-3">
      <div class="flex flex-row">
        <!-- min-w-[82px] We need the min width since the select component is adding ellipsis unnecessarily when label has a space. -->
        <Select
          id="search-mode"
          bind:value={curMode}
          options={DimensionFilterModeOptions}
          onChange={handleModeChange}
          size="md"
          minWidth={82}
          noRightBorder
        />
        <Search
          bind:value={curSearchText}
          label={`${name} search list`}
          showBorderOnFocus={false}
          noLeftBorder
          retainValueOnMount
          placeholder={searchPlaceholder}
          on:submit={onApply}
          on:paste={onPaste}
        />
      </div>
      {#if showExtraInfo}
        <div class="flex flex-row items-center justify-between pt-2 pb-1">
          {#if curMode !== DimensionFilterMode.Select}
            <DropdownMenu.Label
              class="pb-0 uppercase text-[10px] text-gray-500"
              aria-label={`${name} result count`}
            >
              {searchResultCountText}
            </DropdownMenu.Label>
          {:else}
            <div class="grow" />
          {/if}

          <!-- Add it back once we have the docs -->
          <!--  <a-->
          <!--    href="https://docs.rilldata.com/"-->
          <!--    target="_blank"-->
          <!--    class="text-primary-600 font-medium justify-end"-->
          <!--  >-->
          <!--    Learn more-->
          <!--  </a>-->
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

            <svelte:component
              this={curMode === DimensionFilterMode.Select
                ? DropdownMenu.CheckboxItem
                : DropdownMenu.Item}
              class="text-xs cursor-pointer {curMode !==
              DimensionFilterMode.Select
                ? 'pl-3'
                : ''}"
              role="menuitem"
              checked={curMode === DimensionFilterMode.Select && selected}
              showXForSelected={excludeMode}
              disabled={curMode !== DimensionFilterMode.Select}
              on:click={() => onSelect(name)}
            >
              <span>
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
                {/if}
              </span>
            </svelte:component>
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
          label="Include exclude toggle"
        />
        <Label class="font-normal text-xs" for="include-exclude">Exclude</Label>
      </div>
      {#if curMode === DimensionFilterMode.Select}
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
