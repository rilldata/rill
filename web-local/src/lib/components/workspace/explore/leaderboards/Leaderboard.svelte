<script lang="ts">
  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
    useMetaQuery,
  } from "@rilldata/web-local/lib/svelte-query/dashboards";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";

  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    MetricsViewDimension,
    MetricsViewMeasure,
    useRuntimeServiceMetricsViewToplist,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import {
    humanizeGroupValues,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "../../../../util/humanize-numbers";
  import LeaderboardHeader from "../../../leaderboard/LeaderboardHeader.svelte";
  import LeaderboardList from "../../../leaderboard/LeaderboardList.svelte";
  import LeaderboardListItem from "../../../leaderboard/LeaderboardListItem.svelte";
  import DimensionLeaderboardEntrySet from "./DimensionLeaderboardEntrySet.svelte";
  import { hasDefinedTimeSeries } from "../utils";

  export let metricViewName: string;
  export let dimensionName: string;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;

  export let formatPreset: NicelyFormattedTypes;
  export let leaderboardFormatScale: ShortHandSymbols;
  export let isSummableMeasure = false;

  export let slice = 7;
  export let seeMoreSlice = 50;
  let seeMore = false;

  const dispatch = createEventDispatcher();

  $: metaQuery = useMetaQuery($runtimeStore.instanceId, metricViewName);

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let filterExcludeMode: boolean;
  $: filterExcludeMode =
    metricsExplorer?.dimensionFilterExcludeMode.get(dimensionName) ?? false;
  let filterKey: "exclude" | "include";
  $: filterKey = filterExcludeMode ? "exclude" : "include";

  $: dimensionQuery = useMetaDimension(
    $runtimeStore.instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;
  $: displayName = dimension.label || dimension.name;

  $: measureQuery = useMetaMeasure(
    $runtimeStore.instanceId,
    metricViewName,
    metricsExplorer?.leaderboardMeasureName
  );
  let measure: MetricsViewMeasure;
  $: measure = $measureQuery?.data;

  $: filterForDimension = getFilterForDimension(
    metricsExplorer?.filters,
    dimensionName
  );

  let activeValues: Array<unknown>;
  $: activeValues =
    metricsExplorer?.filters[filterKey]?.find((d) => d.name === dimension?.name)
      ?.in ?? [];
  $: atLeastOneActive = !!activeValues?.length;

  let hasTimeSeries;

  $: if (metaQuery && $metaQuery.isSuccess && !$metaQuery.isRefetching) {
    hasTimeSeries = hasDefinedTimeSeries($metaQuery.data);
  }

  function setLeaderboardValues(values) {
    dispatch("leaderboard-value", {
      dimensionName,
      values,
    });
  }

  function toggleFilterMode() {
    metricsExplorerStore.toggleFilterMode(metricViewName, dimensionName);
  }

  function selectDimension(dimensionName) {
    metricsExplorerStore.setMetricDimensionName(metricViewName, dimensionName);
  }

  let topListQuery;

  $: if (
    measure?.name &&
    metricsExplorer &&
    $metaQuery?.isSuccess &&
    !$metaQuery?.isRefetching
  ) {
    let topListParams = {
      dimensionName: dimensionName,
      measureNames: [measure.name],
      limit: "250",
      offset: "0",
      sort: [
        {
          name: measure.name,
          ascending: false,
        },
      ],
      filter: filterForDimension,
    };

    if (hasTimeSeries) {
      topListParams = {
        ...topListParams,
        ...{
          timeStart: metricsExplorer.selectedTimeRange?.start,
          timeEnd: metricsExplorer.selectedTimeRange?.end,
        },
      };
    }

    topListQuery = useRuntimeServiceMetricsViewToplist(
      $runtimeStore.instanceId,
      metricViewName,
      topListParams
    );
  }

  let values = [];

  /** replace data after fetched. */
  $: if (!$topListQuery?.isFetching) {
    values =
      $topListQuery?.data?.data.map((val) => ({
        value: val[measure?.name],
        label: val[dimension?.name],
      })) ?? [];
    setLeaderboardValues(values);
  }
  /** figure out how many selected values are currently hidden */
  // $: hiddenSelectedValues = values.filter((di, i) => {
  //   return activeValues.includes(di.label) && i > slice - 1 && !seeMore;
  // });

  $: if (values) {
    values = formatPreset
      ? humanizeGroupValues(values, formatPreset, {
          scale: leaderboardFormatScale,
        })
      : humanizeGroupValues(values, NicelyFormattedTypes.HUMANIZE, {
          scale: leaderboardFormatScale,
        });
  }

  // get all values that are selected but not visible.
  // we'll put these at the bottom w/ a divider.
  $: selectedValuesThatAreBelowTheFold = activeValues
    ?.filter((label) => {
      return (
        // the value is visible within the fold.
        !values.slice(0, !seeMore ? slice : seeMoreSlice).some((value) => {
          return value.label === label;
        })
      );
    })
    .map((label) => {
      const existingValue = values.find((value) => value.label === label);
      // return the existing value, or if it does not exist, just return the label.
      // FIX ME return values for label which are not in the query
      return existingValue ? { ...existingValue } : { label };
    })
    .sort((a, b) => {
      return b.value - a.value;
    });

  let hovered: boolean;
</script>

{#if topListQuery}
  <div
    style:width="315px"
    on:mouseenter={() => (hovered = true)}
    on:mouseleave={() => (hovered = false)}
  >
    <LeaderboardHeader
      isFetching={$topListQuery.isFetching}
      {displayName}
      on:toggle-filter-mode={toggleFilterMode}
      {filterExcludeMode}
      {hovered}
      dimensionDescription={dimension?.description}
      on:click={() => selectDimension(dimensionName)}
    />

    {#if values}
      <LeaderboardList>
        <!-- place the leaderboard entries that are above the fold here -->
        <DimensionLeaderboardEntrySet
          loading={$topListQuery?.isFetching}
          values={values.slice(0, !seeMore ? slice : seeMoreSlice)}
          {activeValues}
          {filterExcludeMode}
          {atLeastOneActive}
          {referenceValue}
          {isSummableMeasure}
          on:select-item
        />
        <!-- place the selected values that are not above the fold here -->
        {#if selectedValuesThatAreBelowTheFold?.length}
          <hr />
          <DimensionLeaderboardEntrySet
            loading={$topListQuery?.isFetching}
            values={selectedValuesThatAreBelowTheFold}
            {activeValues}
            {filterExcludeMode}
            {atLeastOneActive}
            {referenceValue}
            {isSummableMeasure}
            on:select-item
          />
          <hr />
        {/if}
        {#if $topListQuery?.isError}
          <div class="text-red-500">
            {$topListQuery?.error}
          </div>
        {:else if values.length === 0}
          <div class="p-1 ui-copy-disabled">no available values</div>
        {/if}

        {#if values.length > slice}
          <Tooltip location="right">
            <LeaderboardListItem
              value={0}
              color="=ui-label"
              on:click={() => selectDimension(dimensionName)}
            >
              <div class="ui-copy-muted" slot="title">(Expand Table)</div>
            </LeaderboardListItem>
            <TooltipContent slot="tooltip-content"
              >Expand dimension to see more values</TooltipContent
            >
          </Tooltip>
        {/if}
      </LeaderboardList>
    {/if}
  </div>
{/if}
