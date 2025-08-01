<script lang="ts">
  import DimensionFilterChipBody from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterChipBody.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants";
  import {
    useAllSearchResultsCount,
    useDimensionSearch,
  } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-values";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let name: string;
  export let metricsViewNames: string[];
  export let label: string;
  export let mode: DimensionFilterMode;
  export let values: string[];
  export let inputText: string | undefined;
  export let isInclude: boolean;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;

  $: ({ instanceId } = $runtime);

  $: effectiveLabel = isInclude ? label : `Exclude ${label}`;
  $: sanitisedSearchText = inputText?.replace(/^%/, "").replace(/%$/, "");

  $: enableSearchQuery =
    mode === DimensionFilterMode.Select ||
    (mode === DimensionFilterMode.Contains && !!inputText) ||
    (mode === DimensionFilterMode.InList && values.length > 0);
  $: searchResultsQuery = useDimensionSearch(
    instanceId,
    metricsViewNames,
    name,
    {
      mode,
      values,
      searchText: inputText ?? "",
      timeStart,
      timeEnd,
      enabled: enableSearchQuery,
    },
  );
  $: ({ data: searchResults, isFetching: isFetchingFromSearchResults } =
    $searchResultsQuery);
  $: correctedSearchResults = enableSearchQuery ? searchResults : [];
  $: enableSearchCountQuery =
    (mode === DimensionFilterMode.Contains && !!inputText) ||
    (mode === DimensionFilterMode.InList && values.length > 0);
  $: allSearchResultsCountQuery = useAllSearchResultsCount(
    instanceId,
    metricsViewNames,
    name,
    {
      mode,
      values,
      searchText: inputText ?? "",
      timeStart,
      timeEnd,
      enabled: enableSearchCountQuery,
    },
  );
  $: ({
    data: allSearchResultsCount,
    isFetching: isFetchingFromAllSearchResultsCount,
  } = $allSearchResultsCountQuery);

  $: isFetching =
    isFetchingFromSearchResults ?? isFetchingFromAllSearchResultsCount;
  $: effectiveSelectedValues =
    mode !== DimensionFilterMode.InList
      ? values
      : (correctedSearchResults ?? []);
</script>

<Chip type="dimension" label={effectiveLabel} readOnly exclude={!isInclude}>
  <DimensionFilterChipBody
    slot="body"
    label={effectiveLabel}
    values={mode === DimensionFilterMode.InList
      ? values
      : effectiveSelectedValues}
    matchedCount={allSearchResultsCount}
    loading={isFetching}
    search={mode === DimensionFilterMode.Contains
      ? sanitisedSearchText
      : undefined}
  />
</Chip>
