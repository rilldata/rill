<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { isActiveDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
  import { fetchProdParserCommitSha } from "@rilldata/web-admin/features/projects/selectors";
  import { Button } from "@rilldata/web-common/components/button";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    createRuntimeServiceGitMergeToBranchMutation,
    createRuntimeServiceGitPushMutation,
    createRuntimeServiceGitStatus,
    getRuntimeServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Rocket } from "lucide-svelte";
  import { buildPostMergeUrl } from "./post-merge-url";

  export let organization: string;
  export let project: string;
  export let primaryBranch: string | undefined;

  const AUTO_COMMIT_MESSAGE = "Updates from cloud editor";

  let isPublishing = false;

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
  $: alreadyOnPrimary =
    !!primaryBranch && !!currentBranch && currentBranch === primaryBranch;
  $: disabled =
    !primaryBranch ||
    !currentBranch ||
    !projectLoaded ||
    alreadyOnPrimary ||
    isPublishing;

  // Prefetch prod's project parser commit SHA once the deployment info is
  // available, so we can pass it to the deploying page at click time. The
  // page uses it to wait for the parser to advance past this point before
  // redirecting to the dashboard, avoiding the stale-content race.
  let prodParserSha: string | undefined;
  let prodParserShaPrefetched = false;
  $: if (!prodParserShaPrefetched && prodDeployment) {
    prodParserShaPrefetched = true;
    void fetchProdParserCommitSha(prodDeployment, $projectQuery.data?.jwt).then(
      (sha) => {
        prodParserSha = sha;
      },
    );
  }

  async function handlePublish() {
    if (!primaryBranch || isPublishing) return;
    isPublishing = true;

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

    // Open the new tab synchronously so the browser ties window.open() to
    // the click gesture; otherwise pop-up blockers reject the open after
    // the awaited mutations below. The destination pages handle their own
    // loading state, so no placeholder is needed.
    const targetUrl = buildPostMergeUrl({
      organization,
      project,
      pathname: $page.url.pathname,
      hadProdDeployment,
      preCommitSha: prodParserSha,
    });
    const targetWindow = window.open(targetUrl, "_blank");
    if (!targetWindow) {
      eventBus.emit("notification", {
        type: "error",
        message: "Pop-up was blocked. Please allow pop-ups and try again.",
      });
      isPublishing = false;
      return;
    }

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
      eventBus.emit("notification", {
        type: "error",
        message: extractErrorMessage(err) || "Failed to publish",
      });
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
        eventBus.emit("notification", {
          type: "error",
          message: `Changes merged to production, but starting the production deployment failed${
            detail ? `: ${detail}` : ""
          }.`,
        });
        isPublishing = false;
        return;
      }
    }

    isPublishing = false;
  }
</script>

<Tooltip distance={8}>
  <Button
    type="primary"
    {disabled}
    loading={isPublishing}
    loadingCopy="Publishing..."
    onClick={handlePublish}
  >
    <Rocket size="14" />
    Publish
  </Button>
  <TooltipContent slot="tooltip-content" maxWidth="240px">
    <span class="text-xs">
      {#if alreadyOnPrimary}
        Already on production
      {:else if !primaryBranch || !currentBranch || !projectLoaded}
        Loading project...
      {:else if !prodDeployment}
        Publish your project to production. We'll open a new tab where you can
        invite teammates while the deployment reconciles.
      {:else if !prodDeploymentActive}
        Production is hibernated. Publishing will resume it and apply your
        changes. We'll open the deployment in a new tab so you can watch updates
        reconcile.
      {:else}
        Publish your changes to production. We'll open a new tab so you can
        watch updates reconcile.
      {/if}
    </span>
  </TooltipContent>
</Tooltip>
