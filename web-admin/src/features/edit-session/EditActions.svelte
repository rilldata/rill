<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { GitPullRequestCreateArrow, LogOut } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";
  import { isEditPreviewRoute } from "./edit-route-utils";

  export let organization: string;
  export let project: string;
  export let branch: string;

  $: editPrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  $: projectHomeHref = `/${organization}/${project}`;
  $: inPreview = isEditPreviewRoute($page.url.pathname);
  $: navHref = inPreview ? editPrefix : `${editPrefix}/dashboards`;
  $: navLabel = inPreview ? "Edit" : "Preview";
  $: navTooltip = inPreview
    ? "Switch to developer mode"
    : "Switch to preview mode";
</script>

<Tooltip distance={8}>
  <Button type="secondary" href={navHref} class="!bg-surface-base"
    >{navLabel}</Button
  >
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">{navTooltip}</span>
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

<Tooltip distance={8}>
  <a
    href={projectHomeHref}
    class="flex items-center gap-x-2 px-2 py-1 rounded text-fg-primary hover:bg-surface-hover"
  >
    <LogOut size="16" />
    <span class="text-sm font-medium">Exit</span>
  </a>
  <TooltipContent slot="tooltip-content" maxWidth="200px">
    <span class="text-xs">Return to project home</span>
  </TooltipContent>
</Tooltip>
