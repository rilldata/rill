<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { errorStore } from "../../components/errors/error-store";

  export let instanceId: string;
  export let canvasName: string;
  export let navigationEnabled: boolean = true;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: ({ isSuccess, isError, error, data } = $canvasQuery);
  $: isCanvasNotFound = isError && error?.response?.status === 404;

  // We check for canvas.state.validSpec instead of meta.reconcileError. validSpec persists
  // from previous valid canvases, allowing display even when the current canvas spec is invalid
  // and a meta.reconcileError exists.
  $: isCanvasErrored = !data?.canvas?.state?.validSpec;

  // If no dashboard is found, show a 404 page
  $: if (isCanvasNotFound) {
    errorStore.set({
      statusCode: 404,
      header: "Explore not found",
      body: `The Explore dashboard you requested could not be found. Please check that you provided the name of a working dashboard.`,
    });
  }

  $: resource = $canvasQuery.data;
</script>

{#if isSuccess}
  {#if isCanvasErrored}
    <br /> Canvas Error <br />
  {:else if data}
    <CanvasDashboardEmbed {resource} {navigationEnabled} {canvasName} />
  {/if}
{/if}
