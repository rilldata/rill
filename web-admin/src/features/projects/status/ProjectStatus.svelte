<script lang="ts">
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectStatusHeader from "./project-information/ProjectStatusHeader.svelte";
  import ProjectStatusGitHub from "./project-information/ProjectStatusGitHub.svelte";
  import ProjectStatusOlap from "./project-information/ProjectStatusOlap.svelte";
  import ProjectStatusAI from "./project-information/ProjectStatusAI.svelte";
  import ProjectStatusLocalDev from "./project-information/ProjectStatusLocalDev.svelte";

  export let organization: string;
  export let project: string;

  $: ({ instanceId } = $runtime);

  // Instance data for OLAP and AI connectors
  $: instanceQuery = createRuntimeServiceGetInstance(instanceId, {
    sensitive: true,
  });

  $: isLoading = $instanceQuery.isLoading;
  $: isError = $instanceQuery.isError;
  $: instance = $instanceQuery.data?.instance;
  $: olapConnectorName = instance?.olapConnector;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === olapConnectorName,
  );
  $: aiConnectorName = $instanceQuery.data?.instance?.aiConnector;
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === aiConnectorName,
  );
</script>

<!-- Header row with status and version (outside the box) -->
<ProjectStatusHeader {organization} {project} />

<!-- Info grid (inside the box) -->
<div class="info-box">
  <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
    <ProjectStatusGitHub {organization} {project} />
    <ProjectStatusOlap {olapConnector} {isLoading} {isError} />
    <ProjectStatusAI {aiConnector} {isLoading} {isError} />
    <ProjectStatusLocalDev {organization} {project} />
  </div>
</div>

<style lang="postcss">
  .info-box {
    @apply p-4 bg-white border border-gray-200 rounded-lg;
  }
</style>
