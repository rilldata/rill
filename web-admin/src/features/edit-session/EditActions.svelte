<script lang="ts">
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createRuntimeServiceGitStatus } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CommitPopover from "./CommitPopover.svelte";
  import MergePopover from "./MergePopover.svelte";
  import PublishPopover from "./PublishPopover.svelte";

  export let organization: string;
  export let project: string;
  export let branch: string;
  export let primaryBranch: string | undefined = undefined;

  const client = useRuntimeClient();
  const gitStatusQuery = createRuntimeServiceGitStatus(client, {});

  $: closeHref = `/${organization}/${project}${branchPathPrefix(branch)}`;
  $: managedGit = $gitStatusQuery.data?.managedGit;
  $: gitStatusLoaded = $gitStatusQuery.data !== undefined;

  function handleClose(e: MouseEvent) {
    // Full page navigation avoids a race where useRuntimeClient() is called
    // before the project layout's RuntimeProvider remounts.
    e.preventDefault();
    window.location.href = closeHref;
  }
</script>

<Tooltip distance={8}>
  <Button type="secondary" href={closeHref} onClick={handleClose}>Done</Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Return to project home</span>
  </TooltipContent>
</Tooltip>

{#if gitStatusLoaded}
  {#if managedGit}
    <PublishPopover {organization} {project} {primaryBranch} />
  {:else}
    <CommitPopover />
    <MergePopover {organization} {project} {primaryBranch} />
  {/if}
{/if}
