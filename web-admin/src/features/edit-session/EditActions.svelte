<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { GitPullRequestCreateArrow, LogOut } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";

  export let organization: string;
  export let project: string;

  $: projectHomeHref = `/${organization}/${project}`;
</script>

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

<Tooltip distance={8}>
  <Button type="ghost" href={projectHomeHref} compact>
    <LogOut size="14" />
    Exit
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Return to project home</span>
  </TooltipContent>
</Tooltip>
