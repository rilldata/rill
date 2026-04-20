<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    isFreePlan,
    isProPlan,
    isTrialPlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    useParserReconcileError,
    useProjectDeployment,
    useRuntimeVersion,
  } from "../selectors";
  import {
    formatEnvironmentName,
    formatConnectorName,
    getOlapEngineLabel,
    getStatusDotClass,
    getStatusLabel,
    isTransitoryStatus,
  } from "../display-utils";
  import LoadingCircleOutline from "@rilldata/web-common/components/icons/LoadingCircleOutline.svelte";
  import Callout from "@rilldata/web-common/components/callout/Callout.svelte";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import ProjectClone from "./ProjectClone.svelte";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import ClusterSize from "./ClusterSize.svelte";

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();

  $: activeBranch = extractBranchFromPath($page.url.pathname);

  // Deployment
  $: projectDeployment = useProjectDeployment(
    organization,
    project,
    activeBranch,
  );
  $: deployment = $projectDeployment.data;
  $: deploymentStatus =
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;

  // ProjectParser — detects project-level failures (e.g. git branch not found)
  $: parserErrorQuery = useParserReconcileError(runtimeClient);
  $: parserReconcileError = $parserErrorQuery.data ?? "";

  // Project
  $: proj = createAdminServiceGetProject(organization, project);
  $: projectData = $proj.data?.project;
  $: primaryBranch = projectData?.primaryBranch;
  // Last synced
  $: githubLastSynced = useGithubLastSynced(runtimeClient);
  $: dashboardsLastUpdated = useDashboardsLastUpdated(
    runtimeClient,
    organization,
    project,
  );
  $: lastUpdated = $githubLastSynced.data ?? $dashboardsLastUpdated;

  // Runtime
  $: runtimeVersionQuery = useRuntimeVersion(runtimeClient);
  $: version = $runtimeVersionQuery.data?.version?.match(/v[\d.]+/)?.[0] ?? "";

  // Connectors — sensitive: true is needed to read projectConnectors (OLAP/AI connector types)
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  // Repo — only shown when the user connected their own GitHub
  $: githubUrl = projectData?.gitRemote
    ? getGitUrlFromRemote(projectData.gitRemote)
    : "";
  $: isGithubConnected =
    !!projectData?.gitRemote && !projectData?.managedGitId && !!githubUrl;

  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: olapEngineLabel = getOlapEngineLabel(olapConnector);
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );

  // Slots
  $: currentSlots = Number(projectData?.prodSlots) || 0;

  // Billing plan detection
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: planName = $subscriptionQuery?.data?.subscription?.plan?.name ?? "";
  $: showSlots =
    isTrialPlan(planName) || isFreePlan(planName) || isProPlan(planName);
</script>

<OverviewCard title="Deployment">
  <div slot="header-right" class="flex items-center gap-3">
    <!-- TODO: re-add "Upgrade to Pro" link when ready.
         Gate on: canManage && (isTrialPlan || isFreePlan || isTeamPlan) && !subscriptionQuery.isLoading
    -->
    <ProjectClone
      {organization}
      {project}
      gitRemote={projectData?.gitRemote}
      managedGitId={projectData?.managedGitId}
      disabled={!!parserReconcileError}
    />
  </div>

  <div class="info-grid">
    <div class="info-row">
      <span class="info-label">Status</span>
      <span class="info-value flex items-center gap-2">
        {#if isTransitoryStatus(deploymentStatus)}
          <LoadingCircleOutline size="12px" />
        {:else}
          <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
        {/if}
        {getStatusLabel(deploymentStatus)}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">Environment</span>
      <span class="info-value">
        {formatEnvironmentName(deployment?.environment)}
      </span>
    </div>

    {#if !$subscriptionQuery?.isLoading && showSlots}
      <div class="info-row">
        <span class="info-label">Cluster Size</span>
        <span class="info-value">
          <ClusterSize slots={currentSlots} />
        </span>
      </div>
    {/if}

    {#if isGithubConnected}
      <div class="info-row">
        <span class="info-label">Repo</span>
        <span class="info-value">
          <a
            href={githubUrl}
            target="_blank"
            rel="noopener noreferrer"
            class="repo-link"
          >
            {githubUrl.replace("https://github.com/", "")}
          </a>
        </span>
      </div>
    {/if}

    {#if isGithubConnected && (deployment?.branch || primaryBranch)}
      <div class="info-row">
        <span class="info-label">Branch</span>
        <span class="info-value">{deployment?.branch || primaryBranch}</span>
      </div>
    {/if}

    {#if parserReconcileError}
      <!-- Project failed to load: show the error instead of project-level details
           (Last synced, OLAP, AI) which would be stale defaults -->
      <div class="mt-2">
        <Callout level="error">
          <span class="text-sm">{parserReconcileError}</span>
        </Callout>
      </div>
    {:else}
      {#if lastUpdated}
        <div class="info-row">
          <span class="info-label">Last synced</span>
          <span class="info-value">
            {lastUpdated.toLocaleString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
              hour: "numeric",
              minute: "numeric",
            })}
          </span>
        </div>
      {/if}

      {#if version}
        <div class="info-row">
          <span class="info-label">Runtime</span>
          <span class="info-value">{version}</span>
        </div>
      {/if}

      <div class="info-row">
        <span class="info-label">OLAP Engine</span>
        <span class="info-value">{olapEngineLabel}</span>
      </div>

      <div class="info-row">
        <span class="info-label">AI Connector</span>
        <span class="info-value">
          {#if aiConnector && aiConnector.name !== "admin"}
            {formatConnectorName(aiConnector.type)}
            <span class="text-fg-tertiary text-xs ml-1"
              >({aiConnector.name})</span
            >
          {:else}
            Rill Managed
          {/if}
        </span>
      </div>
    {/if}
  </div>
</OverviewCard>

<style lang="postcss">
  .info-grid {
    @apply flex flex-col;
  }
  .info-row {
    @apply flex items-center py-2;
  }
  .info-row:last-child {
    @apply border-b-0;
  }
  .info-label {
    @apply text-sm text-fg-secondary w-32 shrink-0;
  }
  .info-value {
    @apply text-sm text-fg-primary;
  }
  .status-dot {
    @apply w-2 h-2 rounded-full inline-block;
  }
  .repo-link {
    @apply text-primary-500 text-sm;
  }
  .repo-link:hover {
    @apply underline;
  }
</style>
