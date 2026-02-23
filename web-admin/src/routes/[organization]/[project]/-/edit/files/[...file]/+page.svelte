<script lang="ts">
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ fileArtifact } = data);

  // Fetch file content reactively once the runtime is available.
  // Unlike web-local, the runtime credentials aren't ready during +page.ts load.
  $: if ($runtime.host && $runtime.instanceId && fileArtifact) {
    fileArtifact.fetchContent();
  }
</script>

<WorkspaceDispatcher {fileArtifact} />
