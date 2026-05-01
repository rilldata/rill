<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListDeployments } from "@rilldata/web-admin/client";
  import { isProdDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitStatus,
    getRuntimeServiceGitStatusQueryKey,
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
  // Self-managed git projects can lack a prod deployment when created via
  // `rill project connect-github --skip-deploy` (see `cli/cmd/project/
  // connect_github.go`), so we mirror PublishPopover's first-vs-subsequent
  // logic and route to /-/invite vs /-/deploying accordingly.
  const deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );

  $: currentBranch = $gitStatusQuery.data?.branch ?? "";
  $: branchUrl =
    $gitStatusQuery.data?.githubUrl && currentBranch
      ? `${getGitUrlFromRemote($gitStatusQuery.data.githubUrl)}/tree/${encodeURIComponent(currentBranch)}`
      : "";
  $: deploymentsLoaded = $deploymentsQuery.data !== undefined;
  $: hasProdDeployment =
    $deploymentsQuery.data?.deployments?.some(isProdDeployment) ?? false;
  // Pull the dashboard name from `/-/edit/explore/<name>` or
  // `/-/edit/canvas/<name>` so the deploying page can route the user back to
  // it once the runtime is ready.
  $: currentDashboard = $page.url.pathname.match(
    /\/-\/edit\/(?:explore|canvas)\/([^/?#]+)/,
  )?.[1];
  $: alreadyOnPrimary =
    !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;
  $: disabled =
    !primaryBranch ||
    !currentBranch ||
    !deploymentsLoaded ||
    alreadyOnPrimary ||
    isMerging;

  $: if (!open) {
    errorMessage = "";
  }

  async function handleMerge() {
    if (!primaryBranch || isMerging) return;
    isMerging = true;
    errorMessage = "";

    // Open the destination tab synchronously so the browser ties it to the
    // user's click gesture; otherwise pop-up blockers reject the open after
    // the awaited merge below. Navigated on success; closed on failure.
    const targetWindow = window.open("about:blank", "_blank");
    if (!targetWindow) {
      errorMessage = "Pop-up was blocked. Please allow pop-ups and try again.";
      isMerging = false;
      return;
    }

    const isFirstMerge = !hasProdDeployment;

    try {
      await $gitMergeMutation.mutateAsync({
        branch: primaryBranch,
        force: false,
      });
      // gitStatus tracks currentBranch; refresh so subscribers see the merge.
      await queryClient.invalidateQueries({
        queryKey: getRuntimeServiceGitStatusQueryKey(client.instanceId, {}),
      });
    } catch (err) {
      targetWindow.close();
      errorMessage = extractErrorMessage(err) || "Failed to merge";
      isMerging = false;
      return;
    }

    // Mirror the local Rill Developer flow: first deployment lands on
    // /-/invite (teammates next), subsequent ones land on /-/deploying.
    const params = new URLSearchParams();
    if (currentDashboard) {
      params.set("deploying_dashboard", currentDashboard);
    }
    const search = params.toString();
    const path = isFirstMerge ? "/-/invite" : "/-/deploying";
    targetWindow.location.href = `/${organization}/${project}${path}${
      search ? `?${search}` : ""
    }`;

    isMerging = false;
    open = false;
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
          {#if !hasProdDeployment}
            Merging
            <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
            sets up your production deployment. We'll open a new tab where you
            can invite teammates while it reconciles.
          {:else}
            Merging pushes changes from
            <span class="font-semibold text-fg-primary">"{currentBranch}"</span>
            to production. We'll open a new tab so you can watch updates
            reconcile.
          {/if}
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
      {:else if !primaryBranch || !currentBranch || !deploymentsLoaded}
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
