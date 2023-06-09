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
    useModelAllTimeRange,
    useModelHasTimeSeries,
  } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    createQueryServiceMetricsViewToplist,
    createQueryServiceMetricsViewTotals,
    MetricsViewDimension,
    MetricsViewFilterCond,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getTimeComparisonParametersForComponent } from "../../../lib/time/comparisons";
  import { DEFAULT_TIME_RANGES } from "../../../lib/time/config";
  import type { TimeComparisonOption } from "../../../lib/time/types";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import {
    humanizeGroupByColumns,
    NicelyFormattedTypes,
  } from "../humanize-numbers";
  import {
    computeComparisonValues,
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

  $: dashboardStore = useDashboardStore(metricViewName);

  $: leaderboardMeasureName = $dashboardStore?.leaderboardMeasureName;
  $: leaderboardMeasureQuery = useMetaMeasure(
    instanceId,
    metricViewName,
    leaderboardMeasureName
  );

  let excludeValues: Array<MetricsViewFilterCond>;
  $: excludeValues = $dashboardStore?.filters.exclude;

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

  $: allMeasures = $metaQuery.data?.measures;

  $: sortByColumn = $leaderboardMeasureQuery.data?.name;
  $: sortDirection = sortDirection || "desc";

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
      timeStart: timeStart,
      timeEnd: timeEnd,
      filter: filterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: sortByColumn,
          ascending: sortDirection === "asc" ? true : false,
        },
      ],
    },
    {
      query: {
        enabled:
          (hasTimeSeries ? !!timeStart && !!timeEnd : true) &&
          !!filterSet &&
          !!sortByColumn &&
          !!sortDirection,
      },
    }
  );

  $: allTimeRangeQuery = useModelAllTimeRange(
    $runtime.instanceId,
    $metaQuery.data.model,
    $metaQuery.data.timeDimension,
    {
      query: {
        enabled: !!$metaQuery.data.timeDimension,
      },
    }
  );
  $: allTimeRange = $allTimeRangeQuery?.data;

  // the timeRangeName is the key to a selected time range's associated presets.
  $: timeRangeName = $dashboardStore?.selectedTimeRange?.name;

  // Compose the comparison /toplist query
  $: displayComparison =
    $dashboardStore?.showComparison &&
    comparisonTimeRange.isComparisonRangeAvailable;
  $: comparisonTimeRange = getTimeComparisonParametersForComponent(
    ($dashboardStore?.selectedComparisonTimeRange
      ?.name as TimeComparisonOption) ||
      (DEFAULT_TIME_RANGES[timeRangeName]
        .defaultComparison as TimeComparisonOption),
    allTimeRange?.start,
    allTimeRange?.end,
    $dashboardStore.selectedTimeRange.start,
    $dashboardStore.selectedTimeRange.end
  );
  $: comparisonTimeStart = isFinite(comparisonTimeRange?.start?.getTime())
    ? comparisonTimeRange.start.toISOString()
    : undefined;
  $: comparisonTimeEnd = isFinite(comparisonTimeRange?.end?.getTime())
    ? comparisonTimeRange.end.toISOString()
    : undefined;
  $: comparisonFilterSet = getFilterForComparisonTable(
    filterForDimension,
    dimensionName,
    $topListQuery?.data?.data
  );
  $: comparisonTopListQuery = createQueryServiceMetricsViewToplist(
    $runtime.instanceId,
    metricViewName,
    {
      dimensionName: dimensionName,
      measureNames: [sortByColumn],
      timeStart: comparisonTimeStart,
      timeEnd: comparisonTimeEnd,
      filter: comparisonFilterSet,
      limit: "250",
      offset: "0",
      sort: [
        {
          name: sortByColumn,
          ascending: sortDirection === "asc" ? true : false,
        },
      ],
    },
    {
      query: {
        enabled:
          displayComparison &&
          !!comparisonTimeStart &&
          !!comparisonTimeEnd &&
          !!comparisonFilterSet,
      },
    }
  );

  $: timeStart = $dashboardStore.selectedTimeRange?.start?.toISOString();
  $: timeEnd = $dashboardStore.selectedTimeRange?.end?.toISOString();
  $: totalsQuery = createQueryServiceMetricsViewTotals(
    instanceId,
    metricViewName,
    {
      measureNames: selectedMeasureNames,
      timeStart: timeStart,
      timeEnd: timeEnd,
    },
    {
      query: {
        enabled: hasTimeSeries ? !!timeStart && !!timeEnd : true,
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

  let values = [];
  let columns = [];
  let measureNames = [];

  $: if (!$topListQuery?.isFetching && dimension) {
    let columnsMeta = $topListQuery?.data?.meta || [];
    values = $topListQuery?.data?.data ?? [];

    let columnNames: Array<string> = columnsMeta
      .map((c) => c.name)
      .filter((name) => name !== dimension?.name);

    const selectedMeasure = allMeasures.find((m) => m.name === sortByColumn);
    const sortByColumnIndex = columnNames.indexOf(sortByColumn);
    // Add comparison columns if available
    if (displayComparison) {
      columnNames.splice(sortByColumnIndex + 1, 0, `${sortByColumn}_delta`);

      // Only push percentage delta column if selected measure is not a percentage
      if (selectedMeasure?.format != NicelyFormattedTypes.PERCENTAGE) {
        columnNames.splice(
          sortByColumnIndex + 2,
          0,
          `${sortByColumn}_delta_perc`
        );
      }
    }

    // Make dimension the first column
    columnNames.unshift(dimension?.name);
    measureNames = allMeasures.map((m) => m.name);

    columns = columnNames.map((columnName) => {
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
      } else if (columnName === dimension?.name) {
        // Handle dimension column
        return {
          name: columnName,
          type: "VARCHAR",
          label: dimension?.label,
          enableResize: true,
        };
      } else {
        // Handle delta and delta_perc
        const comparison = getComparisonProperties(columnName, selectedMeasure);
        return {
          name: columnName,
          type: comparison.type,
          label: comparison.label,
          description: comparison.description,
          enableResize: false,
          format: comparison.format,
        };
      }
    });
  }

  function onSelectItem(event) {
    const label = values[event.detail][dimension?.name];
    cancelDashboardQueries(queryClient, metricViewName);
    metricsExplorerStore.toggleFilter(metricViewName, dimension?.name, label);
  }

  function onSortByColumn(event) {
    const columnName = event.detail;
    if (!measureNames.includes(columnName)) return;

    if (columnName === sortByColumn) {
      sortDirection = sortDirection === "desc" ? "asc" : "desc";
    } else {
      metricsExplorerStore.setLeaderboardMeasureName(
        metricViewName,
        columnName
      );
      sortDirection = "desc";
    }
  }

  $: if ($comparisonTopListQuery?.data && values.length && displayComparison) {
    values = computeComparisonValues(
      $comparisonTopListQuery?.data,
      values,
      dimensionName,
      leaderboardMeasureName
    );
  }

  $: if (values) {
    const measureFormatSpec = columns?.map((column) => {
      return {
        columnName: column.name,
        formatPreset: column.format as NicelyFormattedTypes,
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
          dimensionName={dimension?.name}
          {columns}
          {selectedValues}
          rows={values}
          {sortByColumn}
          {excludeMode}
        />
      </div>
    {/if}
  </div>
{/if}
