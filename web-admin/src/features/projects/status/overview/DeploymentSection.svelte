<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isEnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { useGithubLastSynced } from "@rilldata/web-admin/features/projects/selectors";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  import { useProjectDeployment, useRuntimeVersion } from "../selectors";
  import {
    formatEnvironmentName,
    formatConnectorName,
    getOlapEngineLabel,
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import { SLOT_TIERS } from "./slots-utils";
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import ProjectClone from "./ProjectClone.svelte";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();

  // Deployment
  $: projectDeployment = useProjectDeployment(organization, project);
  $: deployment = $projectDeployment.data;
  $: deploymentStatus =
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;

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
  // When hibernated the runtime is unreachable; fall back to the cached connector type from the admin DB.
  $: cachedOlapType = (projectData as any)?.olapConnector as string | undefined;
  $: olapEngineLabel = olapConnector
    ? getOlapEngineLabel(olapConnector)
    : cachedOlapType
      ? formatConnectorName(cachedOlapType)
      : "DuckDB";
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );

  // Slots / Cluster size
  $: currentSlots = Number(projectData?.prodSlots) || 0;
  $: currentTier = SLOT_TIERS.find((t) => t.slots === currentSlots);
  $: clusterLabel =
    currentTier?.instance ?? `${currentSlots * 4}GiB / ${currentSlots}vCPU`;
  $: canManage = $proj.data?.projectPermissions?.manageProject ?? false;

  // Billing plan detection
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: planName = $subscriptionQuery?.data?.subscription?.plan?.name ?? "";
  $: isFree = isTrialPlan(planName);
  $: isEnterprise = planName !== "" && isEnterprisePlan(planName);
</script>

<OverviewCard title="Deployment">
  <div slot="header-right" class="flex items-center gap-3">
    {#if canManage && isFree && !$subscriptionQuery?.isLoading}
      <a class="upgrade-link" href="/{organization}/-/settings/billing">
        Upgrade to Growth
      </a>
    {/if}
    <ProjectClone
      {organization}
      {project}
      gitRemote={projectData?.gitRemote}
      managedGitId={projectData?.managedGitId}
    />
  </div>

  <div class="info-grid">
    <div class="info-row">
      <span class="info-label">Status</span>
      <span class="info-value flex items-center gap-2">
        <span class="status-dot {getStatusDotClass(deploymentStatus)}"></span>
        {getStatusLabel(deploymentStatus)}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">Environment</span>
      <span class="info-value">
        {formatEnvironmentName(deployment?.environment)}
      </span>
    </div>

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
            {githubUrl?.replace("https://github.com/", "")}
          </a>
        </span>
      </div>
    {/if}

    {#if isGithubConnected && primaryBranch}
      <div class="info-row">
        <span class="info-label">Branch</span>
        <span class="info-value">{primaryBranch}</span>
      </div>
    {/if}

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
      <span class="info-value">
        {olapEngineLabel}
        {#if olapConnector}
          <span class="text-fg-tertiary text-xs ml-1"
            >({olapConnector.name})</span
          >
        {:else}
          <span class="text-fg-tertiary text-xs ml-1">(Rill Managed)</span>
        {/if}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">AI Connector</span>
      <span class="info-value">
        {#if aiConnector && aiConnector.name !== "admin"}
          {formatConnectorName(aiConnector.type)}
          <span class="text-fg-tertiary text-xs ml-1">({aiConnector.name})</span
          >
        {:else}
          Rill Managed
        {/if}
      </span>
    </div>

    {#if !$subscriptionQuery?.isLoading && !isEnterprise}
      <div class="info-row">
        <span class="info-label">Cluster Size</span>
        <span class="info-value flex items-center gap-3">
          <a
            href="/{organization}/{project}/-/status/deployments"
            class="slots-link"
          >
            <span class="slots-count">{clusterLabel}</span>
            <span class="slots-secondary"
              >({currentSlots} {currentSlots === 1 ? "slot" : "slots"})</span
            >
            <span class="slots-detail">View details</span>
          </a>
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
  .slots-link {
    @apply flex items-center gap-2 no-underline;
  }
  .slots-link:hover .slots-detail {
    @apply text-primary-600;
  }
  .slots-count {
    @apply text-sm text-fg-primary font-medium tabular-nums;
  }
  .slots-secondary {
    @apply text-xs text-fg-tertiary;
  }
  .slots-detail {
    @apply text-xs text-primary-500;
  }
  .upgrade-link {
    @apply text-xs text-primary-500 no-underline;
  }
  .upgrade-link:hover {
    @apply text-primary-600;
  }
</style>
