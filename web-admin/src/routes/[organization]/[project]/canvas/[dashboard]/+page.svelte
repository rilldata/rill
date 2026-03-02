<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let data;

  $: ({
    canvasName,
    project: { id: projectId },
  } = data);

  const runtimeClient = useRuntimeClient();
</script>

{#key runtimeClient.instanceId}
  <CanvasProvider
    {canvasName}
    instanceId={runtimeClient.instanceId}
    {projectId}
    showBanner
  >
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
{/key}
