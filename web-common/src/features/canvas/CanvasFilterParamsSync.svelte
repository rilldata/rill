<script lang="ts">
  import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { page } from "$app/state";

  /**
   * Applies the canvas state a `CanvasProvider` was given to that canvas's filter manager, once
   * the metrics views are ready.
   *
   * `CanvasEntity.onUrlChange` holds the filter params back while the metrics view specs are
   * still loading, since the filter manager would drop every filter it cannot resolve and then
   * skip the identical params once the specs arrive. Surfaces that render
   * `CanvasDashboardWrapper` recover through its `syncStoreWithSource`; this component is for the
   * ones that do not, i.e. the read-only filter bar in the report form. Without it a report saved
   * from the edit dialog loses the filters it was captured with.
   *
   * Temporary, alongside `CanvasProvider` itself. Renders nothing.
   */
  let {
    canvasName,
    urlStateOverride = undefined,
  }: {
    canvasName: string;
    // Must match the `urlStateOverride` given to the enclosing `CanvasProvider`.
    urlStateOverride?: string | undefined;
  } = $props();

  const runtimeClient = useRuntimeClient();

  let { canvasEntity } = $derived(
    getCanvasStore(canvasName, runtimeClient.instanceId),
  );

  // Mirrors the `effectiveUrl` the provider applied: the override when it has one,
  // the page url otherwise.
  let searchParams = $derived(
    urlStateOverride === undefined
      ? page.url.searchParams
      : new URLSearchParams(urlStateOverride),
  );

  $effect(() => {
    if (!canvasEntity.dashboardProvider.metricsViewsProvider.ready) return;
    canvasEntity.expressionFilterManager.setUrlParams(searchParams);
  });
</script>
