<script lang="ts">
  import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  import ResourceGraph from "../embedding/ResourceGraph.svelte";
  export let seeds: string[] | undefined;

  $: ({ instanceId } = runtimeClient);

  $: resourcesQuery = createRuntimeServiceListResources(
    runtimeClient,
    {},
    {
      query: {
        retry: 2,
        refetchOnMount: true,
        refetchOnWindowFocus: false,
        enabled: !!instanceId,
      },
    },
  );

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
