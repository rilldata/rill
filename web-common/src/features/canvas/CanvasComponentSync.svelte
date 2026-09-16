<script lang="ts">
  import type { BaseCanvasComponent } from "@rilldata/web-common/features/canvas/components/BaseCanvasComponent.ts";
  import { syncStoreWithSource } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";

  let {
    component,
  }: {
    component: BaseCanvasComponent;
  } = $props();

  // TODO: move to CanvasComponent after migrating it to svelte5
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    component.expressionFilters,
    async () => component.syncExpressionFilters(),
    () => component.expressionFilters.metricsViewsProvider.ready,
  );
  // svelte-ignore state_referenced_locally
  syncStoreWithSource(
    component.timeFilters,
    async () => component.syncTimeFilters(),
    () =>
      component.expressionFilters.metricsViewsProvider.ready &&
      component.timeFilters.ready,
  );
</script>
