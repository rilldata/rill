<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getQueryFromDataSpec } from "./data/queryMapper";

  const queryClient = useQueryClient();

  export let chartName: string;
  $: error = "";
  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;
  $: queryName = $chart?.data?.chart?.spec?.queryName;
  $: queryArgs = $chart?.data?.chart?.spec?.queryArgsJson;

  let dataStore;

  $: if (queryName && queryArgs) {
    dataStore = getQueryFromDataSpec(
      $runtime.instanceId,
      queryClient,
      queryName,
      queryArgs,
    );
  }

  $: data = {};

  $: if ($dataStore?.data?.data) {
    data = { table: $dataStore?.data?.data };
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
  {#if error}
    <p>{error}</p>
  {:else if !parsedVegaSpec}
    <p>Chart not available</p>
  {:else}
    <VegaLiteRenderer {data} spec={parsedVegaSpec} />
  {/if}
</div>
