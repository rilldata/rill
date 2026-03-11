<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeployment,
    V1DeploymentStatus,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import {
    extractBranchFromPath,
    branchPathPrefix,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { workspaceRoutePrefix } from "@rilldata/web-common/features/workspaces/workspace-routing";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { onDestroy } from "svelte";
  import { get } from "svelte/store";
  import EditSessionLoading from "@rilldata/web-admin/features/edit-session/EditSessionLoading.svelte";
  import EditSessionTimeoutBanner from "@rilldata/web-admin/features/edit-session/EditSessionTimeoutBanner.svelte";
  import EditSessionToolbar from "@rilldata/web-admin/features/edit-session/EditSessionToolbar.svelte";
  import {
    invalidateDevDeployments,
    useDevDeploymentByBranch,
    useCreateDevDeployment,
  } from "@rilldata/web-admin/features/edit-session/use-edit-session";

  // Read params synchronously at initialization; they're stable for the
  // lifetime of this layout (navigating away from /-/edit/ destroys it).
  const { organization, project } = get(page).params;

  // Extract branch from the original URL (before reroute strips it)
  const branch = extractBranchFromPath(get(page).url.pathname);

  // Set the workspace route prefix for cloud editing
  $workspaceRoutePrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;

  const user = createAdminServiceGetCurrentUser();
  const branchDeployment = useDevDeploymentByBranch(
    organization,
    project,
    branch,
  );
  const createMutation = useCreateDevDeployment();

  let createAttempted = false;
  let runtimeHost: string | null = null;
  let instanceId: string | null = null;
  let accessToken: string | null = null;
  let credentialsError: string | null = null;

  const getCredentialsMutation = createAdminServiceGetDeployment();

  $: currentUserId = $user.data?.user?.id;
  $: deployment = $branchDeployment.data;

  // Check ownership: if the deployment belongs to another user, show error
  $: isOwnDeployment = !deployment || deployment.ownerUserId === currentUserId;
  $: isOtherOwner =
    !!deployment && !!currentUserId && deployment.ownerUserId !== currentUserId;

  // When there's no deployment for this branch, create one
  $: if (
    !$branchDeployment.isLoading &&
    !deployment &&
    !createAttempted &&
    !$createMutation.isPending &&
    !$user.isLoading
  ) {
    createAttempted = true;
    handleCreateDevDeployment();
  }

  // When we have an active running deployment owned by user, fetch credentials
  $: if (
    deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    deployment.id &&
    isOwnDeployment &&
    !runtimeHost
  ) {
    fetchCredentials(deployment);
  }

  $: deploymentStatus = deployment?.status;
  $: isLoading =
    $branchDeployment.isLoading ||
    $user.isLoading ||
    $createMutation.isPending ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING;

  $: isErrored =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;

  $: isReady =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    runtimeHost !== null &&
    isOwnDeployment;

  async function handleCreateDevDeployment() {
    try {
      await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
        },
      });
      void invalidateDevDeployments(organization, project);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start edit session: ${getRpcErrorMessage(err as any)}`,
      });
    }
  }

  async function fetchCredentials(dep: V1Deployment) {
    try {
      const resp = await $getCredentialsMutation.mutateAsync({
        deploymentId: dep.id!,
        data: {},
      });
      runtimeHost = resp.runtimeHost ?? null;
      instanceId = resp.instanceId ?? null;
      accessToken = resp.accessToken ?? null;
    } catch (err) {
      credentialsError =
        getRpcErrorMessage(err as any) ??
        "Failed to get deployment credentials";
    }
  }

  onDestroy(() => {
    $workspaceRoutePrefix = "";
  });
</script>

<div class="edit-session">
  {#if isOtherOwner}
    <!-- Edit URL for a branch the user doesn't own -->
    <CtaLayoutContainer>
      <CtaContentContainer>
        <h1
          class="text-8xl font-extrabold bg-gradient-to-b from-[#CBD5E1] to-[#E2E8F0] text-transparent bg-clip-text"
        >
          403
        </h1>
        <h2 class="text-lg font-semibold">
          This editing session belongs to another user
        </h2>
        <CtaMessage>You can preview this branch in read-only mode.</CtaMessage>
        <CtaButton
          variant="secondary"
          href="/{organization}/{project}{branchPathPrefix(branch)}"
        >
          Preview this branch
        </CtaButton>
      </CtaContentContainer>
    </CtaLayoutContainer>
  {:else if isReady && deployment?.id && instanceId && runtimeHost && accessToken}
    {#key `${runtimeHost}::${instanceId}`}
      <RuntimeProvider host={runtimeHost} {instanceId} jwt={accessToken}>
        <EditSessionTimeoutBanner sessionStartedAt={deployment.createdOn} />
        <EditSessionToolbar
          {organization}
          {project}
          deploymentId={deployment.id}
          {instanceId}
          {branch}
        />
        <FileAndResourceWatcher
          host={runtimeHost}
          {instanceId}
          errorBody="Lost connection to the editing environment. Try ending the session and starting a new one."
        >
          <div class="flex flex-1 overflow-hidden">
            <Navigation showFooterLinks={false} />
            <section class="flex-1 overflow-hidden">
              <slot />
            </section>
          </div>
        </FileAndResourceWatcher>
      </RuntimeProvider>
    {/key}
  {:else if isErrored}
    <ErrorPage
      statusCode={500}
      header="Edit session failed"
      body={deployment?.statusMessage ||
        "The editing environment encountered an error. Please try again."}
    />
  {:else if credentialsError}
    <ErrorPage
      statusCode={500}
      header="Failed to connect"
      body={credentialsError}
    />
  {:else if isLoading}
    <EditSessionLoading
      status={deploymentStatus}
      statusMessage={deployment?.statusMessage}
    />
  {/if}
</div>

<style lang="postcss">
  .edit-session {
    @apply flex flex-col h-full;
  }
</style>
