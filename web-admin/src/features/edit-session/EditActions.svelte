<script lang="ts">
  import { goto } from "$app/navigation";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { LogOut } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";
  import MergePopover from "./MergePopover.svelte";

  export let organization: string;
  export let project: string;
  export let branch: string;
  export let primaryBranch: string | undefined = undefined;

  $: closeHref = `/${organization}/${project}${branchPathPrefix(branch)}`;

  function handleClose(e: MouseEvent) {
    e.preventDefault();
    void goto(closeHref);
  }
</script>

<CommitPopover />
<MergePopover {organization} {project} {primaryBranch} />

<Tooltip distance={8}>
  <Button type="secondary" href={closeHref} onClick={handleClose}>
    <LogOut size="14" />
    Exit
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Return to project home</span>
  </TooltipContent>
</Tooltip>
