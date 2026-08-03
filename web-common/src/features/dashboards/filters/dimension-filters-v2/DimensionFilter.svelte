<script lang="ts">
  import { fly } from "svelte/transition";
  import { Chip } from "web-common/src/components/chip";
  import * as DropdownMenu from "web-common/src/components/dropdown-menu";
  import LoadingSpinner from "web-common/src/components/icons/LoadingSpinner.svelte";
  import { Search } from "web-common/src/components/search";
  import Tooltip from "web-common/src/components/tooltip/Tooltip.svelte";
  import TooltipContent from "web-common/src/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "web-common/src/components/tooltip/TooltipTitle.svelte";
  import { DimensionFilterMode } from "web-common/src/features/dashboards/filters/dimension-filters/constants";
  import {
    getEffectiveSelectedValues,
    getItemLists,
    getSearchPlaceholder,
    shouldDisableApplyButton,
  } from "web-common/src/features/dashboards/filters/dimension-filters/helpers";
  import {
    mergeDimensionSearchValues,
    splitDimensionSearchText,
  } from "web-common/src/features/dashboards/filters/dimension-filters/dimension-search-text-utils";
  import DimensionFilterChipBody from "web-common/src/features/dashboards/filters/dimension-filters/DimensionFilterChipBody.svelte";
  import DimensionFilterFooter from "web-common/src/features/dashboards/filters/dimension-filters/DimensionFilterFooter.svelte";
  import DimensionFilterModeSelector from "web-common/src/features/dashboards/filters/dimension-filters/DimensionFilterModeSelector.svelte";
  import { useRuntimeClient } from "web-common/src/runtime-client/v2";
  import PinButton from "../PinButton.svelte";
  import RequiredButton from "../RequiredButton.svelte";
  import { DimensionFilterManager } from "./DimensionFilterManager.svelte.ts";
  import {
    createDimensionSearchCountQuery,
    createDimensionSearchQuery,
  } from "./queries.svelte.ts";
  import type { ExpressionFilterManager } from "../ExpressionFilterManager.svelte.ts";
  import { onMount } from "svelte";
  import type { YAMLConfigProvider } from "web-common/src/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  let {
    manager,
    dimensionManager,
    yamlConfigProvider,
    openOnMount = true,
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
    yamlConfigProvider: YAMLConfigProvider;
    openOnMount?: boolean;
    timeStart: string | undefined;
    timeEnd: string | undefined;
    timeDimension?: string | undefined;
    timeControlsReady: boolean | undefined;
    smallChip?: boolean;
    side?: "top" | "right" | "bottom" | "left";
    isUrlTooLongAfterInListFilter?: (values: string[]) => boolean;
  } = $props();

  let proxyDimensionManager = $derived(dimensionManager.clone());

  let open = $state(false);
  let curSearchText = $state("");

  const client = useRuntimeClient();

  let pinned = $derived(
    Boolean(yamlConfigProvider.pinnedFilters[dimensionManager.name]),
  );
  let required = $derived(
    Boolean(yamlConfigProvider.requiredFilters[dimensionManager.name]),
  );
  let missingRequired = $derived(Boolean(required && !dimensionManager.expr));

  // Pinned and required are edited locally and only persisted once the dropdown closes.
  // svelte-ignore state_referenced_locally
  let curPinned = $state(pinned);
  // svelte-ignore state_referenced_locally
  let curRequired = $state(required);

  // Only InList mode parses the input as a list of values; the other modes treat it as search text,
  // so this is read behind a mode check everywhere it is used.
  let inListValues = $derived.by(() => {
    const values = splitDimensionSearchText(curSearchText);
    if (values.length <= 1) return curSearchText === "" ? [] : values;

    // Include both the applied values and the ones in the input so the below-fold query can find
    // applied values that might not be in the top 250.
    return [...new Set([...dimensionManager.selectedValues, ...values])];
  });
  let inListTooLong = $derived(
    proxyDimensionManager.mode === DimensionFilterMode.InList &&
      inListValues.length > 0 &&
      isUrlTooLongAfterInListFilter(inListValues),
  );

  let enableSearchQuery = $derived(
    Boolean(timeControlsReady && open) &&
      (proxyDimensionManager.mode === DimensionFilterMode.Select ||
        (proxyDimensionManager.mode === DimensionFilterMode.Contains &&
          curSearchText.length > 0) ||
        (proxyDimensionManager.mode === DimensionFilterMode.InList &&
          inListValues.length > 0)),
  );

  // The queries are created once and follow the dropdown through the args getter, so changing the
  // mode or the search text updates them instead of replacing them.
  const searchResultsQuery = createDimensionSearchQuery(client, () => ({
    manager,
    dimensionName: dimensionManager.name,
    mode: proxyDimensionManager.mode,
    values: inListValues,
    searchText: curSearchText,
    timeStart,
    timeEnd,
    timeDimension,
    enabled: enableSearchQuery,
  }));
  // The query only returns the top 250 values, so applied values outside that window are merged back
  // in: Select mode shows them as checked, InList mode lists them among the values it matched.
  let correctedSearchResults = $derived.by(() => {
    if (!enableSearchQuery) return [];

    const searchResults = searchResultsQuery.current.data;
    const appliedValues =
      proxyDimensionManager.mode === DimensionFilterMode.InList
        ? inListValues
        : proxyDimensionManager.mode === DimensionFilterMode.Select
          ? dimensionManager.selectedValues
          : [];
    if (!appliedValues.length) return searchResults;

    return [...new Set([...searchResults, ...appliedValues])];
  });

  let enableSearchCountQuery = $derived(
    Boolean(timeControlsReady) &&
      ((proxyDimensionManager.mode === DimensionFilterMode.Contains &&
        curSearchText.length > 0) ||
        (proxyDimensionManager.mode === DimensionFilterMode.InList &&
          inListValues.length > 0)),
  );

  const allSearchResultsCountQuery = createDimensionSearchCountQuery(
    client,
    () => ({
      manager,
      dimensionName: dimensionManager.name,
      mode: proxyDimensionManager.mode,
      values: inListValues,
      searchText: curSearchText,
      timeStart,
      timeEnd,
      timeDimension,
      enabled: enableSearchCountQuery,
    }),
  );
  let allSearchResultsCount = $derived(allSearchResultsCountQuery.current.data);
  let searchResultCountText = $derived(
    enableSearchCountQuery
      ? proxyDimensionManager.mode === DimensionFilterMode.Contains
        ? `${allSearchResultsCount} results`
        : `${allSearchResultsCount} of ${inListValues.length} matched`
      : "0 results",
  );

  let searchPlaceholder = $derived(
    getSearchPlaceholder(proxyDimensionManager.mode),
  );

  let error = $derived(
    searchResultsQuery.current.error ??
      allSearchResultsCountQuery.current.error,
  );
  let isFetching = $derived(
    searchResultsQuery.current.isFetching ||
      allSearchResultsCountQuery.current.isFetching,
  );

  let showExtraInfo = $derived(
    proxyDimensionManager.mode !== DimensionFilterMode.Select,
  ); // || curSearchText.length > 0; (Add once we have docs)

  let effectiveSelectedValues = $derived(
    getEffectiveSelectedValues(
      proxyDimensionManager.mode,
      proxyDimensionManager.selectedValues,
      correctedSearchResults,
      dimensionManager.selectedValues,
    ),
  );
  let allSelected = $derived(
    Boolean(
      effectiveSelectedValues.length &&
        correctedSearchResults.length === effectiveSelectedValues.length,
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
      correctedSearchResults,
      dimensionManager.selectedValues,
      curSearchText,
    ),
  );

  function handleModeChange(newMode: DimensionFilterMode) {
    switch (newMode) {
      case DimensionFilterMode.Select:
        proxyDimensionManager.setSelectedValues(
          [...proxyDimensionManager.selectedValues],
          proxyDimensionManager.exclude,
        );
        curSearchText = "";
        break;

      case DimensionFilterMode.InList:
        proxyDimensionManager.setInList(
          [...proxyDimensionManager.selectedValues],
          proxyDimensionManager.exclude,
        );
        // The values carried over stay editable as text.
        curSearchText = mergeDimensionSearchValues(
          proxyDimensionManager.selectedValues,
        );
        break;

      case DimensionFilterMode.Contains:
        proxyDimensionManager.setContainsText(
          "",
          proxyDimensionManager.exclude,
        );
        curSearchText = "";
        break;
    }
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      resetSearchText();
      curPinned = pinned;
      curRequired = required;
    } else {
      persistPinnedAndRequired();

      // Select mode applies when the dropdown closes. The other modes only apply through the Apply
      // button, so drop whatever they staged.
      if (proxyDimensionManager.mode === DimensionFilterMode.Select) {
        dimensionManager.apply(proxyDimensionManager);
      } else {
        proxyDimensionManager.apply(dimensionManager);
        resetSearchText();
      }
    }
  }

  /** Resets the input to the applied filter: the list of values in InList mode, the search text otherwise. */
  function resetSearchText() {
    curSearchText =
      dimensionManager.mode === DimensionFilterMode.InList
        ? mergeDimensionSearchValues(dimensionManager.selectedValues)
        : dimensionManager.inputText;
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
    persistPinnedAndRequired();
    // Select mode stages straight onto the proxy; the other two stage in the input.
    switch (proxyDimensionManager.mode) {
      case DimensionFilterMode.InList:
        proxyDimensionManager.setInList(
          [...inListValues],
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
    dimensionManager.apply(proxyDimensionManager);
    if (close) open = false;
  }

  // Closing via the trigger goes through handleOpenChange, but applying closes the dropdown
  // directly, so both paths persist the staged pinned and required states.
  function persistPinnedAndRequired() {
    if (pinned !== curPinned) {
      yamlConfigProvider.togglePinnedFilter(dimensionManager.name);
    }
    if (required !== curRequired) {
      yamlConfigProvider.toggleRequiredFilter(dimensionManager.name);
    }
  }

  function handleItemClick(value: string) {
    proxyDimensionManager.toggleValue(value, false);
  }

  onMount(() => {
    open =
      openOnMount &&
      !dimensionManager.selectedValues?.length &&
      !dimensionManager.inputText;
    resetSearchText();
  });
</script>

<svelte:window
  onkeydown={(e) => {
    if (open && e.key === "Enter") onApply();
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
        suppress={open}
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
          removable={!pinned && !required}
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
              ? inListValues
              : effectiveSelectedValues}
            matchedCount={allSearchResultsCount}
            loading={allSearchResultsCountQuery.current.isFetching}
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
                >{required
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
      {#if yamlConfigProvider.editable}
        <div
          class="flex flex-row items-center justify-between mb-2 pointer-events-auto"
        >
          <b>{dimensionManager.label}</b>

          <div class="flex flex-row items-center gap-x-1">
            <RequiredButton
              required={curRequired}
              onToggleRequired={() => (curRequired = !curRequired)}
            />
            <PinButton
              pinned={curPinned}
              onTogglePin={() => (curPinned = !curPinned)}
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
      {:else}
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
