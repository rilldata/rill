<script lang="ts">
  import { page } from "$app/stores";
  import {
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
  import ProjectHeader from "../../../../../features/projects/header/ProjectHeader.svelte";
  import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/project-query-options";
  import SlimProjectHeader from "@rilldata/web-admin/features/projects/SlimProjectHeader.svelte";
  import { getThemedLogoUrl } from "@rilldata/web-admin/features/themes/organization-logo";
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
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import BlockingOverlayContainer from "@rilldata/web-common/layout/BlockingOverlayContainer.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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

  $: primaryProjectQuery = createAdminServiceGetProject(
    organization,
    project,
    undefined,
    { query: baseGetProjectQueryOptions },
  );
  $: hasPrimaryDeployment =
    !!$primaryProjectQuery.data?.project?.primaryDeploymentId;

  // Deployment data and credentials come from GetProject (no separate API needed)
  $: deployment = $projectQuery.data?.deployment;
  $: deploymentStatus = deployment?.status;
  $: runtimeHost = deployment?.runtimeHost ?? null;
  $: instanceId = deployment?.runtimeInstanceId ?? null;
  $: jwt = $projectQuery.data?.jwt ?? null;

  // Flipped when the user clicks "Start deployment" on a stopped deployment;
  // keeps the UI in loading state while the backend transitions STOPPED → PENDING → RUNNING.
  let starting = false;

  // Wait for `primaryProjectQuery` too: the cloud readonly notice is gated on
  // `hasPrimaryDeployment`, and it must be registered before `<slot />` renders
  // any file editor (getReadonlyNotice reads the notice non-reactively).
  $: isLoading =
    $projectQuery.isPending ||
    $primaryProjectQuery.isPending ||
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
    jwt !== null;

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

  // Only surface the env notice once the project has a primary deployment.
  // Fail closed: env becomes editable only once we've positively confirmed the
  // project has no primary deployment. A failed or otherwise inconclusive lookup
  // keeps the notice set, so a published project never exposes editable `.env`
  // files while the deployment state is unknown.
  // `isLoading` blocks `<slot />` until `primaryProjectQuery` resolves, so this
  // has run before any file editor reads the notice.
  $: if (!$primaryProjectQuery.isPending) {
    const envEditable = $primaryProjectQuery.isSuccess && !hasPrimaryDeployment;
    setCloudReadonlyNotice(envEditable ? undefined : envEditDisabled);
    fileArtifacts.recheckReadonlyStatus();
  }

  onDestroy(() => {
    $editorRoutePrefix = "";
  });
</script>

<div class="edit-session">
  {#if isLoading}
    <EditSessionLoading status={deploymentStatus} href={`/${organization}`} />
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
      bind:starting
    />
  {:else if isReady && deployment?.id && instanceId && runtimeHost && jwt}
    {#key `${runtimeHost}::${instanceId}::${hasPrimaryDeployment}`}
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

{#if $overlay !== null}
  <BlockingOverlayContainer
    bg="linear-gradient(to right, rgba(0,0,0,.6), rgba(0,0,0,.8))"
  >
    <div slot="title" class="font-bold">
      {$overlay?.title}
    </div>
    <svelte:fragment slot="detail">
      {#if $overlay?.detail}
        <svelte:component
          this={$overlay.detail.component}
          {...$overlay.detail.props}
        />
      {/if}
    </svelte:fragment>
  </BlockingOverlayContainer>
{/if}

{#snippet envEditDisabled()}
  <div class="flex flex-row gap-2 items-center w-fit text-sm">
    <InfoIcon size={14} />
    {m.edit_env_manage_in_settings()}
    <a
      href="/{organization}/{project}/-/settings/environment-variables"
      target="_blank"
      rel="noopener"
    >
      {m.edit_env_settings_link()}
    </a>
  </div>
{/snippet}

<style lang="postcss">
  .edit-session {
    @apply flex flex-col h-full;
  }
</style>
