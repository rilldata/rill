<script lang="ts">
  import VegaLiteRenderer from "@rilldata/web-common/features/custom-dashboards/VegaLiteRenderer.svelte";
  import { useChart } from "@rilldata/web-common/features/charts/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy, onMount } from "svelte";
  import type { VisualizationSpec } from "svelte-vega";

  const observer = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const { width, height } = entry.contentRect;
      clientHeight = height;
      clientWidth = width;
    }
  });

  export let chartName: string;

  let clientHeight: number;
  let clientWidth: number;
  let container: HTMLDivElement;
  let error: unknown = "";
  let parsedVegaSpec: VisualizationSpec | undefined = undefined;

  $: chart = useChart($runtime.instanceId, chartName);

  $: vegaSpec = $chart?.data?.chart?.spec?.vegaLiteSpec;

  $: try {
    parsedVegaSpec = vegaSpec
      ? (JSON.parse(vegaSpec) as VisualizationSpec)
      : undefined;
    error = "";
  } catch (e: unknown) {
    error = e;
  }

  onMount(() => {
    observer.observe(container);
  });

  onDestroy(() => {
    observer.disconnect();
  });
</script>

<div class="h-full w-full rounded-sm overflow-hidden" bind:this={container}>
  {#if error}
    <p>{error}</p>
  {:else if !parsedVegaSpec}
    <p>Chart not available</p>
  {:else}
    <VegaLiteRenderer
      spec={parsedVegaSpec}
      height={clientHeight - 31}
      width={clientWidth}
    />
  {/if}
</div>
