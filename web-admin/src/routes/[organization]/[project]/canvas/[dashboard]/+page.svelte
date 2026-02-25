<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  export let data;

  $: ({
    canvasName,
    project: { id: projectId },
  } = data);

  $: ({ instanceId } = runtimeClient);
</script>

{#key instanceId}
  <CanvasProvider {canvasName} {instanceId} {projectId} showBanner>
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
{/key}
