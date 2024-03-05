<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

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
