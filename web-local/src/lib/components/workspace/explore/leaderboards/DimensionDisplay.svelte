<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
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
  import DimensionContainer from "../../../dimension/DimensionContainer.svelte";
  import DimensionHeader from "../../../dimension/DimensionHeader.svelte";
  import DimensionTable from "../../../dimension/DimensionTable.svelte";
  import {
    humanizeGroupByColumns,
    NicelyFormattedTypes,
  } from "../../../../util/humanize-numbers";
  import {
    MetricsViewDimension,
    MetricsViewFilterCond,
    useRuntimeServiceMetricsViewToplist,
    useRuntimeServiceMetricsViewTotals,
  } from "@rilldata/web-common/runtime-client";

  export let metricViewName: string;
  export let dimensionName: string;

  let searchText = "";

  $: instanceId = $runtimeStore.instanceId;
  $: addNull = "null".includes(searchText);

  $: metaQuery = useMetaQuery(instanceId, metricViewName);

  $: dimensionQuery = useMetaDimension(
    instanceId,
    metricViewName,
    dimensionName
  );
  let dimension: MetricsViewDimension;
  $: dimension = $dimensionQuery?.data;

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  $: leaderboardMeasureName = metricsExplorer?.leaderboardMeasureName;
  $: leaderboardMeasureQuery = useMetaMeasure(
    instanceId,
    metricViewName,
    leaderboardMeasureName
  );

  let excludeValues: Array<MetricsViewFilterCond>;
  $: excludeValues = metricsExplorer?.filters.exclude;

  $: excludeMode =
    metricsExplorer?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: filterForDimension = getFilterForDimension(
    metricsExplorer?.filters,
    dimensionName
  );

  $: selectedMeasureNames = metricsExplorer?.selectedMeasureNames;

  let selectedValues: Array<unknown>;
  $: selectedValues =
    (excludeMode
      ? metricsExplorer?.filters.exclude.find((d) => d.name === dimension?.name)
          ?.in
      : metricsExplorer?.filters.include.find((d) => d.name === dimension?.name)
          ?.in) ?? [];

  let topListQuery;

  $: allMeasures = $metaQuery.data?.measures;

  $: sortByColumn = $leaderboardMeasureQuery.data?.name;
  $: sortDirection = sortDirection || "desc";

  $: if (
    sortByColumn &&
    sortDirection &&
    leaderboardMeasureName &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    let filterData = JSON.parse(JSON.stringify(filterForDimension));

    if (searchText !== "") {
      let foundDimension = false;

      filterData["include"].forEach((filter) => {
        if (filter.name == dimension?.name) {
          filter.like = [`%${searchText}%`];
          foundDimension = true;
          if (addNull) filter.in.push(null);
        }
      });

      if (!foundDimension) {
        filterData["include"].push({
          name: dimension?.name,
          in: addNull ? [null] : [],
          like: [`%${searchText}%`],
        });
      }
    } else {
      filterData["include"] = filterData["include"].filter((f) => f.in.length);
      filterData["include"].forEach((f) => {
        delete f.like;
      });
    }

    topListQuery = useRuntimeServiceMetricsViewToplist(
      instanceId,
      metricViewName,
      {
        dimensionName: dimensionName,
        measureNames: selectedMeasureNames,
        limit: "250",
        offset: "0",
        sort: [
          {
            name: sortByColumn,
            ascending: sortDirection === "asc" ? true : false,
          },
        ],
        timeStart: metricsExplorer.selectedTimeRange?.start,
        timeEnd: metricsExplorer.selectedTimeRange?.end,
        filter: filterData,
      }
    );
  }

  let totalsQuery;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    totalsQuery = useRuntimeServiceMetricsViewTotals(
      instanceId,
      metricViewName,
      {
        measureNames: selectedMeasureNames,
        timeStart: metricsExplorer.selectedTimeRange?.start,
        timeEnd: metricsExplorer.selectedTimeRange?.end,
      }
    );
  }

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
    values = $topListQuery?.data?.data ?? [];

    /* FIX ME
    /* for now getting the column names from the values
    /* in future use the meta field to get column details
    */
    if (values.length) {
      let columnNames = Object.keys(values[0]).sort();

      columnNames = columnNames.filter((name) => name !== dimension?.name);
      columnNames.unshift(dimension?.name);
      measureNames = allMeasures.map((m) => m.name);

      columns = columnNames.map((columnName) => {
        if (measureNames.includes(columnName)) {
          const measure = allMeasures.find((m) => m.name === columnName);
          return {
            name: columnName,
            type: "INT",
            label: measure?.label || measure?.expression,
            total: referenceValues[measure.name] || 0,
            enableResize: false,
          };
        } else
          return {
            name: columnName,
            type: "VARCHAR",
            label: dimension?.label,
            enableResize: true,
          };
      });
    }
  }

  function onSelectItem(event) {
    const label = values[event.detail][dimension?.name];
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

  $: if (values) {
    const measureFormatSpec = allMeasures?.map((m) => {
      return {
        columnName: m.name,
        formatPreset: m.format as NicelyFormattedTypes,
      };
    });
    values = humanizeGroupByColumns(values, measureFormatSpec);
  }
</script>

{#if topListQuery}
  <DimensionContainer>
    <DimensionHeader
      {metricViewName}
      {dimensionName}
      {excludeMode}
      isFetching={$topListQuery?.isFetching}
      on:search={(event) => {
        searchText = event.detail;
      }}
    />

    {#if values && columns.length}
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
    {/if}
  </DimensionContainer>
{/if}
