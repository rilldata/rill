<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

  export let data: PageData;

  $: ({ instanceId } = $runtime);
  $: ({ canvasName } = data);
</script>

{#key instanceId}
  <div class="flex h-full overflow-hidden">
    <div class="flex-1 overflow-hidden">
      <CanvasProvider {canvasName} {instanceId} showBanner>
        <CanvasDashboardEmbed {canvasName} />
      </CanvasProvider>
    </div>
    <DashboardChat kind={ResourceKind.Canvas} />
  </div>
{/key}
