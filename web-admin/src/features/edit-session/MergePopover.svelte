<script lang="ts">
  import {
    createAdminServiceDeleteDeployment,
    createAdminServiceListDeployments,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitStatus,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ExternalLink, GitPullRequest } from "lucide-svelte";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined;

  let open = false;
  let isMerging = false;
  let errorMessage = "";

  const client = useRuntimeClient();
  const gitMergeMutation = createRuntimeServiceGitMergeToBranchMutation(client);
  const gitStatusQuery = createRuntimeServiceGitStatus(client, {});
  const deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );
  const deleteDeploymentMutation = createAdminServiceDeleteDeployment();

  $: currentBranch = $gitStatusQuery.data?.branch ?? "";
  $: devDeploymentId = $deploymentsQuery.data?.deployments?.find(
    (d) => d.runtimeInstanceId === client.instanceId,
  )?.id;
  $: branchUrl =
    $gitStatusQuery.data?.githubUrl && currentBranch
      ? `${getGitUrlFromRemote($gitStatusQuery.data.githubUrl)}/tree/${encodeURIComponent(currentBranch)}`
      : "";
  $: alreadyOnPrimary =
    !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;
  $: disabled = !primaryBranch || alreadyOnPrimary || isMerging;

  $: if (!open) {
    errorMessage = "";
  }

  async function handleMerge() {
    if (!primaryBranch || isMerging) return;
    isMerging = true;
    errorMessage = "";
    try {
      await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: false,
      });
    } catch (err) {
      errorMessage = getRpcErrorMessage(err as RpcStatus) ?? "Failed to merge";
      isMerging = false;
      return;
    }

    // Tear down the dev deployment now that its changes live in production.
    // Failures here are non-fatal: auto-hibernation and (eventually) backend
    // GitHub-event cleanup will catch any orphans.
    if (devDeploymentId) {
      try {
        await $deleteDeploymentMutation.mutateAsync({
          deploymentId: devDeploymentId,
        });
      } catch (err) {
        console.warn("Failed to delete dev deployment after merge", err);
      }
    }

    // Full page navigation matches the Done button: avoids a race where
    // useRuntimeClient() is called before the project layout's
    // RuntimeProvider remounts on the production branch.
    window.location.href = `/${organization}/${project}`;
  }
</script>

<Tooltip distance={8} suppress={open}>
  <Popover.Root bind:open>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="primary" {disabled}>
          <GitPullRequest size="14" />
          Merge to production
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content align="end" class="!w-[320px]">
      <div class="flex flex-col gap-y-3">
        <p class="text-xs text-fg-secondary">
          Merging pushes changes from
          <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
          to production, closes the editor, and returns you to the project
          home. Viewers will see updates as the project reconciles.
        </p>
        {#if branchUrl}
          <a
            class="github-link"
            href={branchUrl}
            target="_blank"
            rel="noopener noreferrer"
          >
            View branch on GitHub
            <ExternalLink size="11" />
          </a>
        {/if}
        <Button
          type="primary"
          small
          disabled={isMerging}
          loading={isMerging}
          loadingCopy="Merging..."
          onClick={handleMerge}
        >
          Merge
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
      {:else if !primaryBranch}
        Loading project...
      {:else}
        Review and confirm before merging
      {/if}
    </span>
  </TooltipContent>
</Tooltip>

<style lang="postcss">
  .github-link {
    @apply inline-flex items-center gap-x-1 text-xs text-fg-secondary;
    @apply hover:text-fg-primary hover:underline;
  }
</style>
