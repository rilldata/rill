<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import {
    DimensionFilterMode,
    DimensionFilterModeOptions,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-mode";
  import {
    mergeDimensionSearchValues,
    splitDimensionSearchText,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
  import DimensionFilterChipBody from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterChipBody.svelte";
  import { mergeDimensionAndMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import { getFiltersForOtherDimensions } from "@rilldata/web-common/features/dashboards/selectors";
  import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
  import type { V1Expression } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { fly } from "svelte/transition";
  import {
    useAllSearchResultsCount,
    useDimensionSearch,
  } from "web-common/src/features/dashboards/filters/dimension-filters/dimension-filter-values";

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
  export let whereFilter: V1Expression;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";
  export let onRemove: () => void;
  export let onApplyInList: (values: string[]) => void;
  export let onSelect: (value: string) => void;
  export let onApplyContainsMode: (inputText: string) => void = () => {};
  export let onToggleFilterMode: () => void;
  export let isUrlTooLongAfterInListFilter: (
    values: string[],
  ) => boolean = () => false;

  let open = openOnMount && !selectedValues.length && !inputText;
  $: sanitisedSearchText = inputText?.replace(/^%/, "").replace(/%$/, "");
  let curMode = mode;
  let curSearchText = "";
  let curExcludeMode = excludeMode;
  let inListTooLong = false;

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
      additionalFilter: sanitiseExpression(
        mergeDimensionAndMeasureFilters(
          getFiltersForOtherDimensions(whereFilter, name),
          [],
        ),
        undefined,
      ),
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
      additionalFilter: sanitiseExpression(
        mergeDimensionAndMeasureFilters(
          getFiltersForOtherDimensions(whereFilter, name),
          [],
        ),
        undefined,
      ),
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

  $: showExtraInfo = curMode !== DimensionFilterMode.Select; // || curSearchText.length > 0; (Add once we have docs)

  $: allSelected = Boolean(
    selectedValues.length &&
      correctedSearchResults?.length === selectedValues.length,
  );
  $: effectiveSelectedValues =
    curMode !== DimensionFilterMode.InList
      ? selectedValues
      : (correctedSearchResults ?? []);

  $: disableApplyButton =
    curMode === DimensionFilterMode.Select ||
    !enableSearchCountQuery ||
    inListTooLong;

  /**
   * Reset filter settings based on params to the component.
   */
  function resetFilterSettings(
    mode: DimensionFilterMode,
    sanitisedSearchText: string | undefined,
  ) {
    curExcludeMode = excludeMode;
    switch (mode) {
      case DimensionFilterMode.Select:
        curMode = DimensionFilterMode.Select;
        curSearchText = "";
        break;

      case DimensionFilterMode.InList:
        curMode = DimensionFilterMode.InList;
        curSearchText = mergeDimensionSearchValues(selectedValues);
        break;

      case DimensionFilterMode.Contains:
        curMode = DimensionFilterMode.Contains;
        curSearchText = sanitisedSearchText ?? "";
        break;
    }
  }

  function checkSearchText(inputText: string) {
    inListTooLong = false;

    // Do not check search text and possibly switch to InList when mode is Contains
    if (curMode === DimensionFilterMode.Contains) return;

    const values = splitDimensionSearchText(inputText);

    if (values.length <= 1) {
      if (curMode === DimensionFilterMode.InList) {
        searchedBulkValues = inputText === "" ? [] : values;
      }
      return;
    }
    searchedBulkValues = values;
    curMode = DimensionFilterMode.InList;
    inListTooLong = isUrlTooLongAfterInListFilter(values);
  }

  function handleModeChange(newMode: DimensionFilterMode) {
    if (newMode !== DimensionFilterMode.InList) {
      searchedBulkValues = [];
      // Since in select mode exclude toggle is reflected immediately, reset the mode when user switches to it.
      if (newMode === DimensionFilterMode.Select) {
        curExcludeMode = excludeMode;
      }
    } else {
      checkSearchText(curSearchText);
    }
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      curSearchText =
        mode === DimensionFilterMode.InList
          ? mergeDimensionSearchValues(selectedValues)
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

  function handleToggleExcludeMode() {
    curExcludeMode = !curExcludeMode;
    const shouldToggleImmediately = mode === curMode;
    if (shouldToggleImmediately) onToggleFilterMode();
  }

  function onToggleSelectAll() {
    correctedSearchResults?.forEach((dimensionValue) => {
      if (!allSelected && selectedValues.includes(dimensionValue)) return;

      onSelect(dimensionValue);
    });
  }

  function onApply() {
    if (disableApplyButton) return;
    switch (curMode) {
      case DimensionFilterMode.Select:
        onToggleSelectAll();
        // Do not close the dropdown.
        break;
      case DimensionFilterMode.InList:
        if (searchedBulkValues.length === 0) return;
        onApplyInList(searchedBulkValues);
        if (curExcludeMode !== excludeMode) onToggleFilterMode();
        open = false;
        break;
      case DimensionFilterMode.Contains:
        if (curSearchText.length === 0) return;
        onApplyContainsMode(curSearchText);
        if (curExcludeMode !== excludeMode) onToggleFilterMode();
        open = false;
        break;
    }
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
        exclude={curExcludeMode}
        label={`${name} filter`}
        theme
        on:remove={onRemove}
        removable={!readOnly}
        {readOnly}
        removeTooltipText="remove {selectedValues.length} value{selectedValues.length !==
        1
          ? 's'
          : ''}"
      >
        <DimensionFilterChipBody
          slot="body"
          label={curExcludeMode ? `Exclude ${label}` : label}
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
    {side}
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
          forcedTriggerStyle="rounded-r-none"
        />
        <Search
          bind:value={curSearchText}
          label={`${name} search list`}
          showBorderOnFocus={false}
          retainValueOnMount
          placeholder={searchPlaceholder}
          on:submit={onApply}
          forcedInputStyle="rounded-l-none"
          multiline
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
      class:pt-1={!showExtraInfo}
      class="flex flex-col flex-1 overflow-y-auto w-full h-fit min-h-24 pb-1"
    >
      {#if isFetching}
        <div class="min-h-9 flex flex-row items-center mx-auto">
          <LoadingSpinner />
        </div>
      {:else if error}
        <div class="min-h-9 p-3 text-center text-red-600 text-xs">error</div>
      {:else if inListTooLong}
        <div class="min-h-9 p-3 text-center text-red-600 text-xs">
          List is too long. Please remove some values.
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
              showXForSelected={curExcludeMode}
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
          checked={curExcludeMode}
          id="include-exclude"
          small
          on:click={handleToggleExcludeMode}
          label="Include exclude toggle"
        />
        <Label class="font-normal text-xs" for="include-exclude">Exclude</Label>
      </div>
      {#if curMode === DimensionFilterMode.Select}
        <Button onClick={onToggleSelectAll} type="plain" class="justify-end">
          {#if allSelected}
            Deselect all
          {:else}
            Select all
          {/if}
        </Button>
      {:else}
        <Button
          onClick={onApply}
          type="primary"
          class="justify-end"
          disabled={disableApplyButton}
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
