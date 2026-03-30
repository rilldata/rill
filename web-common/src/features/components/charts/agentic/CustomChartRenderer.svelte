<script lang="ts">
  import type { PartialMessage, Struct } from "@bufbuild/protobuf";
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import {
    createRuntimeServiceQueryResolver,
    type V1Expression,
    type V1TimeRange,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { View, VisualizationSpec } from "svelte-vega";
  import { derived } from "svelte/store";
  import { convertV1ExpressionToMapstructure } from "./expression-utils";

  export let spec: string | undefined = undefined;
  export let metricsSQL: string[] = [];
  export let renderer: "canvas" | "svg" = "svg";
  export let whereFilter: V1Expression | undefined = undefined;
  export let timeRange: V1TimeRange | undefined = undefined;
  export let showDataTable = false;
  export let name: string = "Custom Chart";

  const viewOptions = ["Chart", "Data"];

  let viewVL: View;
  let parsedSpec: VisualizationSpec | null = null;
  let error: string | null = null;
  let tableColumns: VirtualizedTableColumns[];
  let selectedView = 0; // 0 = Chart, 1 = Data
  let selectedTable = 0; // For switching between tables if multiple queries

  const runtimeClient = useRuntimeClient();

  // Create a unique key that includes whereFilter and timeRange to ensure queries are invalidated when they change
  $: filterKey = JSON.stringify({ whereFilter, timeRange });

  // Only enable queries when the time range has resolved
  $: hasValidTimeRange = !!timeRange?.start && !!timeRange?.end;

  // Create queries that are reactive to whereFilter changes
  $: dataQueries = metricsSQL.map((sql, index) =>
    createRuntimeServiceQueryResolver(
      runtimeClient,
      {
        resolver: "metrics_sql",
        resolverProperties: {
          sql,
          ...(whereFilter?.cond?.exprs?.length
            ? {
                additional_where:
                  convertV1ExpressionToMapstructure(whereFilter),
              }
            : {}),
          ...(timeRange ? { additional_time_range: timeRange } : {}),
        } as unknown as PartialMessage<Struct>,
      },
      {
        query: {
          enabled: !!sql && hasValidTimeRange,
          queryKey: [`metrics_sql`, name, index, sql, filterKey],
        },
      },
    ),
  );

  $: combinedResults = derived(dataQueries, ($dataQueries) =>
    $dataQueries.map((query) => ({
      data: query.data?.data,
      tableSchema: query.data?.schema?.fields?.map((field) => ({
        name: field.name,
        type: field.type?.code,
      })) as VirtualizedTableColumns[],
      isSuccess: query.isSuccess,
      isLoading: query.isLoading,
      error: query.error,
    })),
  );

  $: vegaData = $combinedResults.reduce<Record<string, unknown>>(
    (acc, result, idx) => {
      acc[`query${idx + 1}`] = result.data;
      return acc;
    },
    {},
  );

  $: try {
    if (typeof spec === "string" && spec !== "") {
      parsedSpec = JSON.parse(spec) as VisualizationSpec;
    }
  } catch (e: unknown) {
    error = JSON.stringify(e);
  }

  // Check for any query errors
  $: queryError = $combinedResults.find((r) => r.error)?.error;

  // Table data for the selected query
  $: rows = $combinedResults[selectedTable]?.data;
  $: tableColumns = $combinedResults[selectedTable]?.tableSchema;
</script>

<div class="flex flex-col gap-2 h-full pt-1">
  {#if showDataTable}
    <div class="flex flex-row items-center gap-2 p-1">
      <FieldSwitcher
        fields={viewOptions}
        selected={selectedView}
        onClick={(i) => (selectedView = i)}
        small={true}
      />
      {#if metricsSQL.length > 1 && selectedView === 1}
        <FieldSwitcher
          fields={metricsSQL.map((_, i) => `Query ${i + 1}`)}
          selected={selectedTable}
          onClick={(i) => (selectedTable = i)}
          small={true}
        />
      {/if}
    </div>
  {/if}
  <div class="size-full flex flex-col overflow-hidden">
    {#if selectedView === 0}
      <div class="size-full">
        {#if !spec}
          <div class="text-red-500 items-center justify-center">
            No spec provided
          </div>
        {:else if error}
          <div class="text-red-500 items-center justify-center">
            {error}
          </div>
        {:else if queryError}
          <div class="text-red-500 items-center justify-center">
            {queryError.message || "Error loading data"}
          </div>
        {:else if rows && parsedSpec}
          <VegaLiteRenderer
            {renderer}
            spec={parsedSpec}
            canvasDashboard
            config={getRillTheme(true)}
            data={vegaData}
            bind:viewVL
          />
        {/if}
      </div>
    {:else}
      <div class="flex-1 min-h-0 min-w-0">
        {#if $combinedResults[selectedTable]?.isSuccess && rows}
          <PreviewTable
            {rows}
            columnNames={tableColumns}
            rowHeight={32}
            {name}
          />
        {:else if $combinedResults[selectedTable]?.isLoading}
          <ReconcilingSpinner />
        {:else if $combinedResults[selectedTable]?.error}
          <div class="text-red-500">
            {$combinedResults[selectedTable].error?.message ||
              "Error loading data"}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>
