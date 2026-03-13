<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";
  import { navigateToFile } from "@rilldata/web-common/features/workspaces/workspace-routing";

  const client = useRuntimeClient();

  // List project files to auto-navigate to the first one
  $: filesQuery = createRuntimeServiceListFiles(
    client,
    {},
    {
      query: {
        select: (data) =>
          data?.files
            ?.filter((f) => !f.isDir && !f.path?.startsWith("/tmp"))
            ?.sort((a, b) => (a.path ?? "").localeCompare(b.path ?? "")) ?? [],
      },
    },
  );

  // Navigate to the first file once. Prefer rill.yaml as the landing file;
  // fall back to the first alphabetical file.
  let navigated = false;
  $: if (!navigated && $filesQuery.data?.length) {
    const files = $filesQuery.data;
    const target = files.find((f) => f.path === "/rill.yaml") ?? files[0];
    if (target?.path) {
      navigated = true;
      void navigateToFile(target.path, { replaceState: true });
    }
  }
</script>

<div class="flex items-center justify-center h-full text-fg-muted text-sm">
  {#if $filesQuery.isLoading}
    Loading project files...
  {:else if !$filesQuery.data?.length}
    No files in this project
  {:else}
    Select a file from the sidebar to start editing
  {/if}
</div>
