<script lang="ts">
  import FieldSwitcher from "@rilldata/web-common/components/forms/FieldSwitcher.svelte";
  import PreviewTable from "@rilldata/web-common/components/preview-table/PreviewTable.svelte";
  import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { createRuntimeServiceQueryResolver } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View, VisualizationSpec } from "svelte-vega";
  import { derived } from "svelte/store";

  export let spec: string | undefined = undefined;
  export let metricsSQL: string[] = [];
  export let renderer: "canvas" | "svg" = "svg";
  export let showDataTable = false;
  export let name: string = "Custom Chart";

  const viewOptions = ["Chart", "Data"];

  let viewVL: View;
  let parsedSpec: VisualizationSpec | null = null;
  let error: string | null = null;
  let rows;
  let tableColumns: VirtualizedTableColumns[];
  let selectedView = 0; // 0 = Chart, 1 = Data
  let selectedTable = 0; // For switching between tables if multiple queries

  $: instanceId = $runtime.instanceId;

  $: dataQueries = metricsSQL.map((sql) =>
    createRuntimeServiceQueryResolver(
      instanceId,
      {
        resolver: "metrics_sql",
        resolverProperties: {
          sql,
        },
      },
      {
        query: {
          enabled: !!sql,
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

  $: vegaData = $combinedResults.reduce((acc, result, idx) => {
    acc[`query${idx + 1}`] = result.data;
    return acc;
  }, {});

  $: try {
    if (typeof spec === "string" && spec !== "") {
      parsedSpec = JSON.parse(spec) as VisualizationSpec;
    }
  } catch (e: unknown) {
    error = JSON.stringify(e);
  }

  // Table data for the selected query
  $: rows = $combinedResults[selectedTable]?.data;
  $: tableColumns = $combinedResults[selectedTable]?.tableSchema;
</script>

<div class="flex flex-col gap-2 h-full">
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
  <div class="flex-1 flex flex-col min-h-0 min-w-0">
    {#if selectedView === 0}
      <div class="flex-1">
        {#if !spec}
          <div class="text-red-500 items-center justify-center">
            No spec provided
          </div>
        {:else if error}
          <div class="text-red-500 items-center justify-center">
            {error}
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
