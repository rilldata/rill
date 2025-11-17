<script lang="ts">
  import {
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import WorkspaceCrumb from "./WorkspaceCrumb.svelte";
import ResourceGraphOverlay from "@rilldata/web-common/features/resource-graph/ResourceGraphOverlay.svelte";
import { GitBranch } from "lucide-svelte";
import { ALLOWED_FOR_GRAPH } from "@rilldata/web-common/features/resource-graph/seed-utils";

  export let resource: V1Resource | undefined;
  export let filePath: string;

  $: ({ instanceId } = $runtime);

  $: resourceKind = resource?.meta?.name?.kind as ResourceKind | undefined;
  $: resourceName = resource?.meta?.name?.name;

  $: resourcesQuery = createRuntimeServiceListResources(instanceId, undefined, {
    query: { retry: 2, refetchOnMount: true },
  });
$: allResources = $resourcesQuery.data?.resources ?? [];
$: resourcesLoading = $resourcesQuery.isLoading;
$: resourcesError = $resourcesQuery.error
  ? "Failed to load project resources."
  : null;

let graphOverlayOpen = false;
$: graphSupported =
  resourceKind && ALLOWED_FOR_GRAPH.has(resourceKind)
    ? true
    : false;

  $: lateralResources = allResources.filter(({ meta }) => {
    if (meta?.name?.name === resourceName && meta?.name?.kind === resourceKind)
      return true;
    if (!meta?.refs?.length) return false;

    return meta?.refs?.every(({ name, kind }) =>
      resource?.meta?.refs?.find(
        (ref) => ref?.name === name && ref?.kind === kind,
      ),
    );
  });
</script>

<nav class="resource-breadcrumbs">
  <div class="resource-breadcrumbs__track">
    <div class="resource-breadcrumbs__crumbs">
      <WorkspaceCrumb
        selectedResource={resource}
        resources={lateralResources}
        {allResources}
        {filePath}
        current
      />
    </div>
    {#if resource && graphSupported}
      <button
        type="button"
        class="graph-trigger"
        on:click={() => (graphOverlayOpen = true)}
        aria-label="Open resource graph"
      >
        <GitBranch size="13px" aria-hidden="true" />
        <span class="sr-only">Open resource graph</span>
      </button>
    {/if}
  </div>
</nav>

<ResourceGraphOverlay
  bind:open={graphOverlayOpen}
  anchorResource={resource}
  resources={allResources}
  isLoading={resourcesLoading}
  error={resourcesError}
/>

<style lang="postcss">
  .resource-breadcrumbs {
    @apply flex items-center h-7 flex-none w-full pr-3 gap-x-1.5 truncate line-clamp-1;
  }

  .resource-breadcrumbs__track {
    @apply inline-flex items-center min-w-0 max-w-full;
  }

  .resource-breadcrumbs__crumbs {
    @apply flex items-center gap-x-1.5 flex-1 min-w-0 overflow-hidden truncate line-clamp-1;
  }

  .graph-trigger {
    @apply flex-none inline-flex items-center justify-center rounded-md border transition-colors shadow-sm ml-1 px-2 py-[3px];
    border-color: var(--border, #e5e7eb);
    background-color: var(--surface, #ffffff);
    color: var(--muted-foreground, #6b7280);
    min-width: 30px;
    height: 26px;
  }

  .graph-trigger:hover {
    color: var(--foreground, #1f2937);
    border-color: color-mix(
      in srgb,
      var(--border, #e5e7eb) 70%,
      var(--foreground, #1f2937)
    );
  }

  .graph-trigger:focus-visible {
    @apply outline-none ring ring-offset-1;
    ring-color: var(--ring, #93c5fd);
    ring-offset-color: var(--surface, #ffffff);
  }
</style>
