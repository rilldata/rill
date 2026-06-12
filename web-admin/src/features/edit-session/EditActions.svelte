<script lang="ts">
  import { page } from "$app/stores";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import ExploreEditDropdown from "@rilldata/web-common/features/explores/ExploreEditDropdown.svelte";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { createRuntimeServiceGitStatus } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { GitBranch } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";
  import CloudRemoteChangeManager from "./CloudRemoteChangeManager.svelte";
  import ExitButton from "./ExitButton.svelte";
  import MergePopover from "./MergePopover.svelte";
  import PublishPopover from "./PublishPopover.svelte";
  import CanvasEditButton from "@rilldata/web-common/features/canvas/CanvasEditButton.svelte";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined = undefined;

  // While GitStatus is errored, re-poll so the toolbar recovers without a full page reload once
  // the runtime refreshes the git credentials. The endpoint can keep erroring (e.g. the remote is
  // genuinely unreachable), so back off exponentially from a base delay up to a cap rather than
  // hammering it on a fixed interval.
  const GIT_STATUS_RETRY_BASE_MS = 2000;
  const GIT_STATUS_RETRY_MAX_MS = 60000;

  const client = useRuntimeClient();
  const gitStatusQuery = createRuntimeServiceGitStatus(
    client,
    {},
    {
      query: {
        refetchInterval: (query) => {
          if (query.state.status !== "error") return false;
          // fetchFailureCount grows with each consecutive failure and resets on success.
          const backoff =
            GIT_STATUS_RETRY_BASE_MS * 2 ** (query.state.fetchFailureCount - 1);
          return Math.min(backoff, GIT_STATUS_RETRY_MAX_MS);
        },
      },
    },
  );

  $: managedGit = $gitStatusQuery.data?.managedGit;
  $: gitStatusLoaded = $gitStatusQuery.data !== undefined;
  // Show the parent-level error UI only when GitStatus has never loaded.
  // After a successful load, TanStack keeps `data` populated through transient
  // refetch errors, so the popovers stay mounted and the user keeps the toolbar.
  $: gitStatusErrorMessage =
    !gitStatusLoaded && $gitStatusQuery.isError
      ? extractErrorMessage($gitStatusQuery.error)
      : "";

  $: onExplorePreview = !!$page.route.id?.startsWith(
    "/[organization]/[project]/-/edit/(viz)/explore",
  );
  $: onCanvasPreview = !!$page.route.id?.startsWith(
    "/[organization]/[project]/-/edit/(viz)/canvas",
  );
  $: dashboardName = $page.params.name ?? "";
</script>

{#if onExplorePreview && dashboardName}
  <ExploreEditDropdown exploreName={dashboardName} />
{/if}
{#if onCanvasPreview && dashboardName}
  <CanvasEditButton canvasName={dashboardName} />
{/if}

{#if gitStatusLoaded}
  {#if managedGit}
    <CloudRemoteChangeManager {primaryBranch} />
    <PublishPopover {organization} {project} {primaryBranch} />
  {:else}
    <CommitPopover />
    <MergePopover {organization} {project} {primaryBranch} />
  {/if}
{:else if gitStatusErrorMessage}
  <Tooltip distance={8}>
    <Button
      type="primary"
      loading={$gitStatusQuery.isFetching}
      onClick={() => $gitStatusQuery.refetch()}
    >
      <GitBranch size="14" />
      Git unavailable
    </Button>
    <TooltipContent slot="tooltip-content" maxWidth="220px">
      <span class="text-xs">{gitStatusErrorMessage} Click to retry.</span>
    </TooltipContent>
  </Tooltip>
{/if}

<ExitButton {organization} {project} />
