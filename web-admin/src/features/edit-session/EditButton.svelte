<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceGetCurrentUser,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    branchPathPrefix,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { useDevDeployments } from "./use-edit-session";
  import EditBranchDialog from "./EditBranchDialog.svelte";

  export let organization: string;
  export let project: string;
  /** The branch currently being viewed (from the URL), if any. */
  export let activeBranch: string | undefined = undefined;
  /** The project's primary branch, used as the source for new branches. */
  export let primaryBranch: string | undefined = undefined;

  const user = createAdminServiceGetCurrentUser();
  const devDeployments = useDevDeployments(organization, project);

  let dialogOpen = false;

  $: currentUserId = $user.data?.user?.id;
  $: deployments = $devDeployments.data?.deployments ?? [];
  $: isLoading = $devDeployments.isLoading;

  // If viewing a branch the user owns, clicking the button should go straight
  // there — no dialog.
  $: activeBranchDeployment =
    activeBranch && currentUserId
      ? deployments.find(
          (d) =>
            d.branch === activeBranch &&
            d.ownerUserId === currentUserId &&
            d.editable &&
            d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING &&
            d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
        )
      : undefined;

  $: directEditHref = activeBranchDeployment
    ? `/${organization}/${project}${branchPathPrefix(activeBranchDeployment.branch)}/-/edit`
    : undefined;

  function handleDirectEdit(e: MouseEvent) {
    e.preventDefault();
    requestSkipBranchInjection();
    void goto(directEditHref!);
  }
</script>

{#if directEditHref}
  <Button
    type="secondary"
    href={directEditHref}
    disabled={isLoading}
    onClick={handleDirectEdit}
  >
    Edit
  </Button>
{:else}
  <Button
    type="secondary"
    disabled={isLoading}
    onClick={() => (dialogOpen = true)}
  >
    Edit
  </Button>
  <EditBranchDialog
    bind:open={dialogOpen}
    {organization}
    {project}
    {activeBranch}
    {primaryBranch}
  />
{/if}
