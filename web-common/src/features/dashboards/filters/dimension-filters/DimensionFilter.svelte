<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
  import {
    getEffectiveSelectedValues,
    getItemLists,
    getSearchPlaceholder,
    shouldDisableApplyButton,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/helpers";
  import {
    mergeDimensionSearchValues,
    splitDimensionSearchText,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
  import DimensionFilterChipBody from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterChipBody.svelte";
  import DimensionFilterFooter from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterFooter.svelte";
  import DimensionFilterModeSelector from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterModeSelector.svelte";
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
  export let onMultiSelect: (values: string[]) => void;
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
  let selectedValuesProxy: string[] = [];

  $: ({ instanceId } = $runtime);

  $: resetFilterSettings(mode, sanitisedSearchText);

  // Sync proxy when selectedValues changes (for Select mode)
  $: if (curMode === DimensionFilterMode.Select) {
    selectedValuesProxy = [...selectedValues];
  }

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
      values:
        curMode === DimensionFilterMode.Select
          ? selectedValues
          : searchedBulkValues,
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

  $: searchPlaceholder = getSearchPlaceholder(curMode);

  $: error = errorFromSearchResults ?? errorFromAllSearchResultsCount;
  $: isFetching =
    isFetchingFromSearchResults ?? isFetchingFromAllSearchResultsCount;

  $: showExtraInfo = curMode !== DimensionFilterMode.Select; // || curSearchText.length > 0; (Add once we have docs)

  $: allSelected = Boolean(
    effectiveSelectedValues.length &&
      correctedSearchResults?.length === effectiveSelectedValues.length,
  );
  $: effectiveSelectedValues = getEffectiveSelectedValues(
    curMode,
    selectedValuesProxy,
    correctedSearchResults ?? [],
    selectedValues,
  );

  $: disableApplyButton = shouldDisableApplyButton(
    curMode,
    enableSearchCountQuery,
    inListTooLong,
  );

  // Split results into checked and unchecked for better UX (like SelectionDropdown)
  // Use actual selectedValues (not proxy) so items only sort after dropdown closes
  $: ({ checkedItems, uncheckedItems } = getItemLists(
    curMode,
    correctedSearchResults ?? [],
    selectedValues,
    curSearchText,
  ));

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
        selectedValuesProxy = [...selectedValues];
        break;

      case DimensionFilterMode.InList:
        curMode = DimensionFilterMode.InList;
        curSearchText = mergeDimensionSearchValues(selectedValues);
        searchedBulkValues = selectedValues; // Ensure searchedBulkValues includes existing selections
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

    // When switching to InList mode, include both existing selected values and new search values
    // This ensures the below-fold query can find existing selected values that might not be in top 250
    const allRelevantValues = [...new Set([...selectedValues, ...values])];
    searchedBulkValues = allRelevantValues;
    curMode = DimensionFilterMode.InList;
    inListTooLong = isUrlTooLongAfterInListFilter(values);
  }

  function handleModeChange(newMode: DimensionFilterMode) {
    if (newMode !== DimensionFilterMode.InList) {
      searchedBulkValues = [];
      // Reset proxy when switching to/from Select mode
      if (newMode === DimensionFilterMode.Select) {
        curExcludeMode = excludeMode;
        selectedValuesProxy = [...selectedValues];
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
      // Apply proxy changes for Select mode when dropdown closes
      if (curMode === DimensionFilterMode.Select) {
        applySelectModeChanges();
        // Don't reset immediately for Select mode - let props update first
        return;
      }

      if (selectedValues.length === 0 && !inputText) {
        // filter was cleared. so remove the filter
        onRemove();
      } else {
        // reset the settings on unmount (but not for Select mode)
        resetFilterSettings(mode, sanitisedSearchText);
      }
    }
  }

  function handleToggleExcludeMode() {
    curExcludeMode = !curExcludeMode;
  }

  function onToggleSelectAll() {
    if (curMode === DimensionFilterMode.Select) {
      // Update proxy for select all/deselect all
      if (allSelected) {
        selectedValuesProxy = selectedValuesProxy.filter(
          (v) => !correctedSearchResults?.includes(v),
        );
      } else {
        const newValues =
          correctedSearchResults?.filter(
            (v) => !selectedValuesProxy.includes(v),
          ) ?? [];
        selectedValuesProxy = [...selectedValuesProxy, ...newValues];
      }
    } else {
      correctedSearchResults?.forEach((dimensionValue) => {
        if (!allSelected && effectiveSelectedValues.includes(dimensionValue))
          return;

        onSelect(dimensionValue);
      });
    }
  }

  function onApply() {
    if (disableApplyButton) return;
    switch (curMode) {
      case DimensionFilterMode.Select:
        // Apply proxy changes for Select mode
        applySelectModeChanges();
        open = false;
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

  function applySelectModeChanges() {
    // Find values that were added or removed
    const currentValues = new Set(selectedValues);
    const proxyValues = new Set(selectedValuesProxy);

    // Apply all changes
    onMultiSelect(
      [...currentValues, ...proxyValues].filter((value) => {
        const wasSelected = currentValues.has(value);
        const isSelected = proxyValues.has(value);

        return wasSelected !== isSelected;
      }),
    );

    // Handle exclude mode toggle
    if (curExcludeMode !== excludeMode) {
      onToggleFilterMode();
    }
  }

  function handleItemClick(name: string) {
    if (curMode === DimensionFilterMode.Select) {
      // Update proxy instead of calling onSelect immediately
      if (selectedValuesProxy.includes(name)) {
        selectedValuesProxy = selectedValuesProxy.filter((v) => v !== name);
      } else {
        selectedValuesProxy = [...selectedValuesProxy, name];
      }
    } else {
      onSelect(name);
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
        {onRemove}
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
        <DimensionFilterModeSelector
          bind:mode={curMode}
          onModeChange={handleModeChange}
          size="md"
        />
        <Search
          bind:value={curSearchText}
          label={`${name} search list`}
          showBorderOnFocus={false}
          retainValueOnMount
          placeholder={searchPlaceholder}
          onSubmit={onApply}
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
          <!-- Show checked items first (only in Select mode and when not searching) -->
          {#if curMode === DimensionFilterMode.Select && !curSearchText}
            {#each checkedItems as name (name)}
              {@const selected = effectiveSelectedValues.includes(name)}
              {@const label = name ?? "null"}

              <svelte:component
                this={DropdownMenu.CheckboxItem}
                class="text-xs cursor-pointer"
                role="menuitem"
                checked={selected}
                showXForSelected={curExcludeMode}
                on:click={() => handleItemClick(name)}
              >
                <span>
                  {#if label.length > 240}
                    {label.slice(0, 240)}...
                  {:else}
                    {label}
                  {/if}
                </span>
              </svelte:component>
            {/each}
          {/if}

          <!-- Separator between checked and unchecked items -->
          {#if curMode === DimensionFilterMode.Select && !curSearchText && checkedItems.length > 0 && uncheckedItems.length > 0}
            <DropdownMenu.Separator />
          {/if}

          <!-- Show unchecked items (or all items for non-Select modes) -->
          {#each uncheckedItems as name (name)}
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
              on:click={() => handleItemClick(name)}
            >
              <span>
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
                {/if}
              </span>
            </svelte:component>
          {/each}

          <!-- Show "no results" only if both checked and unchecked are empty -->
          {#if uncheckedItems.length === 0 && (curMode !== DimensionFilterMode.Select || checkedItems.length === 0)}
            <div class="ui-copy-disabled text-center p-2 w-full">
              no results
            </div>
          {/if}
        </DropdownMenu.Group>
      {/if}
    </div>

    <DimensionFilterFooter
      mode={curMode}
      excludeMode={curExcludeMode}
      {allSelected}
      {disableApplyButton}
      onToggleExcludeMode={handleToggleExcludeMode}
      {onToggleSelectAll}
      {onApply}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
