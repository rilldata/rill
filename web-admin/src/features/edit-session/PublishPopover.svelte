<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateDeployment,
    createAdminServiceListDeployments,
    getAdminServiceListDeploymentsQueryKey,
  } from "@rilldata/web-admin/client";
  import { isProdDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
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
  const deploymentsQuery = createAdminServiceListDeployments(
    organization,
    project,
    {},
  );
  const createDeploymentMutation = createAdminServiceCreateDeployment();

  $: currentBranch = $gitStatusQuery.data?.branch ?? "";
  $: hasLocalChanges = $gitStatusQuery.data?.localChanges ?? false;
  // First publish only: managed-git projects are created with `skipDeploy`,
  // so on the first merge to primary we must explicitly create the prod
  // deployment. Subsequent publishes find a prod deployment and skip this.
  $: deploymentsLoaded = $deploymentsQuery.data !== undefined;
  $: hasProdDeployment =
    $deploymentsQuery.data?.deployments?.some(isProdDeployment) ?? false;
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
    !deploymentsLoaded ||
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

    const isFirstPublish = !hasProdDeployment;

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

    if (isFirstPublish) {
      try {
        // CreateDeployment with environment "prod" intentionally omits `branch`
        // and `editable`: the API derives the branch from the project's primary
        // branch and forbids both fields for prod deployments.
        // TODO: detect billing/quota errors here and surface a friendly
        // message, mirroring `getPrettyDeployError` in
        // `web-common/src/features/project/deploy/deploy-errors.ts`.
        await $createDeploymentMutation.mutateAsync({
          org: organization,
          project,
          data: { environment: "prod" },
        });
        // Invalidate via the no-params base key so all deployment-list query
        // variants for this project (BranchSelector, BranchesSection, etc.)
        // are matched by TanStack's prefix matching.
        await queryClient.invalidateQueries({
          queryKey: getAdminServiceListDeploymentsQueryKey(
            organization,
            project,
          ),
        });
      } catch (err) {
        // The merge succeeded but the prod deployment failed to start. Be
        // explicit so the user knows their changes are on the primary branch
        // and that creating the deployment is the part that needs retrying.
        targetWindow.close();
        const detail = extractErrorMessage(err);
        errorMessage = `Changes merged to production, but creating the production deployment failed${
          detail ? `: ${detail}` : ""
        }.`;
        isPublishing = false;
        return;
      }
    }

    // Mirror the local Rill Developer flow: first publish lands on /-/invite
    // (inviting teammates is the natural next step before the runtime is
    // ready), subsequent publishes land on /-/deploying.
    const params = new URLSearchParams();
    if (currentDashboard) {
      params.set("deploying_dashboard", currentDashboard);
    }
    const search = params.toString();
    const path = isFirstPublish ? "/-/invite" : "/-/deploying";
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
          {#if !hasProdDeployment}
            Publish your project to production. We'll open a new tab where you
            can invite teammates while the deployment reconciles.
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
      {:else if !primaryBranch || !currentBranch || !deploymentsLoaded}
        Loading project...
      {:else}
        Publish your latest changes
      {/if}
    </span>
  </TooltipContent>
</Tooltip>
