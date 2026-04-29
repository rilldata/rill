<script lang="ts">
  import { page } from "$app/stores";
  import { branchPathPrefix } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { GitPullRequestCreateArrow } from "lucide-svelte";
  import CommitPopover from "./CommitPopover.svelte";
  import { isEditPreviewRoute } from "./edit-route-utils";

  export let organization: string;
  export let project: string;
  export let branch: string;

  $: editPrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;
  $: inPreview = isEditPreviewRoute($page.url.pathname);
  $: navHref = inPreview ? editPrefix : `${editPrefix}/dashboards`;
  $: navLabel = inPreview ? "Edit" : "Preview";
  $: navTooltip = inPreview
    ? "Switch to developer mode"
    : "Switch to preview mode";
</script>

<Tooltip distance={8}>
  <Button type="secondary" href={navHref}>{navLabel}</Button>
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
