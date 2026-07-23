<script lang="ts">
  import { fly } from "svelte/transition";
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
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import PinButton from "../PinButton.svelte";
  import RequiredButton from "../RequiredButton.svelte";
  import { DimensionFilterManager } from "./dimension-filter-manager.svelte.ts";
  import {
    getAllSearchResultsCount,
    getDimensionSearchQuery,
  } from "@rilldata/web-common/features/dashboards/filters/manager/queries.svelte.ts";
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/expression-filter-manager.svelte.ts";

  let {
    manager,
    dimensionManager,
    openOnMount = true,
    readOnly = false,
    timeStart,
    timeEnd,
    timeDimension = undefined,
    timeControlsReady,
    smallChip = false,
    side = "bottom",
    isUrlTooLongAfterInListFilter = () => false,
  }: {
    manager: ExpressionFilterManager;
    dimensionManager: DimensionFilterManager;
    openOnMount?: boolean;
    readOnly?: boolean;
    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeDimension: string | undefined;
    timeControlsReady: boolean | undefined;
    smallChip?: boolean;
    side?: "top" | "right" | "bottom" | "left";
    isUrlTooLongAfterInListFilter?: (values: string[]) => boolean;
  } = $props();

  let proxyDimensionManager = $derived(dimensionManager.clone());

  let open = $state(
    // eslint-disable-next-line svelte/valid-compile
    openOnMount &&
      // eslint-disable-next-line svelte/valid-compile
      !dimensionManager.selectedValues?.length &&
      // eslint-disable-next-line svelte/valid-compile
      !dimensionManager.inputText,
  );
  // eslint-disable-next-line svelte/valid-compile
  let curSearchText = $state(dimensionManager.inputText ?? "");
  let inListTooLong = $state(false);

  const client = useRuntimeClient();

  let missingRequired = $derived(
    Boolean(dimensionManager.pinned && !dimensionManager.expr),
  );

  $effect(() => checkSearchText(curSearchText));

  let enableSearchQuery = $derived(
    Boolean(timeControlsReady && open) &&
      (proxyDimensionManager.mode === DimensionFilterMode.Select ||
        (proxyDimensionManager.mode === DimensionFilterMode.Contains &&
          curSearchText.length > 0) ||
        (proxyDimensionManager.mode === DimensionFilterMode.InList &&
          proxyDimensionManager.selectedValues.length > 0)),
  );

  let searchResultsQuery = $derived(
    getDimensionSearchQuery(client, manager, dimensionManager, {
      mode: proxyDimensionManager.mode,
      values: proxyDimensionManager.selectedValues,
      searchText: curSearchText,
      timeStart,
      timeEnd,
      timeDimension,
      enabled: enableSearchQuery,
    }),
  );
  let {
    data: searchResults,
    error: errorFromSearchResults,
    isFetching: isFetchingFromSearchResults,
  } = $derived($searchResultsQuery);
  let correctedSearchResults = $derived(enableSearchQuery ? searchResults : []);

  let enableSearchCountQuery = $derived(
    Boolean(timeControlsReady) &&
      ((proxyDimensionManager.mode === DimensionFilterMode.Contains &&
        curSearchText.length > 0) ||
        (proxyDimensionManager.mode === DimensionFilterMode.InList &&
          proxyDimensionManager.selectedValues.length > 0)),
  );

  let allSearchResultsCountQuery = $derived(
    getAllSearchResultsCount(client, manager, dimensionManager, {
      mode: proxyDimensionManager.mode,
      values: proxyDimensionManager.selectedValues,
      searchText: curSearchText,
      timeStart,
      timeEnd,
      timeDimension,
      enabled: enableSearchCountQuery,
    }),
  );
  let {
    data: allSearchResultsCount,
    error: errorFromAllSearchResultsCount,
    isFetching: isFetchingFromAllSearchResultsCount,
  } = $derived($allSearchResultsCountQuery);
  let searchResultCountText = $derived(
    enableSearchCountQuery
      ? proxyDimensionManager.mode === DimensionFilterMode.Contains
        ? `${allSearchResultsCount} results`
        : `${allSearchResultsCount} of ${proxyDimensionManager.selectedValues.length} matched`
      : "0 results",
  );

  let searchPlaceholder = $derived(
    getSearchPlaceholder(proxyDimensionManager.mode),
  );

  let error = $derived(
    errorFromSearchResults ?? errorFromAllSearchResultsCount,
  );
  let isFetching = $derived(
    isFetchingFromSearchResults ?? isFetchingFromAllSearchResultsCount,
  );

  let showExtraInfo = $derived(
    proxyDimensionManager.mode !== DimensionFilterMode.Select,
  ); // || curSearchText.length > 0; (Add once we have docs)

  let effectiveSelectedValues = $derived(
    getEffectiveSelectedValues(
      proxyDimensionManager.mode,
      proxyDimensionManager.selectedValues,
      correctedSearchResults ?? [],
      dimensionManager.selectedValues,
    ),
  );
  let allSelected = $derived(
    Boolean(
      effectiveSelectedValues.length &&
        correctedSearchResults?.length === effectiveSelectedValues.length,
    ),
  );

  let disableApplyButton = $derived(
    shouldDisableApplyButton(
      proxyDimensionManager.mode,
      enableSearchCountQuery,
      inListTooLong,
    ),
  );

  // Split results into checked and unchecked for better UX (like SelectionDropdown)
  // Use actual selectedValues (not proxy) so items only sort after dropdown closes
  let { checkedItems, uncheckedItems } = $derived(
    getItemLists(
      proxyDimensionManager.mode,
      correctedSearchResults ?? [],
      dimensionManager.selectedValues,
      curSearchText,
    ),
  );

  function checkSearchText(inputText: string) {
    inListTooLong = false;

    // Only InList mode parses bulk values. Other modes treat input as search text.
    if (proxyDimensionManager.mode !== DimensionFilterMode.InList) return;

    const values = splitDimensionSearchText(inputText);

    if (values.length <= 1) {
      proxyDimensionManager.selectedValues = inputText === "" ? [] : values;
      return;
    }

    // Include both existing selected values and new search values so the
    // below-fold query can find existing selected values that might not be
    // in the top 250.
    proxyDimensionManager.selectedValues = [
      ...new Set([...proxyDimensionManager.selectedValues, ...values]),
    ];
    inListTooLong = isUrlTooLongAfterInListFilter(
      proxyDimensionManager.selectedValues,
    );
  }

  function handleModeChange(newMode: DimensionFilterMode) {
    curSearchText = "";
    switch (newMode) {
      case DimensionFilterMode.Select:
        proxyDimensionManager.setSelectedValues(
          [...proxyDimensionManager.selectedValues],
          proxyDimensionManager.exclude,
        );
        break;

      case DimensionFilterMode.InList:
        proxyDimensionManager.setInList(
          [...proxyDimensionManager.selectedValues],
          proxyDimensionManager.exclude,
        );
        break;

      case DimensionFilterMode.Contains:
        proxyDimensionManager.setContainsText(
          curSearchText,
          proxyDimensionManager.exclude,
        );
        break;
    }
    checkSearchText(curSearchText);
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      curSearchText =
        dimensionManager.mode === DimensionFilterMode.InList
          ? mergeDimensionSearchValues(dimensionManager.selectedValues)
          : dimensionManager.inputText;
    } else {
      // Apply proxy changes for Select mode when dropdown closes
      if (proxyDimensionManager.mode === DimensionFilterMode.Select) {
        dimensionManager.expr = proxyDimensionManager.expr;
      } else {
        proxyDimensionManager = dimensionManager.clone();
      }
    }
  }

  function handleToggleExcludeMode() {
    proxyDimensionManager.toggleExclude();
  }

  function onToggleSelectAll() {
    proxyDimensionManager.setSelectedValues(
      [
        ...proxyDimensionManager.selectedValues,
        ...(correctedSearchResults ?? []),
      ],
      proxyDimensionManager.exclude,
    );
  }

  function onApply(close = true) {
    if (disableApplyButton) return;
    dimensionManager.expr = proxyDimensionManager.expr;
    if (close) open = false;
  }

  function handleItemClick(value: string) {
    proxyDimensionManager.toggleValue(value, false);
  }
</script>

<svelte:window
  onkeydown={async (e) => {
    if (e.key === "Enter") onApply();
  }}
/>

<DropdownMenu.Root bind:open onOpenChange={handleOpenChange}>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Tooltip
        activeDelay={500}
        alignment="start"
        distance={8}
        location="bottom"
        suppress={open || readOnly}
      >
        <Chip
          {...props}
          type="dimension"
          gray={dimensionManager.selectedValues.length === 0 &&
            !dimensionManager.inputText}
          error={!!missingRequired}
          active={open}
          exclude={proxyDimensionManager.exclude}
          label={`${dimensionManager.name} filter`}
          theme
          onRemove={() => dimensionManager.clear()}
          removable={!readOnly &&
            !proxyDimensionManager.pinned &&
            !dimensionManager.pinned}
          {readOnly}
          removeTooltipText="remove {dimensionManager.selectedValues
            .length} value{dimensionManager.selectedValues.length !== 1
            ? 's'
            : ''}"
        >
          <DimensionFilterChipBody
            slot="body"
            label={proxyDimensionManager.exclude
              ? `Exclude ${dimensionManager.label}`
              : dimensionManager.label}
            show={1}
            {smallChip}
            values={proxyDimensionManager.mode === DimensionFilterMode.InList
              ? proxyDimensionManager.selectedValues
              : effectiveSelectedValues}
            matchedCount={allSearchResultsCount}
            loading={isFetchingFromAllSearchResultsCount}
            search={proxyDimensionManager.mode === DimensionFilterMode.Contains
              ? curSearchText
              : undefined}
          />
        </Chip>
        <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
          <TooltipContent maxWidth="400px">
            <TooltipTitle>
              <svelte:fragment slot="name"
                >{dimensionManager.name}</svelte:fragment
              >
              <svelte:fragment slot="description"
                >{dimensionManager.pinned
                  ? "required dimension"
                  : "dimension"}</svelte:fragment
              >
            </TooltipTitle>
            {#if missingRequired}
              This filter is required. Select a value to load the dashboard.
            {:else}
              Click to edit the filters in this dimension
            {/if}
          </TooltipContent>
        </div>
      </Tooltip>
    {/snippet}
  </DropdownMenu.Trigger>

  <!-- This has significant differences with SearchableMenuContent with how search text is handled.
       So we have a custom implementation here to not overload SearchableMenuContent unnecessarily. -->
  <DropdownMenu.Content
    align="start"
    {side}
    class="flex flex-col max-h-96 w-[400px] overflow-hidden p-0"
  >
    <div class="flex flex-col px-3 pt-3">
      {#if dimensionManager.editing}
        <div
          class="flex flex-row items-center justify-between mb-2 pointer-events-auto"
        >
          <b>{dimensionManager.label}</b>

          <div class="flex flex-row items-center gap-x-1">
            <RequiredButton
              required={proxyDimensionManager.required}
              onToggleRequired={() => proxyDimensionManager.toggleRequired()}
            />
            <PinButton
              pinned={proxyDimensionManager.pinned}
              onTogglePin={() => proxyDimensionManager.togglePinned()}
            />
          </div>
        </div>
      {/if}
      <div class="flex flex-row">
        <DimensionFilterModeSelector
          bind:mode={proxyDimensionManager.mode}
          onModeChange={handleModeChange}
          size="md"
        />
        <Search
          bind:value={curSearchText}
          label={`${dimensionManager.name} search list`}
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
          {#if proxyDimensionManager.mode !== DimensionFilterMode.Select}
            <span
              class="px-2 py-1.5 pb-0 uppercase text-[10px] text-fg-secondary font-semibold"
              aria-label={`${dimensionManager.name} result count`}
            >
              {searchResultCountText}
            </span>
          {:else}
            <div class="grow"></div>
          {/if}
        </div>
      {/if}
    </div>

    {#if showExtraInfo}
      <DropdownMenu.Separator class="bg-gray-200" />
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
        <DropdownMenu.Group
          class="px-1"
          aria-label={`${dimensionManager.name} results`}
        >
          <!-- Show checked items first (only in Select mode and when not searching) -->
          {#if proxyDimensionManager.mode === DimensionFilterMode.Select && !curSearchText}
            {#each checkedItems as name (name)}
              {@const selected = effectiveSelectedValues.includes(name)}
              {@const label = name ?? "null"}

              <DropdownMenu.CheckboxItem
                closeOnSelect={false}
                class="text-xs cursor-pointer"
                checked={selected}
                showXForSelected={proxyDimensionManager.exclude}
                onclick={() => handleItemClick(name)}
              >
                <span>
                  {#if label.length > 240}
                    {label.slice(0, 240)}...
                  {:else}
                    {label}
                  {/if}
                </span>
              </DropdownMenu.CheckboxItem>
            {/each}
          {/if}

          <!-- Separator between checked and unchecked items -->
          {#if proxyDimensionManager.mode === DimensionFilterMode.Select && !curSearchText && checkedItems.length > 0 && uncheckedItems.length > 0}
            <DropdownMenu.Separator />
          {/if}

          <!-- Show unchecked items (or all items for non-Select modes) -->
          {#each uncheckedItems as name (name)}
            {@const selected = effectiveSelectedValues.includes(name)}
            {@const label = name ?? "null"}
            {@const ItemComponent =
              proxyDimensionManager.mode === DimensionFilterMode.Select
                ? DropdownMenu.CheckboxItem
                : DropdownMenu.Item}

            <ItemComponent
              closeOnSelect={false}
              class="text-xs cursor-pointer {proxyDimensionManager.mode !==
              DimensionFilterMode.Select
                ? 'pl-3'
                : ''}"
              checked={proxyDimensionManager.mode ===
                DimensionFilterMode.Select && selected}
              showXForSelected={proxyDimensionManager.exclude}
              disabled={proxyDimensionManager.mode !==
                DimensionFilterMode.Select}
              onclick={() => handleItemClick(name)}
            >
              <span>
                {#if label.length > 240}
                  {label.slice(0, 240)}...
                {:else}
                  {label}
                {/if}
              </span>
            </ItemComponent>
          {/each}

          <!-- Show "no results" only if both checked and unchecked are empty -->
          {#if uncheckedItems.length === 0 && (proxyDimensionManager.mode !== DimensionFilterMode.Select || checkedItems.length === 0)}
            <div class="text-fg-disabled text-center p-2 w-full">
              no results
            </div>
          {/if}
        </DropdownMenu.Group>
      {/if}
    </div>

    <DimensionFilterFooter
      mode={proxyDimensionManager.mode}
      excludeMode={proxyDimensionManager.exclude}
      {allSelected}
      {disableApplyButton}
      onToggleExcludeMode={handleToggleExcludeMode}
      {onToggleSelectAll}
      {onApply}
    />
  </DropdownMenu.Content>
</DropdownMenu.Root>
