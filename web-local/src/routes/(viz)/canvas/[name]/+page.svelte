<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";

  export let data: PageData;

  $: ({ instanceId } = $runtime);
  $: ({ canvasName } = data);
</script>

<CanvasProvider {canvasName} {instanceId} showBanner>
  <CanvasDashboardEmbed {canvasName} />
</CanvasProvider>
