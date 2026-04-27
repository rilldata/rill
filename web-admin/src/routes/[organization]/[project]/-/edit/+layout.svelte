<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceGetProject,
    getAdminServiceGetProjectQueryKey,
    V1DeploymentStatus,
    type V1Organization,
  } from "@rilldata/web-admin/client";
  import {
    branchPathPrefix,
    extractBranchFromPath,
  } from "@rilldata/web-admin/features/branches/branch-utils";
  import BranchDeploymentStopped from "@rilldata/web-admin/features/branches/BranchDeploymentStopped.svelte";
  import EditSessionLoading from "@rilldata/web-admin/features/edit-session/EditSessionLoading.svelte";
  import EditSessionTimeoutBanner from "@rilldata/web-admin/features/edit-session/EditSessionTimeoutBanner.svelte";
  import ProjectHeader from "@rilldata/web-admin/features/projects/ProjectHeader.svelte";
  import SlimProjectHeader from "@rilldata/web-admin/features/projects/SlimProjectHeader.svelte";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { editorRoutePrefix } from "@rilldata/web-common/layout/navigation/editor-routing";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy } from "svelte";
  import { isProjectWelcomePage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";
  import WelcomeRedirector from "@rilldata/web-admin/features/welcome/project/WelcomeRedirector.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  // Extract branch from the current URL (before reroute strips it).
  $: branch = extractBranchFromPath($page.url.pathname);

  // Keep the workspace route prefix in sync for cloud editing.
  $: editorRoutePrefix.set(
    `/${organization}/${project}${branchPathPrefix(branch)}/-/edit`,
  );

  // Root layout data: org permissions, plan display name, organization object
  $: pageData = $page.data;
  $: organizationPermissions = pageData?.organizationPermissions ?? {};
  $: planDisplayName = pageData?.planDisplayName;
  $: organizationLogoUrl = getThemedLogoUrl(
    $themeControl,
    pageData?.organization as V1Organization | undefined,
  );

  // GetProject({branch}): returns deployment status, credentials (runtimeHost,
  // runtimeInstanceId, jwt), and project permissions. Polls at 2s while the
  // deployment is provisioning or updating; stops once it reaches a terminal
  // state (RUNNING, ERRORED, etc.). The parent layout skips its own polling
  // on the edit page to avoid duplicate requests.
  $: projectQuery = createAdminServiceGetProject(
    organization,
    project,
    branch ? { branch } : undefined,
    {
      query: {
        refetchInterval: (query) => {
          const status = query.state.data?.deployment?.status;
          if (
            status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
            status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING ||
            status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
            status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING
          ) {
            return 2000;
          }
          return false;
        },
      },
    },
  );
  $: projectPermissions = $projectQuery.data?.projectPermissions ?? {};
  $: primaryBranch = $projectQuery.data?.project?.primaryBranch;

  // Deployment data and credentials come from GetProject (no separate API needed)
  $: deployment = $projectQuery.data?.deployment;
  $: deploymentStatus = deployment?.status;
  $: runtimeHost = deployment?.runtimeHost ?? null;
  $: instanceId = deployment?.runtimeInstanceId ?? null;
  $: jwt = $projectQuery.data?.jwt ?? null;

  const user = createAdminServiceGetCurrentUser();

  $: currentUserId = $user.data?.user?.id;

  $: isOtherOwner =
    !!deployment && !!currentUserId && deployment.ownerUserId !== currentUserId;

  // Flipped when the user clicks "Start deployment" on a stopped deployment;
  // keeps the UI in loading state while the backend transitions STOPPED → PENDING → RUNNING.
  let starting = false;

  $: isLoading =
    $projectQuery.isLoading ||
    $user.isLoading ||
    starting ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING;

  $: isErrored =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;

  $: isStopped =
    !starting &&
    (deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING);

  $: isReady =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    runtimeHost !== null &&
    instanceId !== null &&
    jwt !== null &&
    !isOtherOwner;

  $: projectUrl = `/${organization}/${project}`;
  $: branchUrl = `/${organization}/${project}${branchPathPrefix(branch)}`;

  $: inProjectWelcomePage = isProjectWelcomePage($page);

  // Invalidating this query refetches a fresh JWT; `runtimeClient.getJwt()`
  // reads the updated value on the next call. Branch must be part of the
  // key or the invalidation misses the branch-scoped cache entry.
  const queryClient = useQueryClient();
  $: onBeforeReconnect = async () => {
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceGetProjectQueryKey(
        organization,
        project,
        branch ? { branch } : undefined,
      ),
    });
  };

  onDestroy(() => {
    $editorRoutePrefix = "";
  });
</script>

<div class="edit-session">
  {#if isOtherOwner}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      readDev={!!projectPermissions?.readDev}
      {primaryBranch}
      {planDisplayName}
      {organizationLogoUrl}
    />
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
        <CtaButton variant="secondary" href={branchUrl}>
          Preview this branch
        </CtaButton>
      </CtaContentContainer>
    </CtaLayoutContainer>
  {:else if isReady && deployment?.id && instanceId && runtimeHost && jwt}
    {#key `${runtimeHost}::${instanceId}`}
      <RuntimeProvider host={runtimeHost} {instanceId} {jwt}>
        {#if !inProjectWelcomePage}
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
            editContext={true}
          />
          <EditSessionTimeoutBanner sessionStartedAt={deployment.createdOn} />
        {/if}
        <FileAndResourceWatcher
          lifecycle="none"
          {onBeforeReconnect}
          errorBody="Lost connection to the editing environment. Try ending the session and starting a new one."
        >
          <div class="flex flex-1 overflow-hidden">
            {#if !inProjectWelcomePage}
              <WelcomeRedirector />
              <Navigation showFooterLinks={false} />
            {/if}
            <section class="flex flex-1 overflow-hidden">
              <div class="flex-1 overflow-hidden">
                <slot />
              </div>
              <DeveloperChat />
            </section>
          </div>
        </FileAndResourceWatcher>
      </RuntimeProvider>
    {/key}
  {:else if isErrored}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      readDev={!!projectPermissions?.readDev}
      {primaryBranch}
      {planDisplayName}
      {organizationLogoUrl}
    />
    <ErrorPage
      statusCode={500}
      header="Edit session failed"
      body={deployment?.statusMessage ||
        "The editing environment encountered an error. Please try again."}
    />
  {:else if isStopped && deployment?.id}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      readDev={!!projectPermissions?.readDev}
      {primaryBranch}
      {planDisplayName}
      {organizationLogoUrl}
    />
    <BranchDeploymentStopped
      {organization}
      {project}
      deploymentId={deployment.id}
      status={deploymentStatus}
      canManage={!!projectPermissions?.manageDev}
      {branch}
      onStarted={() => (starting = true)}
    />
  {:else if isLoading}
    <EditSessionLoading status={deploymentStatus} cancelHref={projectUrl} />
  {:else}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
      readDev={!!projectPermissions?.readDev}
      {primaryBranch}
      {planDisplayName}
      {organizationLogoUrl}
    />
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
