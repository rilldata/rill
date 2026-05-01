<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { isActiveDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Popover from "@rilldata/web-common/components/popover";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitPushMutation,
    createRuntimeServiceGitStatus,
    getRuntimeServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Rocket } from "lucide-svelte";

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
  // Query GetProject without a branch param so `data.deployment` reflects
  // the project's primary (prod) deployment — the same source of truth the
  // project layout uses. ListDeployments is too loose: it includes orphan
  // prod records whose project pointer has been cleared.
  const projectQuery = createAdminServiceGetProject(organization, project);
  const redeployProjectMutation = createAdminServiceRedeployProject();

  $: currentBranch = $gitStatusQuery.data?.branch ?? "";
  $: hasLocalChanges = $gitStatusQuery.data?.localChanges ?? false;
  $: projectLoaded = $projectQuery.data !== undefined;
  $: prodDeployment = $projectQuery.data?.deployment;
  $: prodDeploymentActive =
    !!prodDeployment && isActiveDeployment(prodDeployment);
  // Pull the dashboard name from `/-/edit/explore/<name>` or
  // `/-/edit/canvas/<name>` so first publish can route the user back to it.
  // Other edit paths (file editor, welcome) fall through to undefined.
  $: currentDashboard = $page.url.pathname.match(
    /\/-\/edit\/(?:explore|canvas)\/([^/?#]+)/,
  )?.[1];
  $: alreadyOnPrimary =
    !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;
  $: disabled =
    !primaryBranch ||
    !currentBranch ||
    !projectLoaded ||
    alreadyOnPrimary ||
    isPublishing;

  $: if (!open) {
    errorMessage = "";
  }

  async function handlePublish() {
    if (!primaryBranch || isPublishing) return;
    isPublishing = true;
    errorMessage = "";

    // Open the destination tab synchronously so the browser ties it to the
    // user's click gesture; otherwise pop-up blockers reject the open after
    // the awaited mutations below. We navigate the placeholder once we know
    // whether this is a first publish (-> /-/invite) or a subsequent one
    // (-> /-/deploying), and close it on failure.
    const targetWindow = window.open("about:blank", "_blank");
    if (!targetWindow) {
      errorMessage = "Pop-up was blocked. Please allow pop-ups and try again.";
      isPublishing = false;
      return;
    }

    // Snapshot the prod deployment state at click time. Three relevant cases:
    //   1. No prod deployment yet → first publish; route to /-/invite.
    //   2. Prod deployment exists but is dormant (STOPPED/ERRORED) → wake it
    //      via RedeployProject; route to /-/deploying.
    //   3. Prod deployment is already active (PENDING/RUNNING/UPDATING) → the
    //      merge alone reconciles changes; route to /-/deploying.
    // RedeployProject (admin/projects.go:333) handles cases 1 and 2 with a
    // single RPC: if there's no current deployment it creates one, otherwise
    // it provisions a fresh one and tears down the old.
    const hadProdDeployment = !!prodDeployment;
    const needsRedeploy = !prodDeploymentActive;

    // gitStatus tracks localChanges and currentBranch; the mutations below
    // change both, so refresh once the flow finishes (success or failure).
    const gitStatusQueryKey = getRuntimeServiceGitStatusQueryKey(
      client.instanceId,
      {},
    );

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
      targetWindow.close();
      errorMessage = extractErrorMessage(err) || "Failed to publish";
      isPublishing = false;
      return;
    } finally {
      // Either gitPush or gitMerge may have changed localChanges or
      // currentBranch (success or partial failure). Invalidate once.
      void queryClient.invalidateQueries({ queryKey: gitStatusQueryKey });
    }

    if (needsRedeploy) {
      try {
        // TODO: detect billing/quota errors here and surface a friendly
        // message, mirroring `getPrettyDeployError` in
        // `web-common/src/features/project/deploy/deploy-errors.ts`.
        await $redeployProjectMutation.mutateAsync({
          org: organization,
          project,
        });
        // Refresh the project query so the layout sees the new primary
        // deployment pointer, and ListDeployments so subscribers like
        // BranchSelector and BranchesSection pick up the new row.
        await Promise.all([
          queryClient.invalidateQueries({
            queryKey: getAdminServiceGetProjectQueryKey(organization, project),
          }),
          queryClient.invalidateQueries({
            queryKey: getAdminServiceListDeploymentsQueryKey(
              organization,
              project,
            ),
          }),
        ]);
      } catch (err) {
        // The merge succeeded but the prod deployment failed to (re)start.
        // Be explicit so the user knows their changes are on the primary
        // branch and that the deployment step is what needs retrying.
        targetWindow.close();
        const detail = extractErrorMessage(err);
        errorMessage = `Changes merged to production, but starting the production deployment failed${
          detail ? `: ${detail}` : ""
        }.`;
        isPublishing = false;
        return;
      }
    }

    // Mirror the local Rill Developer flow: a project that has never had a
    // prod deployment lands on /-/invite (inviting teammates is the natural
    // next step before the runtime is ready); everything else lands on
    // /-/deploying.
    const params = new URLSearchParams();
    if (currentDashboard) {
      params.set("deploying_dashboard", currentDashboard);
    }
    const search = params.toString();
    const path = hadProdDeployment ? "/-/deploying" : "/-/invite";
    targetWindow.location.href = `/${organization}/${project}${path}${
      search ? `?${search}` : ""
    }`;

    isPublishing = false;
    open = false;
  }
</script>

<Tooltip distance={8} suppress={open}>
  <Popover.Root bind:open>
    <Popover.Trigger>
      {#snippet child({ props })}
        <Button {...props} type="primary" {disabled}>
          <Rocket size="14" />
          Publish
        </Button>
      {/snippet}
    </Popover.Trigger>
    <Popover.Content align="end" class="!w-[320px]">
      <div class="flex flex-col gap-y-3">
        <p class="text-xs text-fg-secondary">
          {#if !prodDeployment}
            Publish your project to production. We'll open a new tab where you
            can invite teammates while the deployment reconciles.
          {:else if !prodDeploymentActive}
            Wake your project and publish your changes. We'll open the
            deployment in a new tab so you can watch updates reconcile.
          {:else}
            Publish your changes to production. We'll open a new tab so you can
            watch updates reconcile.
          {/if}
        </p>
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
      {:else if !primaryBranch || !currentBranch || !projectLoaded}
        Loading project...
      {:else}
        Publish your latest changes
      {/if}
    </span>
  </TooltipContent>
</Tooltip>
