<script lang="ts">
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ fileArtifact } = data);
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <WorkspaceDispatcher {fileArtifact} />
  </div>
  <DeveloperChat />
</div>
