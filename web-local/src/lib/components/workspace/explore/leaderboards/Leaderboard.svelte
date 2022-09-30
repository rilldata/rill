<script lang="ts">
  import { RootConfig } from "@rilldata/web-local/common/config/RootConfig";

  /**
   * Leaderboard.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import type { DimensionDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import { EntityStatus } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { MeasureDefinitionEntity } from "@rilldata/web-local/common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
  import FilterInclude from "@rilldata/web-local/lib/components/icons/FilterInclude.svelte";
  import FilterRemove from "@rilldata/web-local/lib/components/icons/FilterRemove.svelte";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "../../../../application-state-stores/explorer-stores";
  import LeaderboardContainer from "../../../leaderboard/LeaderboardContainer.svelte";
  import LeaderboardHeader from "../../../leaderboard/LeaderboardHeader.svelte";
  import LeaderboardList from "../../../leaderboard/LeaderboardList.svelte";
  import LeaderboardListItem from "../../../leaderboard/LeaderboardListItem.svelte";
  import Spinner from "../../../Spinner.svelte";
  import Shortcut from "../../../tooltip/Shortcut.svelte";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "../../../tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "../../../tooltip/TooltipTitle.svelte";
  import {
    useMetaDimension,
    useMetaMappedFilters,
    useMetaMeasure,
    useMetaQuery,
  } from "../../../../svelte-query/queries/metrics-views/metadata";
  import { useTopListQuery } from "../../../../svelte-query/queries/metrics-views/top-list";
  import { slideRight } from "../../../../transitions";
  import {
    humanizeGroupValues,
    NicelyFormattedTypes,
    ShortHandSymbols,
  } from "../../../../util/humanize-numbers";
  import { createEventDispatcher, getContext } from "svelte";
  import { getDisplayName } from "../utils";
  import DimensionLeaderboardEntrySet from "./DimensionLeaderboardEntrySet.svelte";

  export let metricsDefId: string;
  export let dimensionId: string;
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

  const config = getContext<RootConfig>("config");

  $: metaQuery = useMetaQuery(config, metricsDefId);

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let filterExcludeMode: boolean;
  $: filterExcludeMode =
    metricsExplorer?.dimensionFilterExcludeMode.get(dimensionId) ?? false;
  let filterKey: string;
  $: filterKey = filterExcludeMode ? "exclude" : "include";
  let otherFilterKey: string;
  $: otherFilterKey = filterExcludeMode ? "include" : "exclude";

  $: dimensionQuery = useMetaDimension(config, metricsDefId, dimensionId);
  let dimension: DimensionDefinitionEntity;
  $: dimension = $dimensionQuery?.data;
  let displayName: string;
  // TODO: select based on label?
  $: displayName = getDisplayName(dimension);

  $: measureQuery = useMetaMeasure(
    config,
    metricsDefId,
    metricsExplorer?.leaderboardMeasureId
  );
  let measure: MeasureDefinitionEntity;
  $: measure = $measureQuery?.data;

  $: mappedFiltersQuery = useMetaMappedFilters(
    config,
    metricsDefId,
    metricsExplorer?.filters
  );

  let activeValues: Array<unknown>;
  $: activeValues =
    metricsExplorer?.filters[filterKey]?.find((d) => d.name === dimension?.id)
      ?.values ?? [];
  $: atLeastOneActive = !!activeValues?.length;

  function setLeaderboardValues(values) {
    dispatch("leaderboard-value", {
      dimensionId,
      values,
    });
  }

  function toggleFilterExcludeMode() {
    metricsExplorerStore.toggleFilterExcludeMode(metricsDefId, dimensionId);
  }

  function selectDimension(dimensionId) {
    metricsExplorerStore.setMetricDimensionId(metricsDefId, dimensionId);
  }

  let topListQuery;

  $: if (
    measure?.id &&
    metricsExplorer &&
    $metaQuery?.isSuccess &&
    !$metaQuery?.isRefetching
  ) {
    topListQuery = useTopListQuery(config, metricsDefId, dimensionId, {
      measures: [measure.sqlName],
      limit: 15,
      offset: 0,
      sort: [
        {
          name: measure.sqlName,
          direction: "desc",
        },
      ],
      time: {
        start: metricsExplorer.selectedTimeRange?.start,
        end: metricsExplorer.selectedTimeRange?.end,
      },
      filter: $mappedFiltersQuery.data,
    });
  }

  let values = [];

  /** replace data after fetched. */
  $: if (!$topListQuery?.isFetching) {
    values =
      $topListQuery?.data?.data.map((val) => ({
        value: val[measure?.sqlName],
        label: val[dimension?.dimensionColumn],
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
  <LeaderboardContainer
    focused={atLeastOneActive}
    on:mouseenter={() => {
      hovered = true;
    }}
    on:mouseleave={() => {
      hovered = false;
    }}
  >
    <LeaderboardHeader
      isActive={atLeastOneActive}
      on:click={() => selectDimension(dimensionId)}
    >
      <div
        slot="title"
        class:text-gray-500={atLeastOneActive}
        class:italic={atLeastOneActive}
      >
        <Tooltip location="top" distance={16}>
          <div class="flex flex-row gap-x-2 items-center">
            {#if $topListQuery?.isFetching}
              <div transition:slideRight|local={{ leftOffset: 8 }}>
                <Spinner size="16px" status={EntityStatus.Running} />
              </div>
            {/if}
            {displayName}
          </div>
          <TooltipContent slot="tooltip-content">
            <TooltipTitle>
              <svelte:fragment slot="name">
                {displayName}
              </svelte:fragment>
              <svelte:fragment slot="description" />
            </TooltipTitle>
            <TooltipShortcutContainer>
              <div>
                {#if dimension?.description}
                  {dimension.description}
                {:else}
                  the leaderboard metrics for {displayName}
                {/if}
              </div>
              <Shortcut />
              <div>Expand leaderboard</div>
              <Shortcut>Click</Shortcut>
            </TooltipShortcutContainer>
          </TooltipContent>
        </Tooltip>
      </div>
      <div slot="right">
        {#if hovered}
          <Tooltip location="top" distance={16}>
            <div on:click|stopPropagation={toggleFilterExcludeMode}>
              {#if filterExcludeMode}<FilterRemove
                  size="16px"
                />{:else}<FilterInclude size="16px" />{/if}
            </div>
            <TooltipContent slot="tooltip-content">
              <TooltipTitle>
                <svelte:fragment slot="name">
                  filter {filterKey} mode
                </svelte:fragment>
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>toggle {otherFilterKey} mode</div>
                <Shortcut>Click</Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>
        {/if}
      </div>
    </LeaderboardHeader>

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
          <div class="p-1 italic text-gray-500">no available values</div>
        {/if}

        {#if values.length > slice}
          <Tooltip location="right">
            <LeaderboardListItem
              value={0}
              color="bg-gray-100"
              on:click={() => selectDimension(dimensionId)}
            >
              <div class="italic text-gray-500" slot="title">
                (Expand Table)
              </div>
            </LeaderboardListItem>
            <TooltipContent slot="tooltip-content"
              >Expand dimension to see more values</TooltipContent
            >
          </Tooltip>
        {/if}
      </LeaderboardList>
    {/if}
  </LeaderboardContainer>
{/if}
