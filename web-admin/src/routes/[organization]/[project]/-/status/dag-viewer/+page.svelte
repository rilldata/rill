<script lang="ts">
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";

  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);
  $: summaryBasePath = `/${$page.params.organization}/${$page.params.project}/-/status/dag-viewer`;
</script>

<div class="size-full flex flex-col gap-y-4">
  <section class="flex flex-col gap-y-2">
    <h2 class="text-lg font-medium">DAG Viewer</h2>
    <p class="text-sm text-fg-secondary">
      Visualize dependencies between sources, models, and dashboards.
    </p>
  </section>
  <GraphContainer {seeds} {summaryBasePath} />
</div>
