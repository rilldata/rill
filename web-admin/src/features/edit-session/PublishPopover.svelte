<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import {
    createAdminServiceCreateDeployment,
    createAdminServiceListDeployments,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { requestSkipBranchInjection } from "@rilldata/web-admin/features/branches/branch-utils";
  import { isProdDeployment } from "@rilldata/web-admin/features/branches/deployment-utils";
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
  $: hasProdDeployment =
    $deploymentsQuery.data?.deployments?.some(isProdDeployment) ?? false;
  // If the user is editing a dashboard, route them back to it after the first
  // publish reconcile completes. Matches `/-/edit/explore/<name>` and
  // `/-/edit/canvas/<name>`; undefined elsewhere (welcome flow, file editor).
  $: currentDashboard = (() => {
    const match = $page.url.pathname.match(
      /\/-\/edit\/(?:explore|canvas)\/([^/?#]+)/,
    );
    return match ? match[1] : undefined;
  })();
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

    const isFirstPublish = !hasProdDeployment;

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
      if (isFirstPublish) {
        // CreateDeployment with environment "prod" intentionally omits `branch`
        // and `editable`: the API derives the branch from the project's primary
        // branch and forbids both fields for prod deployments.
        await $createDeploymentMutation.mutateAsync({
          org: organization,
          project,
          data: { environment: "prod" },
        });
      }
    } catch (err) {
      errorMessage =
        getRpcErrorMessage(err as RpcStatus) ?? "Failed to publish";
      isPublishing = false;
      return;
    }

    let target: string;
    let gotoOpts: Parameters<typeof goto>[1];
    if (isFirstPublish) {
      // Prod runtime is provisioning from scratch and reconciling for the first
      // time; route through the deploying page so the user sees progress
      // instead of an empty project home.
      const params = new URLSearchParams({ source: "publish" });
      if (currentDashboard) {
        params.set("deploying_dashboard", currentDashboard);
      }
      target = `/${organization}/${project}/-/deploying?${params.toString()}`;
      gotoOpts = { replaceState: true };
    } else {
      target = `/${organization}/${project}`;
      gotoOpts = undefined;
    }

    isPublishing = false;
    open = false;

    // Defer goto to the next task. Calling it synchronously after a mutation
    // races with TanStack's invalidation/refetch teardown, whose abort listeners
    // can throw and silently cancel the navigation. Same workaround as
    // welcome/organization/+page.svelte after createOrg.
    setTimeout(() => {
      requestSkipBranchInjection();
      void goto(target, gotoOpts);
    });
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
          Publish your changes to production and return to the project home.
          Viewers will see updates as the project reconciles.
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
      {:else if !primaryBranch || !currentBranch}
        Loading project...
      {:else}
        Publish your latest changes
      {/if}
    </span>
  </TooltipContent>
</Tooltip>
