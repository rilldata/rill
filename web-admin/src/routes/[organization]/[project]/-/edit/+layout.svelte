<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetDeployment,
    createAdminServiceGetProject,
    V1DeploymentStatus,
    type V1Deployment,
    type V1Organization,
  } from "@rilldata/web-admin/client";
  import {
    extractBranchFromPath,
    branchPathPrefix,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { workspaceRoutePrefix } from "@rilldata/web-common/features/workspaces/workspace-routing";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { onDestroy } from "svelte";
  import { get } from "svelte/store";
  import ProjectHeader from "@rilldata/web-admin/features/projects/ProjectHeader.svelte";
  import EditSessionLoading from "@rilldata/web-admin/features/edit-session/EditSessionLoading.svelte";
  import EditSessionTimeoutBanner from "@rilldata/web-admin/features/edit-session/EditSessionTimeoutBanner.svelte";
  import { useDevDeploymentByBranch } from "@rilldata/web-admin/features/edit-session/use-edit-session";

  // Read params synchronously at initialization; they're stable for the
  // lifetime of this layout (navigating away from /-/edit/ destroys it).
  const { organization, project } = get(page).params;

  // Extract branch from the original URL (before reroute strips it)
  const branch = extractBranchFromPath(get(page).url.pathname);

  // Set the workspace route prefix for cloud editing
  $workspaceRoutePrefix = `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`;

  // Root layout data: org permissions, plan display name, organization object
  $: pageData = $page.data;
  $: organizationPermissions = pageData?.organizationPermissions ?? {};
  $: planDisplayName = pageData?.planDisplayName;
  $: organizationLogoUrl = getThemedLogoUrl(
    $themeControl,
    pageData?.organization as V1Organization | undefined,
  );

  // GetProject({branch}) for ProjectHeader metadata (project permissions, primary branch).
  // TanStack Query deduplicates with the project layout's identical query.
  const projectQuery = createAdminServiceGetProject(
    organization,
    project,
    branch ? { branch } : undefined,
  );
  $: projectPermissions = $projectQuery.data?.projectPermissions ?? {};
  $: primaryBranch = $projectQuery.data?.project?.primaryBranch;

  const user = createAdminServiceGetCurrentUser();
  const branchDeployment = useDevDeploymentByBranch(
    organization,
    project,
    branch,
  );
  const getCredentialsMutation = createAdminServiceGetDeployment();

  let runtimeHost: string | null = null;
  let instanceId: string | null = null;
  let accessToken: string | null = null;
  let credentialsError: string | null = null;

  $: currentUserId = $user.data?.user?.id;
  $: deployment = $branchDeployment.data;
  $: deploymentStatus = deployment?.status;

  $: isOtherOwner =
    !!deployment && !!currentUserId && deployment.ownerUserId !== currentUserId;

  // When deployment is running and owned by current user, fetch credentials
  $: if (
    deployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    deployment.id &&
    !isOtherOwner &&
    !runtimeHost &&
    !credentialsError
  ) {
    fetchCredentials(deployment);
  }

  $: isLoading =
    $branchDeployment.isLoading ||
    $user.isLoading ||
    $getCredentialsMutation.isPending ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING ||
    (deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
      !isOtherOwner &&
      !runtimeHost &&
      !credentialsError);

  $: isErrored =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;

  $: isReady =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    runtimeHost !== null &&
    !isOtherOwner;

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
        <ProjectHeader
          {organization}
          {project}
          {projectPermissions}
          manageOrgAdmins={organizationPermissions?.manageOrgAdmins}
          manageOrgMembers={organizationPermissions?.manageOrgMembers}
          readProjects={organizationPermissions?.readProjects}
          {primaryBranch}
          {planDisplayName}
          {organizationLogoUrl}
          editContext={{ deploymentId: deployment.id }}
        />
        <EditSessionTimeoutBanner sessionStartedAt={deployment.createdOn} />
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
  {:else}
    <ErrorPage
      statusCode={404}
      header="No active edit session"
      body="This editing session is no longer active. Use the Edit button to start a new one."
    />
  {/if}
</div>

<style lang="postcss">
  .edit-session {
    @apply flex flex-col h-full;
  }
</style>
