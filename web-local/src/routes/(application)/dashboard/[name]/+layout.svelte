<script lang="ts">
  import { page } from "$app/stores";
  import ModelInspector from "@rilldata/web-common/features/models/workspace/inspector/ModelInspector.svelte";
  import {
    Inspector,
    WorkspaceContainer,
  } from "@rilldata/web-local/lib/components/workspace";
  import DashboardWorkspaceHeader from "@rilldata/web-local/lib/components/workspace/explore/workspace-header/DashboardWorkspaceHeader.svelte";
  export let data;
  $: entry = data.entry;
  $: metricViewName = $page.params.name;
  $: displayName = entry?.metricsView?.label || entry?.metricsView?.name;
  $: view =
    $page?.url?.pathname === `/dashboard/${metricViewName}`
      ? "dashboard"
      : $page?.url?.pathname === `/dashboard/${metricViewName}/edit`
      ? "config"
      : "model";
</script>

<WorkspaceContainer
  assetID={metricViewName}
  bgClass="bg-white"
  viewHasInspector={true}
  inspector={view !== "dashboard"}
>
  <DashboardWorkspaceHeader slot="header" {displayName} {metricViewName} />
  <slot />
  {#if view !== "dashboard"}
    <Inspector>
      <ModelInspector modelName={entry?.metricsView?.model} />
    </Inspector>
  {/if}
</WorkspaceContainer>
