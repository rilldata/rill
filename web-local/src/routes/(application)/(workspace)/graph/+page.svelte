<script lang="ts">
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import { page } from "$app/stores";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";

  // Parse URL parameters using new API (kind/resource instead of seed)
  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <div slot="header" class="header">
    <div class="header-title">
      <div class="header-left">
        <h1>Project graph</h1>
      </div>
    </div>
    <p>Visualize dependencies between sources, models, dashboards, and more.</p>
  </div>

  <div slot="body" class="graph-wrapper">
    <GraphContainer {seeds} />
  </div>
</WorkspaceContainer>

<style lang="postcss">
  .header {
    @apply px-4 pt-3 pb-2;
  }

  .header h1 {
    @apply text-lg font-semibold text-foreground;
  }

  .header-title {
    @apply flex items-center justify-between;
  }
  /* seed-label removed */

  .header p {
    @apply text-sm text-gray-500 mt-1;
  }

  .graph-wrapper {
    @apply h-full w-full;
  }
</style>
