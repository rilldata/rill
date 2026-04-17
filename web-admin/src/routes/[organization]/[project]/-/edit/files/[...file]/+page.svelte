<script lang="ts">
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  export let data: PageData;

  const client = useRuntimeClient();

  $: ({ fileArtifact } = data);

  // Fetch file content reactively once the runtime is available.
  // Unlike web-local, the runtime credentials aren't ready during +page.ts load.
  $: if (client.host && client.instanceId && fileArtifact) {
    fileArtifact.fetchContent();
  }
</script>

<WorkspaceDispatcher {fileArtifact} />
