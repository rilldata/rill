<script lang="ts">
  import {
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import WorkspaceCrumb from "./WorkspaceCrumb.svelte";
  import ResourceGraphOverlay from "@rilldata/web-common/features/resource-graph/embedding/ResourceGraphOverlay.svelte";
  import { ALLOWED_FOR_GRAPH } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";

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
    resourceKind && ALLOWED_FOR_GRAPH.has(resourceKind) ? true : false;

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
        {graphSupported}
        openGraph={() => (graphOverlayOpen = true)}
      />
    </div>
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
</style>
