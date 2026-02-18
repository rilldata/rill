<script lang="ts">
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { workspaces as workspaceStore } from "@rilldata/web-common/layout/workspace/workspace-stores";
  import CanvasWorkspace from "@rilldata/web-common/features/workspaces/CanvasWorkspace.svelte";
  import ExploreWorkspace from "@rilldata/web-common/features/workspaces/ExploreWorkspace.svelte";
  import MetricsWorkspace from "@rilldata/web-common/features/workspaces/MetricsWorkspace.svelte";
  import ModelWorkspace from "@rilldata/web-common/features/workspaces/ModelWorkspace.svelte";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  const workspaceComponents = new Map([
    [ResourceKind.Source, ModelWorkspace],
    [ResourceKind.Model, ModelWorkspace],
    [ResourceKind.MetricsView, MetricsWorkspace],
    [ResourceKind.Explore, ExploreWorkspace],
    [ResourceKind.Canvas, CanvasWorkspace],
    [null, null],
    [undefined, null],
  ]);

  export let data: PageData;

  $: ({ fileArtifact } = data);
  $: ({ fileName, resourceName, inferredResourceKind, path } = fileArtifact);

  $: resourceKind = <ResourceKind | undefined>$resourceName?.kind;

  $: workspace = workspaceComponents.get(resourceKind ?? $inferredResourceKind);

  onMount(() => {
    // Force viz mode (no code editor)
    const ws = workspaceStore.get(path);
    ws.view.set("viz");
  });
</script>

<svelte:head>
  <title>Rill | {fileName}</title>
</svelte:head>

<div class="h-full overflow-hidden">
  {#if workspace}
    <svelte:component this={workspace} {fileArtifact} />
  {:else}
    <!-- Fallback if no workspace found -->
    <div class="flex items-center justify-center h-full">
      <p class="text-fg-muted">Unable to load editor</p>
    </div>
  {/if}
</div>
