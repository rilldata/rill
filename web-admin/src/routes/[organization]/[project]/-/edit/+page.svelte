<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { navigateToFile } from "@rilldata/web-common/features/workspaces/workspace-routing";

  $: ({ instanceId } = $runtime);

  // List project files to auto-navigate to the first one
  $: filesQuery = createRuntimeServiceListFiles(instanceId, undefined, {
    query: {
      select: (data) =>
        data?.files
          ?.filter((f) => !f.isDir && !f.path?.startsWith("/tmp"))
          ?.sort((a, b) => (a.path ?? "").localeCompare(b.path ?? "")) ?? [],
    },
  });

  // Navigate to the first file when available
  $: if ($filesQuery.data?.length) {
    const firstFile = $filesQuery.data[0].path;
    if (firstFile) {
      void navigateToFile(firstFile);
    }
  }
</script>

<div class="flex items-center justify-center h-full text-fg-muted">
  {#if $filesQuery.isLoading}
    Loading project files...
  {:else if !$filesQuery.data?.length}
    No files in this project
  {/if}
</div>
