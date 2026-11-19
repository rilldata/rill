<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import WorkspaceDispatcher from "@rilldata/web-common/features/workspaces/WorkspaceDispatcher.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();

  const client = useRuntimeClient();

  let { fileArtifact } = $derived(data);
  let { remoteContent } = $derived(fileArtifact);

  // The artifact whose latest content fetch failed.
  let failedArtifact: FileArtifact | undefined = $state.raw();
  // Counts fetches so that only the latest one reports a failure;
  // a slower fetch from before a file switch or a retry is ignored.
  let latestFetch = 0;

  function fetchContent(artifact: FileArtifact) {
    const fetchId = ++latestFetch;
    void artifact.fetchContent().then((content) => {
      if (fetchId === latestFetch && content === undefined) {
        failedArtifact = artifact;
      }
    });
  }

  function retry() {
    failedArtifact = undefined;
    fetchContent(fileArtifact);
  }

  // Fetch file content reactively once the runtime is available.
  // Unlike web-local, the runtime credentials aren't ready during +page.ts load.
  // Data files like .parquet have no editable text content and are rendered as
  // a data preview instead, so skip fetching their (binary) content.
  $effect(() => {
    if (
      client.host &&
      client.instanceId &&
      !fileArtifact.isPreviewableDataFile
    ) {
      fetchContent(fileArtifact);
    }
  });

  // Unlike web-local, the page renders before the content arrives,
  // and the runtime can take seconds to answer (e.g. while it restarts after a `rill.yaml` edit).
  // Mounting the workspace without the content shows an empty editor that reads as an empty file,
  // and anything typed into it would be auto-saved over the real content.
  // An empty file still counts as loaded: its content is "", not null.
  // Managed files (e.g. `.env` on cloud) show a notice instead of their content.
  let contentReady = $derived(
    fileArtifact.isPreviewableDataFile ||
      fileArtifact.managed ||
      $remoteContent !== null,
  );
</script>

{#if contentReady}
  <WorkspaceDispatcher {fileArtifact} />
{:else if failedArtifact === fileArtifact}
  <div
    class="size-full flex flex-col gap-y-3 items-center justify-center"
    role="alert"
  >
    <span class="text-sm text-fg-secondary">
      {m.edit_failed_to_load_file()}
    </span>
    <Button type="secondary" onClick={retry}>{m.common_retry()}</Button>
  </div>
{:else}
  <div class="size-full flex items-center justify-center" aria-busy="true">
    <DelayedSpinner isLoading size="24px" />
  </div>
{/if}
