<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import GeneratingCanvasMessage from "@rilldata/web-common/features/canvas/ai-generation/GeneratingCanvasMessage.svelte";
  import { generatingCanvas } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas.ts";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ instanceId } = $runtime);
  $: ({ canvasName } = data);
</script>

{#if $generatingCanvas}
  <GeneratingCanvasMessage />
{:else}
  {#key instanceId}
    <CanvasProvider {canvasName} {instanceId} showBanner>
      <CanvasDashboardEmbed {canvasName} />
    </CanvasProvider>
  {/key}
{/if}
