<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteDeployment,
    createAdminServiceListDeployments,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { optimisticallyRemoveDeployment } from "@rilldata/web-admin/features/branches/branch-actions";
  import { requestSkipBranchInjection } from "@rilldata/web-admin/features/branches/branch-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitPushMutation,
    createRuntimeServiceGitStatus,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Send } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined;

  const AUTO_COMMIT_MESSAGE = "Updates from cloud editor";

  let open = false;
  let isPublishing = false;
  let errorMessage = "";

  const client = useRuntimeClient();
  const gitPushMutation = createRuntimeServiceGitPushMutation(client);
  const gitMergeMutation = createRuntimeServiceGitMergeToBranchMutation(client);
  const gitStatusQuery = createRuntimeServiceGitStatus(client, {});
  const deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );
  const deleteDeploymentMutation = createAdminServiceDeleteDeployment();

  $: currentBranch = $gitStatusQuery.data?.branch ?? "";
  $: hasLocalChanges = $gitStatusQuery.data?.localChanges ?? false;
  $: gitStatusErrorMessage = $gitStatusQuery.isError
    ? (getRpcErrorMessage($gitStatusQuery.error as RpcStatus) ??
      "Couldn't load branch info")
    : "";
  $: devDeploymentId = $deploymentsQuery.data?.deployments?.find(
    (d) => d.runtimeInstanceId === client.instanceId,
  )?.id;
  $: alreadyOnPrimary =
    !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;
  $: disabled =
    !primaryBranch || !currentBranch || alreadyOnPrimary || isPublishing;

  $: if (!open) {
    errorMessage = "";
  }

  async function handlePublish() {
    if (!primaryBranch || isPublishing) return;
    isPublishing = true;
    errorMessage = "";

    try {
      if (hasLocalChanges) {
        await $gitPushMutation.mutateAsync({
          commitMessage: AUTO_COMMIT_MESSAGE,
          force: false,
        });
      }
      await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: false,
      });
    } catch (err) {
      errorMessage =
        getRpcErrorMessage(err as RpcStatus) ?? "Failed to publish";
      isPublishing = false;
      return;
    }

    // First, leave the edit page. Deleting the deployment while the page is still
    // mounted would 404 its deployment queries and flash an error. The skip call
    // opts out of the project layout's `beforeNavigate` branch injection so we
    // actually land on the production project home, not back on the dev branch.
    requestSkipBranchInjection();
    await goto(`/${organization}/${project}`);

    // Second, delete the dev deployment. On success, drop it from the
    // ListDeployments cache so the BranchSelector on the destination page
    // doesn't show the now-deleted branch.
    // Note that the browser may cancel this request on page tear-down, so a better approach may be to
    // hand off the deployment id via sessionStorage and fire the delete from the destination.
    if (devDeploymentId) {
      const id = devDeploymentId;
      $deleteDeploymentMutation.mutate({ deploymentId: id });
      void optimisticallyRemoveDeployment(organization, project, id);
    }
  }
</script>

<Tooltip distance={8} suppress={open}>
  <Popover.Root bind:open>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="primary" {disabled}>
          <Send size="14" />
          Publish
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content align="end" class="!w-[320px]">
      <div class="flex flex-col gap-y-3">
        {#if gitStatusErrorMessage}
          <p class="text-xs text-red-600">{gitStatusErrorMessage}</p>
        {:else}
          <p class="text-xs text-fg-secondary">
            Publish your changes to production and return to the project home.
            Viewers will see updates as the project reconciles.
          </p>
        {/if}
        <Button
          type="primary"
          small
          disabled={isPublishing}
          loading={isPublishing}
          loadingCopy="Publishing..."
          onClick={handlePublish}
        >
          Publish
        </Button>
        {#if errorMessage}
          <p class="text-xs text-red-600">{errorMessage}</p>
        {/if}
      </div>
    </Popover.Content>
  </Popover.Root>
  <TooltipContent slot="tooltip-content" maxWidth="220px">
    <span class="text-xs">
      {#if alreadyOnPrimary}
        Already on production
      {:else if gitStatusErrorMessage}
        {gitStatusErrorMessage}
      {:else if !primaryBranch || !currentBranch}
        Loading project...
      {:else}
        Publish your latest changes
      {/if}
    </span>
  </TooltipContent>
</Tooltip>
