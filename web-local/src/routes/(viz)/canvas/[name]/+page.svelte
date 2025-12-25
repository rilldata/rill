<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";

  export let data: PageData;

  const instanceId = httpClient.getInstanceId();
  $: ({ canvasName } = data);
</script>

<CanvasProvider {canvasName} {instanceId} showBanner>
  <CanvasDashboardEmbed {canvasName} />
</CanvasProvider>
