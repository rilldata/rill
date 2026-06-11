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

  // While GitStatus is errored, re-poll on this interval. The runtime force-refreshes the git
  // credentials on auth failures and self-heals, so re-polling lets the toolbar recover without
  // a full page reload (e.g. after the managed git token expires and is rotated).
  const GIT_STATUS_ERROR_REFETCH_INTERVAL_MS = 5000;

  const client = useRuntimeClient();
  const gitStatusQuery = createRuntimeServiceGitStatus(
    client,
    {},
    {
      query: {
        refetchInterval: (query) =>
          query.state.status === "error"
            ? GIT_STATUS_ERROR_REFETCH_INTERVAL_MS
            : false,
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
