<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import { cancelDashboardQueries } from "@rilldata/web-common/features/dashboards/dashboard-queries";
  import {
    getFilterForDimension,
    useMetaDimension,
    useMetaMeasure,
    useMetaQuery,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import {
    createQueryServiceMetricsViewToplist,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    MetricsViewMeasure,
    V1MetricsViewToplistResponseDataItem,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { SortDirection } from "../proto-state/derived-types";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { humanizeGroupByColumns, FormatPreset } from "../humanize-numbers";
  import {
    computeComparisonValues,
    computePercentOfTotal,
    getComparisonProperties,
    getFilterForComparisonTable,
    updateFilterOnSearch,
  } from "./dimension-table-utils";
  import DimensionHeader from "./DimensionHeader.svelte";
  import DimensionTable from "./DimensionTable.svelte";

  export let metricViewName: string;
  export let dimensionName: string;

  let searchText = "";

  const queryClient = useQueryClient();

  $: instanceId = $runtime.instanceId;

  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: dimensionQuery = useMetaDimension(
    instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;
  $: dimensionColumn = dimension?.column || dimension?.name;

  $: dashboardStore = useDashboardStore(metricViewName);

  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: leaderboardMeasureName = $dashboardStore?.leaderboardMeasureName;
  $: isBeingCompared =
    $dashboardStore?.selectedComparisonDimension === dimensionName;
  $: leaderboardMeasureQuery = useMetaMeasure(
    instanceId,
    metricViewName,
    leaderboardMeasureName
  );

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterForDimension = getFilterForDimension(
    $dashboardStore?.filters,
    dimensionName
  );

  $: selectedMeasureNames = $dashboardStore?.selectedMeasureNames;

  let selectedValues: Array<unknown>;
  $: selectedValues =
    (excludeMode
      ? $dashboardStore?.filters.exclude.find((d) => d.name === dimension?.name)
          ?.in
      : $dashboardStore?.filters.include.find((d) => d.name === dimension?.name)
          ?.in) ?? [];

  $: allMeasures = $metaQuery.data?.measures.filter((m) =>
    $dashboardStore?.visibleMeasureKeys.has(m.name)
  );

  $: sortAscending = $dashboardStore.sortDirection === SortDirection.ASCENDING;

  $: metricTimeSeries = useModelHasTimeSeries(instanceId, metricViewName);
  $: hasTimeSeries = $metricTimeSeries.data;

  $: filterSet = updateFilterOnSearch(
    filterForDimension,
    searchText,
    dimension?.name
  );
  $: topListQuery = createQueryServiceMetricsViewToplist(
    instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
      filter: filterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: leaderboardMeasureName,
          ascending: sortAscending,
        },
      ],
    },
    {
      query: {
        enabled:
          $timeControlsStore.ready && !!filterSet && !!leaderboardMeasureName,
      },
    }
  );

  // Compose the comparison /toplist query
  $: displayComparison = $timeControlsStore.showComparison;
  $: comparisonFilterSet = getFilterForComparisonTable(
    filterForDimension,
    dimensionName,
    dimensionColumn,
    $topListQuery?.data?.data
  );
  $: comparisonTopListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [leaderboardMeasureName],
      timeStart: $timeControlsStore.comparisonTimeStart,
      timeEnd: $timeControlsStore.comparisonTimeEnd,
      filter: comparisonFilterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: leaderboardMeasureName,
          ascending: sortAscending,
        },
      ],
    },
    {
      query: {
        enabled: Boolean(
          $timeControlsStore.showComparison && !!comparisonFilterSet
        ),
      },
    }
  );

  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: $timeControlsStore.timeStart,
      timeEnd: $timeControlsStore.timeEnd,
    },
    {
      query: {
        enabled: $timeControlsStore.ready,
      },
    }
  );

  let referenceValues = {};
  $: if ($totalsQuery?.data?.data) {
    allMeasures.map((m) => {
      const isSummableMeasure =
        m?.expression.toLowerCase()?.includes("count(") ||
        m?.expression?.toLowerCase()?.includes("sum(");
      if (isSummableMeasure) {
        referenceValues[m.name] = $totalsQuery.data.data?.[m.name];
      }
    });
  }

  let values: V1MetricsViewToplistResponseDataItem[] = [];
  let columns = [];
  let measureNames = [];

  $: if (!$topListQuery?.isFetching && dimension) {
    let columnsMeta = $topListQuery?.data?.meta || [];
    values = $topListQuery?.data?.data ?? [];

    let columnNames: Array<string> = columnsMeta
      .map((c) => c.name)
      .filter(
        (name) =>
          name !== dimensionColumn &&
          $dashboardStore.visibleMeasureKeys.has(name)
      );

    const selectedMeasure = allMeasures.find(
      (m) => m.name === leaderboardMeasureName
    );
    const sortByColumnIndex = columnNames.indexOf(leaderboardMeasureName);
    // Add comparison columns if available
    let percentOfTotalSpliceIndex = 1;
    if (displayComparison) {
      percentOfTotalSpliceIndex = 2;
      columnNames.splice(
        sortByColumnIndex + 1,
        0,
        `${leaderboardMeasureName}_delta`
      );

      // Only push percentage delta column if selected measure is not a percentage
      if (selectedMeasure?.format != FormatPreset.PERCENTAGE) {
        percentOfTotalSpliceIndex = 3;
        columnNames.splice(
          sortByColumnIndex + 2,
          0,
          `${leaderboardMeasureName}_delta_perc`
        );
      }
    }
    if (validPercentOfTotal) {
      columnNames.splice(
        sortByColumnIndex + percentOfTotalSpliceIndex,
        0,
        `${leaderboardMeasureName}_percent_of_total`
      );
    }

    // Make dimension the first column
    columnNames.unshift(dimensionColumn);
    measureNames = allMeasures.map((m) => m.name);

    columns = columnNames
      .map((columnName) => {
        if (measureNames.includes(columnName)) {
          // Handle all regular measures
          const measure = allMeasures.find((m) => m.name === columnName);
          return {
            name: columnName,
            type: "INT",
            label: measure?.label || measure?.expression,
            description: measure?.description,
            total: referenceValues[measure.name] || 0,
            enableResize: false,
            format: measure?.format,
          };
        } else if (columnName === dimensionColumn) {
          // Handle dimension column
          return {
            name: columnName,
            type: "VARCHAR",
            label: dimension?.label,
            enableResize: true,
          };
        } else if (selectedMeasure) {
          // Handle delta and delta_perc
          const comparison = getComparisonProperties(
            columnName,
            selectedMeasure
          );
          return {
            name: columnName,
            type: comparison.type,
            label: comparison.label,
            description: comparison.description,
            enableResize: false,
            format: comparison.format,
          };
        }
        return undefined;
      })
      .filter((column) => !!column);
  }

  function onSelectItem(event) {
    const label = values[event.detail][dimensionColumn];
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimension?.name, label);
  }

  function onSortByColumn(event) {
    const columnName = event.detail;
    if (!measureNames.includes(columnName)) return;

    if (columnName === leaderboardMeasureName) {
      metricsExplorerStore.toggleSort(metricViewName);
    } else {
      metricsExplorerStore.setLeaderboardMeasureName(
        metricViewName,
        columnName
      );
      metricsExplorerStore.setSortDescending(metricViewName);
    }
  }

  function toggleComparisonDimension(dimensionName, isBeingCompared) {
    metricsExplorerStore.setComparisonDimension(
      metricViewName,
      isBeingCompared ? undefined : dimensionName
    );
  }

  $: if ($comparisonTopListQuery?.data && values.length && displayComparison) {
    values = computeComparisonValues(
      $comparisonTopListQuery?.data,
      values,
      dimensionName,
      dimensionColumn,
      leaderboardMeasureName
    );
  }

  $: validPercentOfTotal = (
    $leaderboardMeasureQuery?.data as MetricsViewMeasure
  )?.validPercentOfTotal;

  $: if (validPercentOfTotal && values.length && leaderboardMeasureName) {
    const referenceValue = $totalsQuery.data?.data?.[leaderboardMeasureName];
    values = computePercentOfTotal(
      values,
      referenceValue,
      leaderboardMeasureName
    );
  }

  $: if (values) {
    const measureFormatSpec = columns?.map((column) => {
      return {
        columnName: column.name,
        formatPreset: column.format as FormatPreset,
      };
    });
    if (measureFormatSpec) {
      values = humanizeGroupByColumns(values, measureFormatSpec);
    }
  }
</script>

{#if topListQuery}
  <div class="h-full flex flex-col" style:min-width="365px">
    <div class="flex-none" style:height="50px">
      <DimensionHeader
        {metricViewName}
        {dimensionName}
        {excludeMode}
        isFetching={$topListQuery?.isFetching}
        on:search={(event) => {
          searchText = event.detail;
        }}
      />
    </div>

    {#if values && columns.length}
      <div class="grow" style="overflow-y: hidden;">
        <DimensionTable
          on:select-item={(event) => onSelectItem(event)}
          on:sort={(event) => onSortByColumn(event)}
          on:toggle-dimension-comparison={() =>
            toggleComparisonDimension(dimensionName, isBeingCompared)}
          {sortAscending}
          dimensionName={dimensionColumn}
          {isBeingCompared}
          {columns}
          {selectedValues}
          rows={values}
          sortByColumn={leaderboardMeasureName}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
