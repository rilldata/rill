<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/canvas/state-managers/StateManagersProvider.svelte";
  import CanvasStateProvider from "@rilldata/web-common/features/canvas/stores/CanvasStateProvider.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: canvasName = data.dashboardName;
</script>

<StateManagersProvider {canvasName}>
  <CanvasStateProvider>
    <CanvasThemeProvider>
      <CanvasDashboardEmbed items={data.dashboard.spec?.items || []} />
    </CanvasThemeProvider>
  </CanvasStateProvider>
</StateManagersProvider>
