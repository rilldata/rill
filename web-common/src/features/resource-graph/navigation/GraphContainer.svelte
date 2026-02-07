<script lang="ts">
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ResourceGraph from "../embedding/ResourceGraph.svelte";
  export let seeds: string[] | undefined;
  export let searchQuery = "";
  export let statusFilter: "all" | "pending" | "errored" = "all";

  $: ({ instanceId } = $runtime);

  $: resourcesQuery = createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      retry: 2,
      refetchOnMount: true,
      refetchOnWindowFocus: false,
      enabled: !!instanceId,
    },
  });

  $: resources = $resourcesQuery.data?.resources ?? [];
  $: errorMessage = $resourcesQuery.error
    ? "Failed to load project resources."
    : null;
</script>

<ResourceGraph
  {resources}
  isLoading={$resourcesQuery.isLoading}
  error={errorMessage}
  {seeds}
  {searchQuery}
  {statusFilter}
/>
