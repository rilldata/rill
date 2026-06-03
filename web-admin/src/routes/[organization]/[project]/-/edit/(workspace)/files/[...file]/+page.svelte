<script lang="ts">
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  const client = useRuntimeClient();

  let { fileArtifact } = $derived(data);

  // Fetch file content reactively once the runtime is available.
  // Unlike web-local, the runtime credentials aren't ready during +page.ts load.
  $effect(() => {
    if (client.host && client.instanceId && fileArtifact) {
      void fileArtifact.fetchContent();
    }
  });
</script>

<WorkspaceDispatcher {fileArtifact} />
