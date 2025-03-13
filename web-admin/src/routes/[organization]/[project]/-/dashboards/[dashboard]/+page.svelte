<script lang="ts">
  import { page } from "$app/stores";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";

  $: ({ instanceId } = $runtime);
  $: canvasName = $page.params.dashboard;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: canvasResource = $canvasQuery.data;
</script>

<StateManagersProvider {canvasName}>
  <CanvasThemeProvider>
    <CanvasDashboardEmbed resource={canvasResource} />
  </CanvasThemeProvider>
</StateManagersProvider>
