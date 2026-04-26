<script lang="ts">
  import { page } from "$app/stores";
  import GraphContainer from "./navigation/GraphContainer.svelte";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "./navigation/seed-parser";
  import WorkspaceContainer from "../../layout/workspace/WorkspaceContainer.svelte";

  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);
</script>

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
