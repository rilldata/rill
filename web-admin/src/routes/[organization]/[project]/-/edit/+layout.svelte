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
  import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/project-query-options";
  import SlimProjectHeader from "@rilldata/web-admin/features/projects/SlimProjectHeader.svelte";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo";
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { editorRoutePrefix } from "@rilldata/web-common/layout/navigation/editor-routing";
  import RuntimeProvider from "@rilldata/web-common/runtime-client/v2/RuntimeProvider.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { onDestroy } from "svelte";
  import { setCloudReadonlyNotice } from "@rilldata/web-common/features/entity-management/actions/protected-files.ts";
  import { isProjectWelcomePage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";
  import WelcomeRedirector from "@rilldata/web-admin/features/welcome/project/WelcomeRedirector.svelte";
  import { InfoIcon } from "lucide-svelte";

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

  // Polling and JWT-refresh cadence are governed by `baseGetProjectQueryOptions`,
  // shared with the parent project layout so both observers stay in sync.
  $: projectQuery = createAdminServiceGetProject(
    organization,
    project,
    branch ? { branch } : undefined,
    { query: baseGetProjectQueryOptions },
  );
  $: projectPermissions = $projectQuery.data?.projectPermissions ?? {};
  $: primaryBranch = $projectQuery.data?.project?.primaryBranch;
  $: devTtlSeconds = $projectQuery.data?.project?.devTtlSeconds;

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
    $projectQuery.isPending ||
    $user.isPending ||
    starting ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING;

  $: isErrored =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;

  $: isStopped =
    !starting &&
    (deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING);

  $: isReady =
    (deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
      deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING) &&
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

  setCloudReadonlyNotice(envEditDisabled);

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
          <EditSessionTimeoutBanner
            usedOn={deployment.usedOn}
            {devTtlSeconds}
          />
        {/if}
        <FileAndResourceWatcher
          lifecycle="none"
          {onBeforeReconnect}
          errorBody="Lost connection to the editing environment. Try ending the session and starting a new one."
        >
          {#if !inProjectWelcomePage}
            <WelcomeRedirector />
          {/if}
          <slot />
        </FileAndResourceWatcher>
      </RuntimeProvider>
    {/key}
  {:else if isErrored}
    <SlimProjectHeader
      {organization}
      {project}
      readProjects={organizationPermissions?.readProjects}
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

{#snippet envEditDisabled()}
  <div class="flex flex-row gap-2 items-center w-fit text-sm">
    <InfoIcon size={14} /> Manage environment variables in
    <a
      href="/{organization}/{project}/-/settings/environment-variables"
      target="_blank"
      rel="noopener"
    >
      Settings →
    </a>
  </div>
{/snippet}

<style lang="postcss">
  .edit-session {
    @apply flex flex-col h-full;
  }
</style>
