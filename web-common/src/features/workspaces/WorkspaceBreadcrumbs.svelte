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

  $: lateralResources = allResources.filter((r) => {
    if (
      r.meta?.name?.name === resourceName &&
      r.meta?.name?.kind === resourceKind
    )
      return true;
    if (!r.meta?.refs?.length) return false;
    return r.meta?.refs?.every((reference) =>
      resource?.meta?.refs?.find(
        (ref) => ref?.name === reference.name && ref?.kind === reference.kind,
      ),
    );
  });
</script>

<nav
  class="flex gap-x-1.5 items-center h-7 flex-none w-full pr-3 truncate line-clamp-1"
>
  <WorkspaceCrumb
    selected
    resources={lateralResources}
    {allResources}
    {filePath}
  />
</nav>
