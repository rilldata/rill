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

  export let spec: string | undefined = undefined;
  export let metricsSQL: string;
  export let renderer: "canvas" | "svg" = "svg";
  export let showDataTable = false;
  export let name: string = "Custom Chart";

  let viewVL: View;
  let parsedSpec: VisualizationSpec | null = null;
  let error: string | null = null;
  let rows;
  let tableColumns: VirtualizedTableColumns[];
  let selectedView = 0; // 0 = Chart, 1 = Table
  const viewOptions = ["Chart", "Table"];

  $: instanceId = $runtime.instanceId;

  $: dataQuery = createRuntimeServiceQueryResolver(instanceId, {
    resolver: "metrics_sql",
    resolverProperties: {
      sql: metricsSQL,
    },
  });

  $: data = $dataQuery.data?.data;

  $: try {
    if (typeof spec === "string") {
      parsedSpec = JSON.parse(spec) as VisualizationSpec;
    } else {
      parsedSpec = spec ?? null;
    }
  } catch (e: unknown) {
    error = JSON.stringify(e);
  }

  $: {
    if ($dataQuery.isSuccess) {
      rows = $dataQuery.data.data;
      tableColumns = $dataQuery.data.schema?.fields?.map((field) => ({
        name: field.name,
        type: field.type?.code,
      })) as VirtualizedTableColumns[];
    }
  }
</script>

<div class="flex flex-col gap-2 h-full">
  {#if showDataTable}
    <div class="flex flex-row justify-end items-center">
      <FieldSwitcher
        fields={viewOptions}
        selected={selectedView}
        onClick={(i) => (selectedView = i)}
        small={true}
      />
    </div>
  {/if}
  <div class="flex-1 flex flex-col min-h-0 min-w-0">
    {#if selectedView === 0}
      <div class="flex-1">
        {#if spec && error}
          {error}
        {:else if data && parsedSpec}
          <VegaLiteRenderer
            {renderer}
            spec={parsedSpec}
            config={getRillTheme(true)}
            data={{ metrics: data }}
            bind:viewVL
          />
        {/if}
      </div>
    {:else}
      <div class="flex-1 min-h-0 min-w-0">
        {#if rows}
          <PreviewTable
            {rows}
            columnNames={tableColumns}
            rowHeight={32}
            {name}
          />
        {:else if $dataQuery.isLoading}
          <ReconcilingSpinner />
        {/if}
      </div>
    {/if}
  </div>
</div>
