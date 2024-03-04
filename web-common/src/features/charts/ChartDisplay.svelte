<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/charts/render/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let chartName: string;

  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;
  $: parsedVegaSpec = vegaSpec ? JSON.parse(vegaSpec) : undefined;
</script>

<div class="m-2 w-1/2">
  <VegaLiteRenderer spec={parsedVegaSpec} />
</div>
