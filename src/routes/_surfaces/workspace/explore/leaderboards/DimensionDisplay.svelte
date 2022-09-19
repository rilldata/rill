<script lang="ts">
  /**
   * DimensionDisplay.svelte
   * -------------------------
   * Create a table with the selected dimension and measures
   * to be displayed in explore
   */
  import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import DimensionContainer from "$lib/components/dimension/DimensionContainer.svelte";
  import DimensionHeader from "$lib/components/dimension/DimensionHeader.svelte";
  import DimensionTable from "$lib/components/dimension/DimensionTable.svelte";
  import {
    useDimensionFromMetaQuery,
    useMappedFiltersFromMetaQuery,
    useMeasureFromMetaQuery,
    useMeasureNamesFromMetaQuery,
    useMetaQuery,
  } from "$lib/svelte-query/queries/metrics-views/metrics-views-metadata";
  import { useTopListQuery } from "$lib/svelte-query/queries/metrics-views/metrics-views-top-list";
  import { useTotalsQuery } from "$lib/svelte-query/queries/metrics-views/metrics-views-totals";
  import { humanizeGroupByColumns } from "$lib/util/humanize-numbers";

  export let metricsDefId: string;
  export let dimensionId: string;

  $: metaQuery = useMetaQuery(metricsDefId);

  $: dimensionQuery = useDimensionFromMetaQuery(metricsDefId, dimensionId);
  let dimension: DimensionDefinitionEntity;
  $: dimension = dimensionQuery?.data;

  $: leaderboardMeasureId = metricsExplorer?.leaderboardMeasureId;
  $: leaderboardMeasureQuery = useMeasureFromMetaQuery(
    metricsDefId,
    metricsExplorer?.leaderboardMeasureId
  );

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  $: mappedFiltersQuery = useMappedFiltersFromMetaQuery(
    metricsDefId,
    metricsExplorer?.filters
  );

  $: selectedMeasureNames = useMeasureNamesFromMetaQuery(
    metricsDefId,
    metricsExplorer?.selectedMeasureIds
  );

  let activeValues: Array<unknown>;
  $: activeValues =
    metricsExplorer?.filters.include.find((d) => d.name === dimension?.id)
      ?.values ?? [];

  let topListQuery;

  $: allMeasures = $metaQuery.data?.measures;

  $: sortByColumn = $leaderboardMeasureQuery.data?.sqlName;
  $: sortDirection = sortDirection || "desc";

  $: if (
    sortByColumn &&
    sortDirection &&
    leaderboardMeasureId &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    topListQuery = useTopListQuery(metricsDefId, dimensionId, {
      measures: $selectedMeasureNames.data,
      limit: 250,
      offset: 0,
      sort: [
        {
          name: sortByColumn,
          direction: sortDirection,
        },
      ],
      time: {
        start: metricsExplorer?.selectedTimeRange?.start,
        end: metricsExplorer?.selectedTimeRange?.end,
      },
      filter: $mappedFiltersQuery.data,
    });
  }

  let totalsQuery;
  $: if (
    metricsExplorer &&
    metaQuery &&
    $metaQuery.isSuccess &&
    !$metaQuery.isRefetching
  ) {
    totalsQuery = useTotalsQuery(metricsDefId, {
      measures: $selectedMeasureNames.data,
      time: {
        start: metricsExplorer.selectedTimeRange?.start,
        end: metricsExplorer.selectedTimeRange?.end,
      },
    });
  }

  let referenceValues = {};
  $: if ($totalsQuery?.data?.data) {
    allMeasures.map((m) => {
      const isSummableMeasure =
        m?.expression.toLowerCase()?.includes("count(") ||
        m?.expression?.toLowerCase()?.includes("sum(");
      if (isSummableMeasure) {
        referenceValues[m.sqlName] = $totalsQuery.data.data?.[m.sqlName];
      }
    });
  }

  let values = [];
  let columns = [];
  let measureNames = [];

  $: if (!$topListQuery?.isFetching) {
    values = $topListQuery?.data?.data ?? [];

    /* FIX ME
    /* for now getting the column names from the values
    /* in future use the meta field to get column details
    */
    if (values.length) {
      let columnNames = Object.keys(values[0]).sort();

      columnNames = columnNames.filter(
        (name) => name !== dimension?.dimensionColumn
      );
      columnNames.unshift(dimension?.dimensionColumn);
      measureNames = allMeasures.map((m) => m.sqlName);

      columns = columnNames.map((columnName) => {
        if (measureNames.includes(columnName)) {
          const measure = allMeasures.find((m) => m.sqlName === columnName);
          return {
            name: columnName,
            type: "INT",
            label: measure?.label || measure?.expression,
            total: referenceValues[measure.sqlName] || 0,
            enableResize: false,
          };
        } else
          return {
            name: columnName,
            type: "VARCHAR",
            label: dimension?.labelSingle,
            enableResize: true,
          };
      });
    }
  }

  function onSelectItem(event) {
    const label = values[event.detail][dimension?.dimensionColumn];
    metricsExplorerStore.toggleFilter(metricsDefId, dimension?.id, label);
  }

  function onSortByColumn(event) {
    const columnName = event.detail;
    if (!measureNames.includes(columnName)) return;

    if (columnName === sortByColumn) {
      sortDirection = sortDirection === "desc" ? "asc" : "desc";
    } else {
      metricsExplorerStore.setLeaderboardMeasureId(
        metricsDefId,
        allMeasures.find((m) => m.sqlName === columnName)?.id
      );
      sortDirection = "desc";
    }
  }

  $: if (values) {
    const measureFormatSpec = allMeasures.map((m) => {
      return { columnName: m.sqlName, formatPreset: m.formatPreset };
    });
    values = humanizeGroupByColumns(values, measureFormatSpec);
  }
</script>

{#if topListQuery}
  <DimensionContainer>
    <DimensionHeader {metricsDefId} isFetching={$topListQuery?.isFetching} />

    {#if values}
      <DimensionTable
        on:select-item={(event) => onSelectItem(event)}
        on:sort={(event) => onSortByColumn(event)}
        {columns}
        {activeValues}
        rows={values}
        {sortByColumn}
      />
    {/if}
  </DimensionContainer>
{/if}
