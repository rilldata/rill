<script lang="ts">
  import {
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { ResourceKind } from "../entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import WorkspaceCrumb from "./WorkspaceCrumb.svelte";

  export let resource: V1Resource | undefined;
  export let filePath: string;

  $: ({ instanceId } = $runtime);

  $: resourceKind = resource?.meta?.name?.kind as ResourceKind | undefined;
  $: resourceName = resource?.meta?.name?.name;

  $: resourcesQuery = createRuntimeServiceListResources(instanceId);
  $: allResources = $resourcesQuery.data?.resources ?? [];

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

<nav
  class="flex gap-x-1.5 items-center h-7 flex-none w-full pr-3 truncate line-clamp-1 -pl-1"
>
  <WorkspaceCrumb
    selectedResource={resource}
    resources={lateralResources}
    {allResources}
    {filePath}
    current
  />
</nav>
