<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ instanceId } = $runtime);
  $: ({ canvasName } = data);
</script>

<svelte:head>
  <title>Rill | {canvasName}</title>
</svelte:head>

{#key instanceId}
  <div class="h-full overflow-hidden">
    <CanvasProvider {canvasName} {instanceId} showBanner>
      <CanvasDashboardEmbed {canvasName} />
    </CanvasProvider>
  </div>
{/key}
