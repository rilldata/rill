<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  const client = useRuntimeClient();

  let { fileArtifact } = $derived(data);
  let { remoteContent } = $derived(fileArtifact);

  // The artifact whose most recent content fetch has settled, successfully or not.
  let settledArtifact: FileArtifact | undefined = $state.raw();

  // Fetch file content reactively once the runtime is available.
  // Unlike web-local, the runtime credentials aren't ready during +page.ts load.
  // Data files like .parquet have no editable text content and are rendered as
  // a data preview instead, so skip fetching their (binary) content.
  $effect(() => {
    if (
      client.host &&
      client.instanceId &&
      fileArtifact &&
      !fileArtifact.isPreviewableDataFile
    ) {
      const artifact = fileArtifact;
      void artifact.fetchContent().finally(() => {
        settledArtifact = artifact;
      });
    }
  });

  // Unlike web-local, the page renders before the content arrives,
  // and the runtime can take seconds to answer (e.g. while it restarts after a `rill.yaml` edit).
  // Mounting the workspace earlier shows an empty editor that reads as an empty file,
  // and anything typed into it would be auto-saved over the real content.
  // A failed fetch still mounts the workspace, as before.
  let contentReady = $derived(
    fileArtifact.isPreviewableDataFile ||
      $remoteContent !== null ||
      settledArtifact === fileArtifact,
  );
</script>

{#if contentReady}
  <WorkspaceDispatcher {fileArtifact} />
{:else}
  <div class="size-full flex items-center justify-center" aria-busy="true">
    <DelayedSpinner isLoading size="24px" />
  </div>
{/if}
