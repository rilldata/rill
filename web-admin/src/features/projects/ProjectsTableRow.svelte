<script lang="ts">
  import {
    createAdminServiceGetProject,
    type V1ProjectPermissions,
  } from "@rilldata/web-admin/client";
  import ProjectStatusBadge from "./ProjectStatusBadge.svelte";
  import ProjectsTableActionsCell from "./ProjectsTableActionsCell.svelte";

  export let organization: string;
  export let project: string;
  export let isPublic: boolean;

  $: proj = createAdminServiceGetProject(organization, project);

  function getRoleLabel(perms: V1ProjectPermissions | undefined): string {
    if (!perms) return "";
    if (perms.admin) return "Admin";
    if (perms.manageProject) return "Editor";
    if (perms.readProject) return "Viewer";
    return "";
  }

  $: deploymentStatus = $proj.data?.deployment?.status;
  $: hasDeployment = !!$proj.data?.deployment;
  $: permissions = $proj.data?.projectPermissions;
  $: roleLabel = getRoleLabel(permissions);
</script>

<div class="row">
  <a
    href={`/${organization}/${project}`}
    class="cell cell-name text-fg-primary text-sm font-medium truncate hover:text-accent-primary-action"
  >
    {project}
  </a>
  <div class="cell">
    <ProjectStatusBadge {deploymentStatus} {isPublic} {hasDeployment} />
  </div>
  <div class="cell text-fg-primary text-sm">
    {isPublic ? "Public" : "Private"}
  </div>
  <div class="cell text-fg-primary text-sm">
    {roleLabel}
  </div>
  <div class="cell cell-actions">
    <ProjectsTableActionsCell {organization} {project} {permissions} />
  </div>
</div>

<style lang="postcss">
  .row {
    @apply grid items-center w-full border-b border-border;
    @apply h-[52px];
    grid-template-columns: minmax(0, 1fr) 200px 200px 200px 60px;
  }

  .cell {
    @apply px-2 min-w-0;
  }

  .cell-name {
    @apply block;
  }

  .cell-actions {
    @apply flex items-center justify-end;
  }
</style>
