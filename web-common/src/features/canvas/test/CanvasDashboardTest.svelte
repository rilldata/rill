<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders a canvas dashboard the way the Rill Cloud canvas page does.
   */
  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();
</script>

<!-- The app layout supplies the tooltip provider, which the filter bar needs. -->
<Tooltip.Provider>
  <CanvasProvider {canvasName} instanceId={runtimeClient.instanceId}>
    <CanvasDashboardEmbed {canvasName} />
  </CanvasProvider>
</Tooltip.Provider>
