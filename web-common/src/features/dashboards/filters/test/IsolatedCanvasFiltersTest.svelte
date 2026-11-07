<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import CanvasFilterParamsSync from "@rilldata/web-common/features/canvas/CanvasFilterParamsSync.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import CanvasFilters from "@rilldata/web-common/features/canvas/filters/CanvasFilters.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  /**
   * Test component that renders the filter bar the way the report form does: an isolated
   * `CanvasProvider` fed the report's captured state, with no `CanvasDashboardWrapper` and so no
   * `syncStoreWithSource`. `CanvasFilterParamsSync` is the only thing keeping the canvas entity's
   * `expressionFilterManager` in sync, which is where the test reads the filter from.
   *
   * Mirrors `BaseScheduledReportForm`; keep the two in step.
   */
  let {
    canvasName,
    canvasStateOverride,
  }: { canvasName: string; canvasStateOverride: string | undefined } = $props();

  const runtimeClient = useRuntimeClient();
</script>

<!-- The app layout normally supplies the tooltip provider the time controls need. -->
<Tooltip.Provider>
  <CanvasProvider
    {canvasName}
    instanceId={runtimeClient.instanceId}
    isolated
    urlStateOverride={canvasStateOverride}
  >
    <CanvasFilterParamsSync
      {canvasName}
      urlStateOverride={canvasStateOverride}
    />
    <CanvasFilters {canvasName} maxWidth={820} readOnly />
  </CanvasProvider>
</Tooltip.Provider>
