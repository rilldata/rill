<script lang="ts">
  import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
  import { Chip } from "@rilldata/web-common/components/chip";
  import DimensionFilterChipBody from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterChipBody.svelte";
  import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
  import {
    createDimensionSearchCountQuery,
    createDimensionSearchQuery,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/queries.svelte.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let {
    manager,
    dimensionManager,
    yamlConfigProvider,
    timeStart,
    timeEnd,
    showPinned,
  }: {
    manager: ExpressionFilterManager;
    dimensionManager: DimensionFilterManager;
    yamlConfigProvider: YAMLConfigProvider;
    timeStart: string | undefined;
    timeEnd: string | undefined;
    showPinned: boolean;
  } = $props();

  let effectiveLabel = $derived(
    dimensionManager.exclude
      ? `Exclude ${dimensionManager.label}`
      : dimensionManager.label,
  );

  let pinned = $derived(
    Boolean(
      showPinned && yamlConfigProvider.pinnedFilters[dimensionManager.name],
    ),
  );
  let required = $derived(
    Boolean(
      showPinned && yamlConfigProvider.requiredFilters[dimensionManager.name],
    ),
  );
  let missingRequired = $derived(Boolean(required && !dimensionManager.expr));

  const runtimeClient = useRuntimeClient();
  let enableSearchQuery = $derived(
    dimensionManager.mode === DimensionFilterMode.Select ||
      (dimensionManager.mode === DimensionFilterMode.Contains &&
        !!dimensionManager.inputText) ||
      (dimensionManager.mode === DimensionFilterMode.InList &&
        dimensionManager.selectedValues.length > 0),
  );
  const searchResultsQuery = createDimensionSearchQuery(runtimeClient, () => ({
    manager,
    dimensionName: dimensionManager.name,
    mode: dimensionManager.mode,
    values: dimensionManager.selectedValues,
    searchText: dimensionManager.inputText,
    timeStart,
    timeEnd,
    enabled: enableSearchQuery,
  }));
  let { data: searchResults, isFetching: isFetchingFromSearchResults } =
    $derived(searchResultsQuery.current);
  let correctedSearchResults = $derived(enableSearchQuery ? searchResults : []);

  let enableSearchCountQuery = $derived(
    (dimensionManager.mode === DimensionFilterMode.Contains &&
      !!dimensionManager.inputText) ||
      (dimensionManager.mode === DimensionFilterMode.InList &&
        dimensionManager.selectedValues.length > 0),
  );
  const allSearchResultsCountQuery = createDimensionSearchCountQuery(
    runtimeClient,
    () => ({
      manager,
      dimensionName: dimensionManager.name,
      mode: dimensionManager.mode,
      values: dimensionManager.selectedValues,
      searchText: dimensionManager.inputText,
      timeStart,
      timeEnd,
      enabled: enableSearchCountQuery,
    }),
  );
  let {
    data: allSearchResultsCount,
    isFetching: isFetchingFromAllSearchResultsCount,
  } = $derived(allSearchResultsCountQuery.current);

  let isFetching = $derived(
    isFetchingFromSearchResults || isFetchingFromAllSearchResultsCount,
  );
  let effectiveSelectedValues = $derived(
    dimensionManager.mode !== DimensionFilterMode.InList
      ? dimensionManager.selectedValues
      : (correctedSearchResults ?? []),
  );
</script>

<Chip
  type="dimension"
  theme
  label={effectiveLabel}
  readOnly
  exclude={dimensionManager.exclude}
  showPinnedIcon={pinned}
  gray={pinned && dimensionManager.selectedValues.length === 0}
  error={missingRequired}
>
  <DimensionFilterChipBody
    slot="body"
    label={effectiveLabel}
    values={dimensionManager.mode === DimensionFilterMode.InList
      ? dimensionManager.selectedValues
      : effectiveSelectedValues}
    matchedCount={allSearchResultsCount}
    loading={isFetching}
    search={dimensionManager.mode === DimensionFilterMode.Contains
      ? dimensionManager.inputText
      : undefined}
  />
</Chip>
