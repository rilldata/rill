<script lang="ts">
  import ChartPromptStatusDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptStatusDisplay.svelte";
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { createRuntimeServiceGetChartData } from "@rilldata/web-common/runtime-client/manual-clients";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let filePath: string;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: chartName = extractFileName(filePath);

  const queryClient = useQueryClient();

  $: error = "";
  $: chart = fileArtifact.getResource(queryClient, $runtime.instanceId);
  $: metricsQuery = $chart?.data?.chart?.spec?.resolverProperties;
  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;
  $: data = {};

  $: chartDataQuery = createRuntimeServiceGetChartData(
    queryClient,
    $runtime.instanceId,
    chartName,
    metricsQuery,
  );

  $: if (!$chartDataQuery.isFetching && $chartDataQuery?.data) {
    data = { table: $chartDataQuery?.data };
  }

  let parsedVegaSpec = undefined;
  $: try {
    parsedVegaSpec = vegaSpec ? JSON.parse(vegaSpec) : undefined;
    error = "";
  } catch (e) {
    error = e;
  }
</script>

<div class="m-2 w-1/2">
  <ChartPromptStatusDisplay {chartName}>
    {#if error}
      <p>{error}</p>
    {:else if !parsedVegaSpec}
      <p>Chart not available</p>
    {:else}
      <div class="w-full h-1/2">
        <VegaLiteRenderer {data} spec={parsedVegaSpec} />
      </div>
    {/if}
  </ChartPromptStatusDisplay>
</div>
