<script lang="ts">
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import ResourceGraph from "../embedding/ResourceGraph.svelte";
  export let seeds: string[] | undefined;

  const instanceId = httpClient.getInstanceId();

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
/>
