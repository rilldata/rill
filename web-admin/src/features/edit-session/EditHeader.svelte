<script lang="ts">
  import { page } from "$app/stores";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import EditActions from "@rilldata/web-admin/features/edit-session/EditActions.svelte";
  import { isEditPreviewRoute } from "@rilldata/web-admin/features/edit-session/edit-route-utils";
  import { useBreadcrumbProjectPaths } from "@rilldata/web-admin/features/navigation/breadcrumb-selectors";
  import Breadcrumbs from "@rilldata/web-common/components/navigation/breadcrumbs/Breadcrumbs.svelte";
  import Slash from "@rilldata/web-common/components/navigation/breadcrumbs/Slash.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import Header from "@rilldata/web-common/layout/header/Header.svelte";
  import HeaderLogo from "@rilldata/web-common/layout/header/HeaderLogo.svelte";
  import { GitBranchIcon } from "lucide-svelte";
  import {
    createAdminServiceGetCurrentUser,
    type V1ProjectPermissions,
  } from "../../client";
  import AvatarButton from "../authentication/AvatarButton.svelte";
  import ViewAsUserChip from "../view-as-user/ViewAsUserChip.svelte";
  import { viewAsUserStore } from "../view-as-user/viewAsUserStore";

  export let organization: string;
  export let project: string;
  export let projectPermissions: V1ProjectPermissions;
  export let readProjects: boolean = false;

  const user = createAdminServiceGetCurrentUser();
  const { developerChat } = featureFlags;

  $: activeBranch = extractBranchFromPath($page.url.pathname);
  $: previewMode = isEditPreviewRoute($page.url.pathname);

  $: projectPathsQuery = useBreadcrumbProjectPaths(organization, readProjects);

  // Edit header collapses the breadcrumb to "Project / Branch": no org row,
  // no trial pill, no dashboard segment.
  $: pathParts = [null, { options: $projectPathsQuery.data ?? new Map() }];
  $: currentPath = [undefined, project];
</script>

<Header borderBottom tinted>
  <HeaderLogo href={`/${organization}/${project}`} />
  {#if activeBranch}
    <Tag
      text={previewMode ? "Preview" : "Developer"}
      color="gray"
      class="!bg-surface-base"
    />
  {/if}
  <Breadcrumbs {pathParts} {currentPath}>
    <svelte:fragment slot="after-project">
      {#if activeBranch}
        <Slash />
        <li class="flex items-center gap-x-1.5 px-2">
          <GitBranchIcon size="14" class="text-fg-muted" />
          <span class="text-fg-muted">
            {activeBranch.length > 12
              ? activeBranch.slice(0, 11) + "…"
              : activeBranch}
          </span>
        </li>
      {/if}
    </svelte:fragment>
  </Breadcrumbs>

  <div class="flex gap-x-2 items-center ml-auto">
    {#if previewMode && $viewAsUserStore}
      <ViewAsUserChip />
    {/if}
    {#if $developerChat && !previewMode}
      <ChatToggle class="!bg-surface-base" />
    {/if}
    <EditActions {organization} {project} branch={activeBranch ?? ""} />
    {#if $user.isSuccess && $user.data?.user}
      <AvatarButton {projectPermissions} />
    {/if}
  </div>
</Header>
