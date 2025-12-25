<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  export let data;

  $: ({
    canvasName,
    project: {
      project: { id: projectId },
    },
  } = data);

  const instanceId = httpClient.getInstanceId();
</script>

<CanvasProvider {canvasName} {instanceId} {projectId} showBanner>
  <CanvasDashboardEmbed {canvasName} />
</CanvasProvider>
