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

  $: previewHref = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit/dashboards`;
</script>

<Tooltip distance={8}>
  <Button type="secondary" href={previewHref}>Preview</Button>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Switch to preview mode</span>
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
