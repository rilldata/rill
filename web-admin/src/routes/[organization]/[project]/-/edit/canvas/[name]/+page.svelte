<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  export let data: PageData;

  const client = useRuntimeClient();

  $: ({ canvasName } = data);
</script>

<svelte:head>
  <title>Rill | {canvasName}</title>
</svelte:head>

{#key client.instanceId}
  <div class="h-full overflow-hidden">
    <CanvasProvider {canvasName} instanceId={client.instanceId} showBanner>
      <CanvasDashboardEmbed {canvasName} />
    </CanvasProvider>
  </div>
{/key}
