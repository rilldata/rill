<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders a canvas the way the canvas dashboard page does, without the
   * page's URL-driven state: a `CanvasProvider` around a `CanvasDashboardEmbed`.
   */
  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();
</script>

<!-- The app layout normally supplies the tooltip provider. -->
<Tooltip.Provider>
  <CanvasProvider {canvasName} instanceId={runtimeClient.instanceId} isolated>
    <CanvasDashboardEmbed {canvasName} navigationEnabled={false} />
  </CanvasProvider>
</Tooltip.Provider>
