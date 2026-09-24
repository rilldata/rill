<script lang="ts">
  import CanvasDashboardWrapper from "@rilldata/web-common/features/canvas/CanvasDashboardWrapper.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { DEFAULT_DASHBOARD_WIDTH } from "@rilldata/web-common/features/canvas/layout-util";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders the filter bar the way a canvas dashboard does: `CanvasProvider`
   * resolves the canvas and builds the store, and `CanvasDashboardWrapper` holds the filter bar
   * and the url sync. The filter state lands on the canvas entity's `expressionFilterManager`,
   * which is where the test reads it from.
   */
  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();
</script>

<CanvasProvider {canvasName} instanceId={runtimeClient.instanceId}>
  <!-- `CanvasDashboardEmbed` reads both of these off the spec; they are fixed here. -->
  <CanvasDashboardWrapper
    {canvasName}
    maxWidth={DEFAULT_DASHBOARD_WIDTH}
    filtersEnabled
  >
    <div>Dashboard loaded!</div>
  </CanvasDashboardWrapper>
</CanvasProvider>
