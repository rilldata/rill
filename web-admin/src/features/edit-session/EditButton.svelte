<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceCreateDeployment,
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import {
    deriveDefaultBranchName,
    injectBranchIntoPath,
    requestSkipBranchInjection,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { AxiosError } from "axios";
  import EditBranchDialog from "./EditBranchDialog.svelte";
  import {
    invalidateDeployments,
    useOwnDevDeployments,
  } from "./use-edit-session";

  export let organization: string;
  export let project: string;
  /** The branch currently being viewed (from the URL), if any. */
  export let activeBranch: string | undefined = undefined;
  /** The project's primary branch, used as the source for new branches. */
  export let primaryBranch: string | undefined = undefined;

  const userQuery = createAdminServiceGetCurrentUser();
  const projectQuery = createAdminServiceGetProject(organization, project);
  const ownQuery = useOwnDevDeployments(organization, project);
  const createMutation = createAdminServiceCreateDeployment();

  let dialogOpen = false;
  let dialogInitialBranchName: string | undefined = undefined;

  $: email = $userQuery.data?.user?.email;
  $: ({ ownDeployments } = $ownQuery);
  $: isManagedGit = !!$projectQuery.data?.project?.managedGitId;

  $: isLoading =
    $userQuery.isLoading || $ownQuery.isLoading || $projectQuery.isLoading;
  $: isStarting = $createMutation.isPending;

  // Direct-edit shortcut: the user is viewing a branch they own and it's
  // editable. Skip all the branching logic and link straight to it. Preserved
  // as an `href` so middle-click / cmd-click open in a new tab.
  $: activeBranchDeployment = activeBranch
    ? ownDeployments.find((d) => d.branch === activeBranch)
    : undefined;

  $: directEditHref = activeBranchDeployment?.branch
    ? injectBranchIntoPath(
        `/${organization}/${project}/-/edit`,
        activeBranchDeployment.branch,
      )
    : undefined;

  function handleDirectEdit(e: MouseEvent) {
    e.preventDefault();
    requestSkipBranchInjection();
    void goto(directEditHref!);
  }

  function gotoBranch(branch: string) {
    requestSkipBranchInjection();
    void goto(
      injectBranchIntoPath(`/${organization}/${project}/-/edit`, branch),
    );
  }

  function openDialog(initialName?: string) {
    dialogInitialBranchName = initialName;
    dialogOpen = true;
  }

  async function autoCreateAndEdit(branchName: string) {
    try {
      const resp = await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
          branch: branchName,
        },
      });
      // Await invalidation so the destination edit page reads a fresh
      // deployment list — the create response isn't auto-inserted into the
      // ListDeployments cache.
      await invalidateDeployments(organization, project);
      const newBranch = resp.deployment?.branch ?? branchName;
      gotoBranch(newBranch);
    } catch (err) {
      // Cross-user collision: someone else's branch already occupies this name.
      // Hand off to the dialog with the name pre-filled so the user can pick
      // a different one.
      const code = (err as AxiosError<RpcStatus>)?.response?.data?.code;
      if (code === 6 /* ALREADY_EXISTS */) {
        openDialog(branchName);
        return;
      }
      const message =
        getRpcErrorMessage(err as RpcStatus) ?? "Failed to start edit session.";
      eventBus.emit("notification", { type: "error", message });
    }
  }

  async function handleEditClick() {
    if (isLoading || isStarting) return;

    // Rill-managed users are limited to one dev branch each (UI convention).
    // If they already have one, route directly to it — never show the modal.
    if (ownDeployments.length > 0 && isManagedGit) {
      gotoBranch(ownDeployments[0].branch ?? primaryBranch ?? "main");
      return;
    }

    // Self-managed with at least one branch: existing modal flow.
    if (ownDeployments.length > 0) {
      openDialog();
      return;
    }

    // Zero owned dev branches: auto-create one named after the email's local
    // part. If the email is missing or sanitization yields nothing usable,
    // fall back to the modal so the user can name the branch themselves.
    const branchName = deriveDefaultBranchName(email);
    if (!branchName) {
      openDialog();
      return;
    }
    await autoCreateAndEdit(branchName);
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
    disabled={isLoading || isStarting}
    loading={isStarting}
    loadingCopy="Starting..."
    onClick={handleEditClick}
  >
    Edit
  </Button>
  <EditBranchDialog
    bind:open={dialogOpen}
    {organization}
    {project}
    {activeBranch}
    {primaryBranch}
    initialBranchName={dialogInitialBranchName}
  />
{/if}
