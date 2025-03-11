<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";

  export let instanceId: string;
  export let canvasName: string;

  $: canvasQuery = useResource(instanceId, canvasName, ResourceKind.Canvas);

  $: resource = $canvasQuery.data;
</script>

<StateManagersProvider {canvasName}>
  <CanvasThemeProvider>
    <CanvasDashboardEmbed {resource} />
  </CanvasThemeProvider>
</StateManagersProvider>
