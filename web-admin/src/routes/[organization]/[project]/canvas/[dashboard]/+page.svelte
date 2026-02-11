<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  export let data;

  $: ({
    canvasName,
    project: { id: projectId },
  } = data);

  $: ({ instanceId } = $runtime);
</script>

{#key instanceId}
  <CanvasProvider {canvasName} {instanceId} {projectId} showBanner>
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
{/key}
