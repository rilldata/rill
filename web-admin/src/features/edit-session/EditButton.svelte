<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    injectBranchIntoPath,
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

  const devDeployments = useDevDeployments(organization, project);

  let dialogOpen = false;

  $: isLoading = $devDeployments.isLoading;

  // On a branch view, jump straight into edit mode for that branch.
  // On production view, fall through to the dialog so the user can pick
  // an existing dev branch or create a new one.
  $: directEditHref = activeBranch
    ? injectBranchIntoPath(`/${organization}/${project}/-/edit`, activeBranch)
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
