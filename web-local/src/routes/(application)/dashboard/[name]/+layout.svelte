<script lang="ts">
  import { page } from "$app/stores";
  import { WorkspaceContainer } from "@rilldata/web-local/lib/components/workspace";
  import ExploreWorkspaceHeader from "@rilldata/web-local/lib/components/workspace/explore/workspace-header/ExploreWorkspaceHeader.svelte";
  export let data;
  $: entry = data.entry;
  $: metricViewName = $page.params.name;
  $: displayName = entry?.metricsView?.label || entry?.metricsview?.name;
</script>

<WorkspaceContainer
  assetID={metricViewName}
  bgClass="bg-white"
  inspector={false}
>
  <ExploreWorkspaceHeader slot="header" {displayName} {metricViewName} />
  <svelte:fragment slot="body">
    <slot />
  </svelte:fragment>
</WorkspaceContainer>
