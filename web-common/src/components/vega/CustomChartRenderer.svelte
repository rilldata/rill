<script lang="ts">
  import { PreviewTable } from "@rilldata/web-common/components/preview-table";
  import VegaLiteRenderer from "@rilldata/web-common/components/vega/VegaLiteRenderer.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import ReconcilingSpinner from "@rilldata/web-common/features/entity-management/ReconcilingSpinner.svelte";
  import { createRuntimeServiceQueryResolver } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { View, VisualizationSpec } from "svelte-vega";

  export let spec: string | undefined = undefined;
  export let metricsSQL: string;
  export let renderer: "canvas" | "svg" = "canvas";
  export let name: string = "Custom Chart";

  let viewVL: View;
  let parsedSpec: VisualizationSpec | null = null;
  let error: string | null = null;
  let rows;
  let tableColumns: VirtualizedTableColumns[];

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

  $: console.log("data", data, tableColumns);
</script>

{#if rows}
  <PreviewTable {rows} columnNames={tableColumns} rowHeight={32} {name} />
{:else if $dataQuery.isLoading}
  <ReconcilingSpinner />
{/if}

{#if spec && error}
  {error}
{:else if data && parsedSpec}
  <VegaLiteRenderer
    {renderer}
    spec={parsedSpec}
    data={{ metrics: data }}
    bind:viewVL
  />
{/if}
