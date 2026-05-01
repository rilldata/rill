<script context="module" lang="ts">
  // Editing `/-/edit/files/dashboards/<name>.yaml` previews the corresponding
  // explore/canvas dashboard directly; from any other editor page, preview
  // lands on the branch's cloud preview root.
  const DASHBOARD_FILE_RE =
    /\/-\/edit\/files\/dashboards\/(.+)\.yaml(?:$|[/?#])/;

  function getPreviewSubpath(
    pathname: string,
    dashboards: V1Resource[],
  ): string {
    const match = pathname.match(DASHBOARD_FILE_RE);
    if (!match) return "";
    const name = match[1];
    const resource = dashboards.find((r) => r.meta?.name?.name === name);
    if (resource?.explore) return `/explore/${name}`;
    if (resource?.canvas) return `/canvas/${name}`;
    return "";
  }
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { useDashboards } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
  const runtimeClient = useRuntimeClient();

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: branchHref = `/${organization}/${project}${branchPathPrefix(activeBranch)}`;
  $: dashboardsQuery = useDashboards(runtimeClient);
  $: dashboards = $dashboardsQuery.data ?? [];
  $: previewSubpath = getPreviewSubpath($page.url.pathname, dashboards);
  $: cloudPreviewHref = `${branchHref}${previewSubpath}`;
</script>

<Header borderBottom>
  <HeaderLogo href={`/${organization}/${project}`} />
  {#if activeBranch}
    <span
      class="inline-flex items-center h-7 px-2.5 rounded-2xl border border-border bg-surface-base text-fg-primary text-sm font-medium shadow-sm"
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
