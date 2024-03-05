<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { getAllErrorsForFile } from "@rilldata/web-common/features/entity-management/resources-store";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let chartName: string;
  let error = "";

  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;
  let parsedVegaSpec = undefined;
  $: try {
    parsedVegaSpec = vegaSpec ? JSON.parse(vegaSpec) : undefined;
  } catch (e) {
    error = e;
  }

  const queryClient = useQueryClient();
  $: allErrors = getAllErrorsForFile(
    queryClient,
    $runtime.instanceId,
    `/charts/${chartName}.yaml`,
  );
  $: console.log($allErrors);
</script>

<div class="m-2 w-1/2">
  {#if error}
    <p>{error}</p>
  {:else if !parsedVegaSpec}
    <p>Chart not available</p>
  {:else}
    <VegaLiteRenderer spec={parsedVegaSpec} />
  {/if}
</div>
