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
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useProjectDeployment } from "../selectors";
  import {
    getStatusDotClass,
    getStatusLabel,
  } from "../display-utils";
  import {
    detectTierSlots,
    MIN_INFRA_SLOTS,
    MANAGED_SLOT_RATE_PER_HR,
    CLUSTER_SLOT_RATE_PER_HR,
    RILL_SLOT_RATE_PER_HR,
    HOURS_PER_MONTH,
  } from "../overview/slots-utils";
  import { useOlapInfo, isMotherDuck } from "../overview/olapInfo";
  import ManageSlotsModal from "../overview/ManageSlotsModal.svelte";
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

  // Connectors
  $: instanceQuery = createRuntimeServiceGetInstance(runtimeClient, {
    sensitive: true,
  });
  $: instance = $instanceQuery.data?.instance;
  $: olapConnector = instance?.projectConnectors?.find(
    (c) => c.name === instance?.olapConnector,
  );
  $: cachedOlapType = projectData?.olapConnector;

  // Slots
  $: currentSlots = Number(projectData?.prodSlots) || 0;
  $: isRillManaged = olapConnector
    ? (olapConnector.type === "duckdb" && !isMotherDuck(olapConnector)) ||
      olapConnector.provision === true
    : cachedOlapType
      ? cachedOlapType !== "clickhouse"
      : true;
  $: canManage = $proj.data?.projectPermissions?.manageProject ?? false;
  let slotsModalOpen = false;

  // Billing
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: planName = $subscriptionQuery?.data?.subscription?.plan?.name ?? "";
  $: isTrial = isTrialPlan(planName);
  $: isFree = isFreePlan(planName);
  $: isGrowth = isGrowthPlan(planName);
  $: isEnterprise = planName !== "" && isEnterprisePlan(planName);
  $: devForceNewPricing = $page.url.searchParams.get("newPricing") === "true";
  $: useNewPricing = isFree || isGrowth || devForceNewPricing;

  // OLAP cluster info
  $: olapInfoQuery = useOlapInfo(
    runtimeClient,
    !isRillManaged ? olapConnector : undefined,
  );
  $: olapInfo = $olapInfoQuery?.data;
  $: isRefreshingCluster = $olapInfoQuery?.isFetching && !$olapInfoQuery?.isLoading;

  function refreshClusterSlots() {
    $olapInfoQuery?.refetch();
  }
  $: detectedClusterSlots =
    olapInfo?.vcpus && olapInfo.vcpus > 0
      ? olapInfo.vcpus
      : detectTierSlots(parseMemoryToGb(olapInfo?.memory));

  // Computed slots
  $: clusterSlots = !isRillManaged
    ? detectedClusterSlots ||
      Number(projectData?.clusterSlots) ||
      MIN_INFRA_SLOTS
    : 0;
  $: rillSlots =
    useNewPricing && !isRillManaged
      ? Math.max(0, currentSlots - clusterSlots)
      : isRillManaged
        ? currentSlots
        : 0;
  $: provisionedSlots = !isRillManaged
    ? clusterSlots + rillSlots
    : currentSlots;

  // Estimated costs
  $: rillSlotRate = isRillManaged
    ? MANAGED_SLOT_RATE_PER_HR
    : RILL_SLOT_RATE_PER_HR;
  $: clusterMonthlyCost = Math.round(
    clusterSlots * CLUSTER_SLOT_RATE_PER_HR * HOURS_PER_MONTH,
  );
  $: rillMonthlyCost = Math.round(rillSlots * rillSlotRate * HOURS_PER_MONTH);
  $: totalMonthlyCost = clusterMonthlyCost + rillMonthlyCost;

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

<div class="page">
  <!-- Header -->
  <div class="page-header">
    <div class="flex items-center gap-3">
      <h2 class="page-title">Deployments</h2>
      <span class="status-badge {getStatusDotClass(deploymentStatus)}">
        {getStatusLabel(deploymentStatus)}
      </span>
    </div>
    {#if canManage && !isTrial && !isEnterprise && !$subscriptionQuery?.isLoading}
      <button
        class="manage-btn"
        on:click={() => (slotsModalOpen = true)}
      >
        Manage Slots
      </button>
    {/if}
  </div>

  <!-- Slot summary cards -->
  <div class="slot-cards">
    <div class="slot-card">
      <span class="slot-card-label">Provisioned Slots</span>
      <span class="slot-card-value">{provisionedSlots}</span>
      <span class="slot-card-sub">
        ~${totalMonthlyCost.toLocaleString()}/mo
      </span>
    </div>

    {#if !isRillManaged}
      <div class="slot-card">
        <div class="slot-card-header">
          <span class="slot-card-label">Cluster Slots</span>
          <button
            class="refresh-btn"
            class:refreshing={isRefreshingCluster}
            on:click={refreshClusterSlots}
            disabled={isRefreshingCluster}
            title="Refresh from OLAP cluster"
          >
            <svg class="refresh-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M13.65 2.35A7.96 7.96 0 0 0 8 0C3.58 0 0 3.58 0 8s3.58 8 8 8c3.73 0 6.84-2.55 7.73-6h-2.08A5.99 5.99 0 0 1 8 14 6 6 0 1 1 8 2c1.66 0 3.14.69 4.22 1.78L9 7h7V0l-2.35 2.35Z" fill="currentColor"/>
            </svg>
          </button>
        </div>
        <span class="slot-card-value">{clusterSlots}</span>
        <span class="slot-card-sub">
          @ ${CLUSTER_SLOT_RATE_PER_HR}/slot/hr
          (~${clusterMonthlyCost.toLocaleString()}/mo)
        </span>
      </div>
    {/if}

    <div class="slot-card">
      <span class="slot-card-label">Rill Slots</span>
      <span class="slot-card-value">{rillSlots}</span>
      <span class="slot-card-sub">
        @ ${rillSlotRate}/slot/hr
        (~${rillMonthlyCost.toLocaleString()}/mo)
      </span>
    </div>
  </div>

  <!-- Slot breakdown detail -->
  <section class="detail-section">
    <h3 class="section-title">Slot Breakdown</h3>
    <div class="detail-table">
      <div class="detail-header">
        <span class="detail-cell detail-cell-name">Type</span>
        <span class="detail-cell">Slots</span>
        <span class="detail-cell">Rate</span>
        <span class="detail-cell">Est. Monthly</span>
        <span class="detail-cell">Source</span>
      </div>

      {#if !isRillManaged}
        <div class="detail-row">
          <span class="detail-cell detail-cell-name">
            Cluster Slots
          </span>
          <span class="detail-cell font-medium">{clusterSlots}</span>
          <span class="detail-cell">${CLUSTER_SLOT_RATE_PER_HR}/slot/hr</span>
          <span class="detail-cell">~${clusterMonthlyCost.toLocaleString()}</span>
          <span class="detail-cell">
            <span class="source-badge source-auto">
              {#if detectedClusterSlots}
                Auto-detected
              {:else if projectData?.clusterSlots}
                Backend override
              {:else}
                Default ({MIN_INFRA_SLOTS})
              {/if}
            </span>
          </span>
        </div>
      {/if}

      <div class="detail-row">
        <span class="detail-cell detail-cell-name">
          Rill Slots
        </span>
        <span class="detail-cell font-medium">{rillSlots}</span>
        <span class="detail-cell">${rillSlotRate}/slot/hr</span>
        <span class="detail-cell">~${rillMonthlyCost.toLocaleString()}</span>
        <span class="detail-cell">
          <span class="source-badge source-user">User-configured</span>
        </span>
      </div>

      <div class="detail-row detail-row-total">
        <span class="detail-cell detail-cell-name font-semibold">Total</span>
        <span class="detail-cell font-semibold">{provisionedSlots}</span>
        <span class="detail-cell"></span>
        <span class="detail-cell font-semibold">~${totalMonthlyCost.toLocaleString()}</span>
        <span class="detail-cell"></span>
      </div>
    </div>
  </section>

  <!-- Analytics placeholder -->
  <section class="detail-section">
    <h3 class="section-title">Slot Usage Over Time</h3>
    <div class="analytics-placeholder">
      <div class="placeholder-chart">
        <div class="placeholder-bar-group">
          {#each Array(12) as _, i}
            <div
              class="placeholder-bar"
              style:height="{20 + Math.sin(i * 0.8) * 15 + Math.random() * 10}%"
            />
          {/each}
        </div>
        <div class="placeholder-axis" />
      </div>
      <p class="placeholder-text">
        Slot usage analytics coming soon. This will show provisioned slots, cluster slots, and rill slots over time.
      </p>
    </div>
  </section>

  <section class="detail-section">
    <h3 class="section-title">Estimated Cost Over Time</h3>
    <div class="analytics-placeholder">
      <div class="placeholder-chart">
        <div class="placeholder-bar-group">
          {#each Array(12) as _, i}
            <div
              class="placeholder-bar placeholder-bar-cost"
              style:height="{25 + Math.cos(i * 0.6) * 12 + Math.random() * 8}%"
            />
          {/each}
        </div>
        <div class="placeholder-axis" />
      </div>
      <p class="placeholder-text">
        Cost analytics coming soon. This will show cluster slot costs and rill slot costs broken down over time.
      </p>
    </div>
  </section>
</div>

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
  .page {
    @apply flex flex-col gap-6 w-full;
  }
  .page-header {
    @apply flex items-center justify-between;
  }
  .page-title {
    @apply text-lg font-semibold text-fg-primary;
  }
  .status-badge {
    @apply text-xs px-2 py-0.5 rounded-full font-medium;
  }
  .status-badge.bg-green-500 {
    @apply bg-green-100 text-green-700;
  }
  .status-badge.bg-yellow-500 {
    @apply bg-yellow-100 text-yellow-700;
  }
  .status-badge.bg-red-500 {
    @apply bg-red-100 text-red-700;
  }
  .status-badge.bg-gray-400 {
    @apply bg-gray-100 text-gray-600;
  }
  .manage-btn {
    @apply text-sm text-white bg-primary-500 border-none rounded-md px-3 py-1.5 cursor-pointer font-medium;
  }
  .manage-btn:hover {
    @apply bg-primary-600;
  }

  /* Slot summary cards */
  .slot-cards {
    @apply grid gap-4;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  }
  .slot-card {
    @apply flex flex-col gap-1 border border-border rounded-lg p-4;
  }
  .slot-card-header {
    @apply flex items-center justify-between;
  }
  .slot-card-label {
    @apply text-xs text-fg-secondary uppercase tracking-wide font-medium;
  }
  .slot-card-value {
    @apply text-2xl font-semibold text-fg-primary tabular-nums;
  }
  .slot-card-sub {
    @apply text-xs text-fg-tertiary;
  }

  /* Refresh button */
  .refresh-btn {
    @apply p-1 text-fg-tertiary bg-transparent border-none cursor-pointer rounded;
  }
  .refresh-btn:hover {
    @apply text-fg-secondary bg-surface-subtle;
  }
  .refresh-btn:disabled {
    @apply cursor-not-allowed;
  }
  .refresh-icon {
    @apply w-3 h-3;
  }
  .refreshing .refresh-icon {
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  /* Detail sections */
  .detail-section {
    @apply border border-border rounded-lg p-5;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide mb-4;
  }

  /* Detail table */
  .detail-table {
    @apply border border-border rounded-md overflow-hidden;
  }
  .detail-header {
    @apply flex bg-surface-subtle text-xs font-semibold text-fg-secondary uppercase tracking-wide;
  }
  .detail-header .detail-cell {
    @apply px-3 py-2;
  }
  .detail-row {
    @apply flex text-sm border-t border-border;
  }
  .detail-row .detail-cell {
    @apply px-3 py-2.5;
  }
  .detail-row-total {
    @apply bg-surface-subtle;
  }
  .detail-cell {
    @apply flex-1 flex items-center;
  }
  .detail-cell-name {
    @apply flex-[1.5];
  }

  /* Source badges */
  .source-badge {
    @apply text-[10px] px-1.5 py-0.5 rounded-full leading-none font-medium;
  }
  .source-auto {
    @apply text-green-700 bg-green-100;
  }
  .source-user {
    @apply text-primary-600 bg-primary-100;
  }

  /* Analytics placeholder */
  .analytics-placeholder {
    @apply flex flex-col items-center gap-4 py-6;
  }
  .placeholder-chart {
    @apply w-full max-w-md flex flex-col items-stretch;
  }
  .placeholder-bar-group {
    @apply flex items-end justify-between gap-1.5 h-24;
  }
  .placeholder-bar {
    @apply flex-1 bg-primary-100 rounded-t-sm;
    min-height: 8px;
  }
  .placeholder-bar-cost {
    @apply bg-green-100;
  }
  .placeholder-axis {
    @apply h-px bg-border mt-1;
  }
  .placeholder-text {
    @apply text-sm text-fg-tertiary text-center max-w-sm;
  }
</style>
