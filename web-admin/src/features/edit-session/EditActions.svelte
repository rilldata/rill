<script lang="ts">
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { GitPullRequestCreateArrow } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";

  export let organization: string;
  export let project: string;
  export let branch: string;

  $: closeHref = `/${organization}/${project}${branchPathPrefix(branch)}`;

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

<CommitPopover />

<Tooltip distance={8}>
  <Button type="primary" disabled>
    <GitPullRequestCreateArrow size="14" />
    Open PR
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Coming soon</span>
  </TooltipContent>
</Tooltip>
