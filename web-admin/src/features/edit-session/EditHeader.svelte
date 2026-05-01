<script lang="ts">
  import { page } from "$app/stores";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import { GitBranchIcon, PlayIcon } from "lucide-svelte";
  import {
    createAdminServiceGetCurrentUser,
    type V1ProjectPermissions,
  } from "../../client";
  import AvatarButton from "../authentication/AvatarButton.svelte";

  export let organization: string;
  export let project: string;
  export let projectPermissions: V1ProjectPermissions;

  const user = createAdminServiceGetCurrentUser();

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  // Cloud preview = the production-style branch view at `/{org}/{project}/@{branch}`,
  // without the `/-/edit` chrome.
  $: cloudPreviewHref = `/${organization}/${project}${branchPathPrefix(activeBranch)}/dashboards`;
</script>

<Header borderBottom tinted>
  {#if activeBranch}
    <span
      class="inline-flex items-center h-7 px-2.5 rounded-2xl border border-border bg-surface-base text-fg-primary text-sm font-medium shadow-xs"
    >
      Developer
    </span>
  {/if}
  <nav class="flex gap-x-2 items-center">
    <ol class="flex flex-row items-center">
      <li class="flex items-center gap-x-2 px-2">
        <span class="text-fg-muted">{project}</span>
      </li>
      {#if activeBranch}
        <Slash />
        <li class="flex items-center gap-x-2 px-2">
          <span
            class="text-fg-primary font-medium flex flex-row items-center gap-x-2"
          >
            <GitBranchIcon size="14" class="text-fg-primary" />
            {activeBranch.length > 12
              ? activeBranch.slice(0, 11) + "…"
              : activeBranch}
          </span>
        </li>
      {/if}
    </ol>
  </nav>

  <div class="flex gap-x-2 items-center ml-auto">
    {#if activeBranch}
      <Tooltip distance={8}>
        <Button type="secondary" href={cloudPreviewHref}>
          <PlayIcon size="14" />
          Preview
        </Button>
        <TooltipContent slot="tooltip-content" maxWidth="200px">
          <span class="text-xs">Preview this branch as a viewer</span>
        </TooltipContent>
      </Tooltip>
    {/if}
    <EditActions {organization} {project} />
    {#if $user.isSuccess && $user.data?.user}
      <AvatarButton {projectPermissions} />
    {/if}
  </div>
</Header>
