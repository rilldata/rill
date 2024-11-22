<script lang="ts">
  import {
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import WorkspaceCrumb from "./WorkspaceCrumb.svelte";

  const upstreamMapping = new Map([
    [ResourceKind.MetricsView, ResourceKind.Explore],
    [ResourceKind.Source, ResourceKind.Model],
    [ResourceKind.Model, ResourceKind.MetricsView],
  ]);

  export let resource: V1Resource | undefined;
  export let filePath: string;

  $: ({ instanceId } = $runtime);

  $: resourceKind = resource?.meta?.name?.kind as ResourceKind | undefined;
  $: resourceName = resource?.meta?.name?.name;

  $: resourcesQuery = createRuntimeServiceListResources(instanceId);
  $: allResources = $resourcesQuery.data?.resources ?? [];

  $: upstreamKind = resourceKind && upstreamMapping.get(resourceKind);

  $: downstreamResources = upstreamKind
    ? allResources.filter(({ meta }) => {
        return (
          meta?.name?.kind === upstreamKind &&
          meta?.refs?.find(({ kind, name }) => {
            return kind === resourceKind && name === resourceName;
          })
        );
      })
    : [];
</script>

<nav class="flex gap-x-1.5 items-center h-7 mt-2 flex-none w-full pr-3">
  <WorkspaceCrumb selected resources={[resource]} {allResources} {filePath} />

  {#if downstreamResources.length}
    <WorkspaceCrumb
      downstream
      resources={downstreamResources}
      {allResources}
      {filePath}
    />
  {/if}
</nav>
