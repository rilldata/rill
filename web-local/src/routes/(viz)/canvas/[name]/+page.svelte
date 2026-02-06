<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ instanceId } = $runtime);
  $: ({ canvasName } = data);
</script>

{#key instanceId}
  <CanvasProvider {canvasName} {instanceId} showBanner>
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
{/key}
