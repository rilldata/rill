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
  import { useProjectDeployment } from "../selectors";
  import { getStatusDotClass, getStatusLabel } from "../display-utils";
  import {
    SLOT_RATE_PER_HR,
    HOURS_PER_MONTH,
    SLOT_TIERS,
  } from "../overview/slots-utils";
  import ManageSlotsModal from "../overview/ManageSlotsModal.svelte";

  let {
    organization,
    project,
  }: {
    organization: string;
    project: string;
  } = $props();

  // Deployment
  let projectDeployment = $derived(useProjectDeployment(organization, project));
  let deployment = $derived($projectDeployment.data);
  let deploymentStatus = $derived(
    deployment?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
  );

  // Project
  let proj = $derived(createAdminServiceGetProject(organization, project));
  let projectData = $derived($proj.data?.project);

  // Slots
  let currentSlots = $derived(Number(projectData?.prodSlots) || 0);
  let canManage = $derived(
    $proj.data?.projectPermissions?.manageProject ?? false,
  );
  // Self-managed: any non-DuckDB OLAP connector (ClickHouse, MotherDuck, Druid, Pinot, StarRocks)
  let olapType = $derived((projectData as any)?.olapConnector ?? "");
  let isRillManaged = $derived(olapType === "" || olapType === "duckdb");
  let prodModalOpen = $state(false);

  // Billing
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isTrial = $derived(isTrialPlan(planName));
  let isEnterprise = $derived(planName !== "" && isEnterprisePlan(planName));

  // Slot types
  let prodSlots = $derived(currentSlots);
  let devSlots = $derived(2); // TODO: wire to project data when dev slots are available
  let totalSlots = $derived(prodSlots + devSlots);

  // Cluster info
  let prodTier = $derived(SLOT_TIERS.find((t) => t.slots === prodSlots));
  let prodClusterLabel = $derived(
    prodTier?.instance ?? `${prodSlots * 4}GiB / ${prodSlots}vCPU`,
  );
  let devClusterLabel = $derived(
    devSlots > 0 ? `${devSlots * 4}GiB / ${devSlots}vCPU` : "\u2014",
  );
  let prodMonthlyCost = $derived(
    Math.round(prodSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  );
  let devMonthlyCost = $derived(
    Math.round(devSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  );
  let totalMonthlyCost = $derived(
    Math.round(totalSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  );

  // Bar percentages
  let prodPct = $derived(totalSlots > 0 ? (prodSlots / totalSlots) * 100 : 50);
  let devPct = $derived(totalSlots > 0 ? (devSlots / totalSlots) * 100 : 50);
</script>

{#if !isEnterprise}
  <div class="page">
    <!-- Page header -->
    <div class="page-header">
      <div class="flex items-center gap-3">
        <h2 class="page-title">Deployments</h2>
        <span class="status-badge {getStatusDotClass(deploymentStatus)}">
          {getStatusLabel(deploymentStatus)}
        </span>
      </div>
      <div class="page-header-right">
        <span class="total-cost">
          {totalSlots}
          {totalSlots === 1 ? "slot" : "slots"} total &middot; ~${totalMonthlyCost.toLocaleString()}/mo
        </span>
        <a href="/{organization}/-/settings/usage" class="pricing-link">
          See price breakdown
        </a>
      </div>
    </div>

    <!-- Slot allocation bar -->
    <div class="alloc-bar-container">
      <div class="alloc-bar">
        {#if prodSlots > 0}
          <div
            class="alloc-segment alloc-segment-prod"
            style="width: {prodPct}%"
          >
            <span class="alloc-segment-text">
              Prod &middot; {prodSlots}
            </span>
          </div>
        {/if}
        {#if devSlots > 0}
          <div class="alloc-segment alloc-segment-dev" style="width: {devPct}%">
            <span class="alloc-segment-text">
              Dev &middot; {devSlots}
            </span>
          </div>
        {:else}
          <div
            class="alloc-segment alloc-segment-dev-empty"
            style="width: {devPct > 0 ? devPct : 25}%"
          >
            <span class="alloc-segment-text alloc-segment-text-muted">
              Dev &middot; 0
            </span>
          </div>
        {/if}
      </div>
    </div>

    <!-- Two-section grid -->
    <div class="section-grid">
      <!-- Production -->
      <div class="section-card">
        <div class="section-header">
          <div class="section-title-row">
            <span class="section-dot section-dot-prod"></span>
            <h3 class="section-title">Production</h3>
          </div>
          {#if canManage && !$subscriptionQuery?.isLoading}
            <button
              class="section-manage-btn"
              onclick={() => (prodModalOpen = true)}
            >
              Manage
            </button>
          {/if}
        </div>
        <div class="section-body">
          <div class="section-metric">
            <span class="metric-value">{prodClusterLabel}</span>
            <span class="metric-label">Cluster Size</span>
          </div>
          <div class="section-details">
            <div class="detail-row">
              <span class="detail-label">Slots</span>
              <span class="detail-value">{prodSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value"
                >~${prodMonthlyCost.toLocaleString()}/mo</span
              >
            </div>
            <div class="detail-row">
              <span class="detail-label">Rate</span>
              <span class="detail-value">${SLOT_RATE_PER_HR}/slot/hr</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Development -->
      <div class="section-card section-card-dev">
        <div class="section-header">
          <div class="section-title-row">
            <span class="section-dot section-dot-dev"></span>
            <h3 class="section-title">Development</h3>
          </div>
          {#if canManage && !$subscriptionQuery?.isLoading}
            <!-- TODO: re-add on:click={() => (devModalOpen = true)} when dev slots are available -->
            <button
              class="section-manage-btn section-manage-btn-dev"
              disabled
              title="Dev slots coming soon"
            >
              Manage
            </button>
          {/if}
        </div>
        <div class="section-body">
          <div class="section-metric">
            <span class="metric-value" class:metric-value-empty={devSlots === 0}
              >{devClusterLabel}</span
            >
            <span class="metric-label">Cluster Size</span>
          </div>
          <div class="section-details">
            <div class="detail-row">
              <span class="detail-label">Slots</span>
              <span class="detail-value">{devSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value">
                {devSlots > 0 ? `~$${devMonthlyCost.toLocaleString()}/mo` : "—"}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Rate</span>
              <span class="detail-value">${SLOT_RATE_PER_HR}/slot/hr</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <ManageSlotsModal
    bind:open={prodModalOpen}
    {organization}
    {project}
    {currentSlots}
    {isRillManaged}
    {isTrial}
  />
{/if}

<style lang="postcss">
  .page {
    @apply flex flex-col gap-6 w-full;
  }

  /* Page header */
  .page-header {
    @apply flex items-center justify-between;
  }
  .page-title {
    @apply text-lg font-semibold text-fg-primary;
  }
  .page-header-right {
    @apply flex items-center gap-3;
  }
  .total-cost {
    @apply text-sm text-fg-secondary tabular-nums;
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
  .pricing-link {
    @apply text-xs text-primary-500 no-underline;
  }
  .pricing-link:hover {
    @apply text-primary-600 underline;
  }

  /* Allocation bar */
  .alloc-bar-container {
    @apply flex flex-col gap-2.5;
  }
  .alloc-bar {
    @apply flex w-full h-10 rounded-lg overflow-hidden gap-0.5;
  }
  .alloc-segment {
    @apply flex items-center justify-center transition-all duration-300;
    min-width: 48px;
  }
  .alloc-segment-prod {
    @apply bg-primary-500 rounded-l-lg;
  }
  .alloc-segment-dev {
    @apply bg-amber-400 rounded-r-lg;
  }
  .alloc-segment-dev-empty {
    @apply rounded-r-lg border border-dashed border-border bg-surface-subtle;
  }
  .alloc-segment-text {
    @apply text-xs font-semibold text-white whitespace-nowrap;
  }
  .alloc-segment-text-muted {
    @apply text-fg-tertiary;
  }

  /* Section grid */
  .section-grid {
    @apply grid gap-5;
    grid-template-columns: repeat(2, 1fr);
  }
  .section-card {
    @apply border border-border rounded-xl p-5 flex flex-col gap-4;
  }
  .section-card-dev {
    @apply opacity-75;
  }
  .section-header {
    @apply flex items-center justify-between;
  }
  .section-title-row {
    @apply flex items-center gap-2.5;
  }
  .section-dot {
    @apply w-3 h-3 rounded-full shrink-0;
  }
  .section-dot-prod {
    @apply bg-primary-500;
  }
  .section-dot-dev {
    @apply bg-amber-400;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .section-manage-btn {
    @apply text-xs font-medium text-primary-500 bg-transparent border border-primary-300 rounded-md px-3 py-1 cursor-pointer;
  }
  .section-manage-btn:hover {
    @apply bg-primary-50 text-primary-600 border-primary-400;
  }
  .section-manage-btn:disabled {
    @apply opacity-40 cursor-not-allowed;
  }
  .section-manage-btn:disabled:hover {
    @apply bg-transparent text-primary-500 border-primary-300;
  }

  /* Section body */
  .section-body {
    @apply flex flex-col gap-4;
  }
  .section-metric {
    @apply flex flex-col gap-1;
  }
  .metric-value {
    @apply text-2xl font-semibold text-fg-primary tabular-nums tracking-tight;
  }
  .metric-value-empty {
    @apply text-fg-tertiary;
  }
  .metric-label {
    @apply text-xs text-fg-tertiary uppercase tracking-wide;
  }
  .section-details {
    @apply flex flex-col gap-1.5 border-t border-border pt-3;
  }
  .detail-row {
    @apply flex items-center justify-between;
  }
  .detail-label {
    @apply text-xs text-fg-tertiary;
  }
  .detail-value {
    @apply text-sm text-fg-primary tabular-nums;
  }
  .detail-value-muted {
    @apply text-fg-tertiary italic;
  }
</style>
