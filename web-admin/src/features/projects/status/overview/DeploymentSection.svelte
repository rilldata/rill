<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isFreePlan,
    isGrowthPlan,
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
  import { getGitUrlFromRemote } from "@rilldata/web-common/features/project/deploy/github-utils";
  import ProjectClone from "./ProjectClone.svelte";
  import ManageSlotsModal from "./ManageSlotsModal.svelte";
  import { detectTierSlots, MIN_INFRA_SLOTS } from "./slots-utils";
  import { useOlapInfo, isMotherDuck } from "./olapInfo";
  import OverviewCard from "./OverviewCard.svelte";
  import { page } from "$app/stores";

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
  $: cachedOlapType = projectData?.olapConnector;
  $: olapEngineLabel = olapConnector
    ? getOlapEngineLabel(olapConnector)
    : cachedOlapType
      ? formatConnectorName(cachedOlapType)
      : "DuckDB";
  $: aiConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.aiConnector,
  );

  // Slots
  $: currentSlots = Number(projectData?.prodSlots) || 0;
  // Live Connect only when a non-DuckDB connector explicitly has provision=false.
  // DuckDB is always Rill-managed (including local dev where provision=false).
  // When hibernated (no olapConnector), use cached type: clickhouse means Live Connect.
  $: isRillManaged = olapConnector
    ? (olapConnector.type === "duckdb" && !isMotherDuck(olapConnector)) ||
      olapConnector.provision === true
    : cachedOlapType
      ? cachedOlapType !== "clickhouse"
      : true;
  $: canManage = $proj.data?.projectPermissions?.manageProject ?? false;
  let slotsModalOpen = false;

  // Billing plan detection
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: planName = $subscriptionQuery?.data?.subscription?.plan?.name ?? "";
  $: isTrial = isTrialPlan(planName);
  $: isFree = isFreePlan(planName);
  $: isGrowth = isGrowthPlan(planName);
  $: isEnterprise = planName !== "" && isEnterprisePlan(planName);
  // New pricing applies to Free and Growth plans; ?newPricing=true forces it for testing
  $: devForceNewPricing = $page.url.searchParams.get("newPricing") === "true";
  $: useNewPricing = isFree || isGrowth || devForceNewPricing;

  // SQL-based cluster info: runs for any non-Rill-managed project to detect cluster slots.
  $: olapInfoQuery = useOlapInfo(runtimeClient, !isRillManaged ? olapConnector : undefined);
  $: olapInfo = $olapInfoQuery?.data;
  $: console.log("[olapInfo] useNewPricing:", useNewPricing, "| olapConnector:", olapConnector, "| query:", { isLoading: $olapInfoQuery?.isLoading, isError: $olapInfoQuery?.isError, error: $olapInfoQuery?.error, data: olapInfo });
  // Detected cluster slots from SQL (vcpus when available, else memory-tier fallback).
  $: detectedClusterSlots =
    olapInfo?.vcpus && olapInfo.vcpus > 0
      ? olapInfo.vcpus
      : detectTierSlots(parseMemoryToGb(olapInfo?.memory));

  // Backend quota overrides (set via `rill sudo project edit`)
  $: backendClusterSlots = Number(projectData?.clusterSlots) || undefined;

  // Cluster Slots: prefer SQL-detected value, fall back to RillMinSlots from backend.
  // Only applies to Live Connect (not Rill-managed).
  $: clusterSlots = !isRillManaged
    ? detectedClusterSlots ||
      Number(projectData?.clusterSlots) ||
      MIN_INFRA_SLOTS
    : 0;
  // Rill Slots = additional slots on top of cluster_slots (user-controlled).
  $: rillSlots =
    useNewPricing && !isRillManaged
      ? Math.max(0, currentSlots - clusterSlots)
      : 0;

  // Effective provisioned slots: for Live Connect, use detected cluster + rill;
  // for Rill Managed, use DB prod_slots directly.
  $: provisionedSlots = !isRillManaged
    ? clusterSlots + rillSlots
    : currentSlots;

  /**
   * Parses a human-readable memory string from the OLAP SQL queries into GB.
   * Handles formats like "8.00 GiB", "16.00 GB", "7.45 GiB".
   */
  function parseMemoryToGb(memory: string | undefined): number | undefined {
    if (!memory) return undefined;
    const m = memory.match(/^([\d.]+)\s*(GiB|GB|MiB|MB)/i);
    if (!m) return undefined;
    const value = parseFloat(m[1]);
    const unit = m[2].toLowerCase();
    if (unit === "gib" || unit === "gb") return value;
    if (unit === "mib" || unit === "mb") return value / 1024;
    return undefined;
  }
</script>

<OverviewCard title="Deployment">
  <div slot="header-right" class="flex items-center gap-3">
    {#if canManage && isFree && !$subscriptionQuery?.isLoading}
      <a
        class="upgrade-link"
        href="/{organization}/-/settings/billing"
      >
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
            {githubUrl.replace("https://github.com/", "")}
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
      <span class="info-value flex items-center gap-2">
        {olapEngineLabel}
        {#if olapInfo}
          <span class="text-fg-tertiary text-xs">
            ({olapInfo.vcpus} vCPU{olapInfo.vcpus !== 1 ? "s" : ""}, {olapInfo.memory}{olapInfo.replicas > 1 ? `, ${olapInfo.replicas} replicas` : ""})
          </span>
        {/if}
      </span>
    </div>

    <div class="info-row">
      <span class="info-label">AI</span>
      <span class="info-value">
        {#if aiConnector}
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
        <span class="info-label">Provisioned Slots</span>
        <span class="info-value flex items-center gap-3">
          <span class="slots-count">{provisionedSlots}</span>
          {#if canManage && !isTrial}
            <button
              class="manage-slots-btn"
              on:click={() => (slotsModalOpen = true)}
            >
              Manage
            </button>
          {/if}
        </span>
      </div>
    {/if}
  </div>
</OverviewCard>

<ManageSlotsModal
  bind:open={slotsModalOpen}
  {organization}
  {project}
  {currentSlots}
  {isRillManaged}
  viewOnly={isTrial}
  detectedSlots={detectedClusterSlots}
  {useNewPricing}
  clusterSlots={clusterSlots}
  currentRillSlots={rillSlots}
/>

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
  .slots-count {
    @apply text-sm text-fg-primary font-medium tabular-nums;
  }
  .upgrade-link {
    @apply text-xs text-primary-500 no-underline;
  }
  .upgrade-link:hover {
    @apply text-primary-600;
  }
  .manage-slots-btn {
    @apply text-xs text-primary-500 bg-transparent border-none cursor-pointer p-0 no-underline;
  }
  .manage-slots-btn:hover {
    @apply text-primary-600;
  }
</style>
