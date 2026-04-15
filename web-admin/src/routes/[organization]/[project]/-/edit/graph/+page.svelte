<script lang="ts">
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";

  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);
</script>

<svelte:head>
  <title>Rill | Project graph</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <div slot="header" class="header">
    <div class="header-title">
      <h1>Project graph</h1>
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
    @apply text-lg font-semibold text-fg-primary;
  }

  .header-title {
    @apply flex items-center justify-between;
  }

  .header p {
    @apply text-sm text-fg-secondary mt-1;
  }

  .graph-wrapper {
    @apply h-full w-full;
  }
</style>
