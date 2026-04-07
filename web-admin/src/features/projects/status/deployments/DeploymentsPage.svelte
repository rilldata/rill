<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
    createAdminServiceListDeployments,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
  import { getStatusDotClass, getStatusLabel } from "../display-utils";
  import { SLOT_RATE_PER_HR, HOURS_PER_MONTH } from "../overview/slots-utils";
  import ManageSlotsModal from "../overview/ManageSlotsModal.svelte";

  let {
    organization,
    project,
  }: {
    organization: string;
    project: string;
  } = $props();

  // Project
  let proj = $derived(createAdminServiceGetProject(organization, project));
  let projectData = $derived($proj.data?.project);

  // Slots
  let currentSlots = $derived(Number(projectData?.prodSlots) || 0);
  let canManage = $derived(
    $proj.data?.projectPermissions?.manageProject ?? false,
  );
  let prodModalOpen = $state(false);
  let devModalOpen = $state(false);

  // Billing
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isEnterprise = $derived(planName !== "" && isEnterprisePlan(planName));

  // Billing cycle dates
  let cycleStart = $derived(
    $subscriptionQuery?.data?.subscription?.currentBillingCycleStartDate,
  );
  let cycleEnd = $derived(
    $subscriptionQuery?.data?.subscription?.currentBillingCycleEndDate,
  );

  function formatCycleDate(dateStr: string | undefined): string {
    if (!dateStr) return "";
    const d = new Date(dateStr);
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  }

  // Slot types
  let prodSlots = $derived(currentSlots);
  let devSlots = $derived(2); // TODO: wire to project data when dev slots are available
  let totalSlots = $derived(prodSlots + devSlots);

  // Cluster info (split into number + unit for display)
  let prodMemory = $derived(prodSlots * 4);
  let prodCpu = $derived(prodSlots);
  let devMemory = $derived(devSlots * 4);
  let devCpu = $derived(devSlots);

  // Cost calculations (with decimals)
  let prodMonthlyCost = $derived(
    (prodSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH).toFixed(2),
  );
  let devMonthlyCost = $derived(
    (devSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH).toFixed(2),
  );
  let totalMonthlyCost = $derived(
    (totalSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH).toFixed(2),
  );

  // Bar percentages
  let prodPct = $derived(totalSlots > 0 ? (prodSlots / totalSlots) * 100 : 50);
  let devPct = $derived(totalSlots > 0 ? (devSlots / totalSlots) * 100 : 50);

  // Dev deployments table
  let deploymentsQuery = $derived(
    createAdminServiceListDeployments(organization, project),
  );
  let primaryBranch = $derived(projectData?.primaryBranch ?? "main");
  let devDeployments = $derived(
    ($deploymentsQuery.data?.deployments ?? []).filter(
      (d) => d.branch && d.branch !== primaryBranch,
    ),
  );
</script>

{#if !isEnterprise}
  <div class="page">
    <!-- Page header -->
    <h2 class="page-title">Usage & Slots</h2>

    <!-- Summary bar -->
    <div class="summary-bar">
      <div class="summary-panel">
        <span class="summary-label">TOTAL SLOTS</span>
        <span class="summary-value">{totalSlots}</span>
        <span class="summary-breakdown">
          <span class="text-prod">{prodSlots} production</span>
          <span class="text-fg-tertiary">&middot;</span>
          <span class="text-dev">{devSlots} development</span>
        </span>
        <!-- Mini segmented bar -->
        <div class="mini-bar">
          {#if prodSlots > 0}
            <div
              class="mini-segment mini-segment-prod"
              style="width: {prodPct}%"
            ></div>
          {/if}
          {#if devSlots > 0}
            <div
              class="mini-segment mini-segment-dev"
              style="width: {devPct}%"
            ></div>
          {/if}
        </div>
      </div>

      <div class="summary-divider"></div>

      <div class="summary-panel">
        <span class="summary-label">EST. MONTHLY COST</span>
        <span class="summary-value">${totalMonthlyCost}</span>
        <span class="summary-breakdown-plain">
          ${prodMonthlyCost} prod + ${devMonthlyCost} dev
        </span>
        {#if cycleStart || cycleEnd}
          <span class="summary-cycle">
            Billing cycle: {formatCycleDate(cycleStart)} – {formatCycleDate(
              cycleEnd,
            )}
          </span>
        {/if}
      </div>
    </div>

    <!-- Slots Breakdown -->
    <div class="section-heading">
      <h3 class="section-heading-text">SLOTS BREAKDOWN</h3>
    </div>

    <div class="section-grid">
      <!-- Production card -->
      <div class="breakdown-card breakdown-card-prod">
        <div class="breakdown-header">
          <div class="breakdown-title-row">
            <h4 class="breakdown-title">Production</h4>
          </div>
          {#if canManage && !$subscriptionQuery?.isLoading}
            <button class="manage-btn" onclick={() => (prodModalOpen = true)}>
              Manage
            </button>
          {/if}
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <span class="metric-value">
              {prodMemory}<span class="metric-unit">GiB</span>
              <span class="metric-slash">/</span>
              {prodCpu}<span class="metric-unit">vCPU</span>
            </span>
            <span class="metric-label">Cluster size</span>
          </div>
          <div class="breakdown-details">
            <div class="detail-row">
              <span class="detail-label">Slots</span>
              <span class="detail-value text-prod">{prodSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value-sm">~${prodMonthlyCost}/mo</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Rate</span>
              <span class="detail-value-sm">${SLOT_RATE_PER_HR}/slot/hr</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Development card -->
      <div class="breakdown-card breakdown-card-dev">
        <div class="breakdown-header">
          <div class="breakdown-title-row">
            <h4 class="breakdown-title">Development</h4>
          </div>
          {#if canManage && !$subscriptionQuery?.isLoading}
            <button class="manage-btn" onclick={() => (devModalOpen = true)}>
              Manage
            </button>
          {/if}
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <span
              class="metric-value"
              class:metric-value-empty={devSlots === 0}
            >
              {#if devSlots > 0}
                {devMemory}<span class="metric-unit">GiB</span>
                <span class="metric-slash">/</span>
                {devCpu}<span class="metric-unit">vCPU</span>
              {:else}
                —
              {/if}
            </span>
            <span class="metric-label">Cluster size</span>
          </div>
          <div class="breakdown-details">
            <div class="detail-row">
              <span class="detail-label">Slots</span>
              <span class="detail-value text-dev">{devSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value-sm">
                {devSlots > 0 ? `~$${devMonthlyCost}/mo` : "—"}
              </span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Rate</span>
              <span class="detail-value-sm">${SLOT_RATE_PER_HR}/slot/hr</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Development Details table -->
    <div class="section-heading">
      <h3 class="section-heading-text">DEVELOPMENT DETAILS</h3>
    </div>

    <div class="dev-table-container">
      <table class="dev-table">
        <thead>
          <tr>
            <th>Branch</th>
            <th>Author</th>
            <th>Status</th>
            <th>Slots</th>
            <th>Last updated</th>
          </tr>
        </thead>
        <tbody>
          {#if $deploymentsQuery.isLoading}
            <tr>
              <td colspan="5" class="table-empty">Loading...</td>
            </tr>
          {:else if devDeployments.length === 0}
            <tr>
              <td colspan="5" class="table-empty">No development deployments</td
              >
            </tr>
          {:else}
            {#each devDeployments as dep (dep.id)}
              <tr>
                <td class="branch-cell">
                  <span class="branch-name">{dep.branch ?? "—"}</span>
                </td>
                <td class="text-fg-secondary">{dep.ownerUserId ?? "—"}</td>
                <td>
                  <span
                    class="status-badge {getStatusDotClass(
                      dep.status ??
                        V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
                    )}"
                  >
                    {getStatusLabel(
                      dep.status ??
                        V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED,
                    )}
                  </span>
                </td>
                <td class="tabular-nums">{devSlots}</td>
                <td class="text-fg-secondary">
                  {#if dep.updatedOn}
                    {new Date(dep.updatedOn).toLocaleDateString(undefined, {
                      month: "short",
                      day: "numeric",
                      hour: "numeric",
                      minute: "numeric",
                    })}
                  {:else}
                    —
                  {/if}
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
    </div>
  </div>

  <ManageSlotsModal
    bind:open={prodModalOpen}
    title="Prod Cluster Size"
    {organization}
    {project}
    {currentSlots}
  />
  <ManageSlotsModal
    bind:open={devModalOpen}
    title="Dev Cluster Size"
    {organization}
    {project}
    currentSlots={devSlots}
  />
{/if}

<style lang="postcss">
  .page {
    @apply flex flex-col gap-6 w-full;
  }

  /* Page header */
  .page-title {
    @apply text-lg font-semibold text-fg-primary;
  }

  /* Summary bar */
  .summary-bar {
    @apply flex rounded-xl overflow-hidden bg-surface-subtle border border-border;
  }
  .summary-panel {
    @apply flex-1 flex flex-col gap-2 p-5;
  }
  .summary-divider {
    @apply w-px bg-border my-4;
  }
  .summary-label {
    @apply text-[11px] font-semibold tracking-widest text-fg-tertiary uppercase;
  }
  .summary-value {
    @apply text-3xl font-bold text-fg-primary tabular-nums tracking-tight;
  }
  .summary-breakdown {
    @apply text-sm flex items-center gap-1.5;
  }
  .summary-breakdown-plain {
    @apply text-sm text-fg-secondary;
  }
  .summary-cycle {
    @apply text-xs text-fg-tertiary;
  }
  .text-prod {
    color: #8b5cf6;
  }
  .text-dev {
    color: #65a30d;
  }

  /* Mini bar */
  .mini-bar {
    @apply flex w-full h-1.5 rounded-full overflow-hidden gap-0.5 mt-1;
    @apply bg-surface-subtle;
  }
  .mini-segment {
    @apply h-full transition-all duration-300;
    min-width: 8px;
  }
  .mini-segment-prod {
    background: #8b5cf6;
    border-radius: 9999px 0 0 9999px;
  }
  .mini-segment-dev {
    background: #65a30d;
    border-radius: 0 9999px 9999px 0;
  }

  /* Section headings */
  .section-heading {
    @apply mt-2;
  }
  .section-heading-text {
    @apply text-[11px] font-semibold tracking-widest text-fg-tertiary uppercase;
  }

  /* Breakdown cards */
  .section-grid {
    @apply grid gap-5;
    grid-template-columns: repeat(2, 1fr);
  }
  .breakdown-card {
    @apply border border-border rounded-xl bg-surface-subtle flex flex-col;
    padding-top: 24px;
    padding-bottom: 24px;
    gap: 12px;
  }
  .breakdown-header {
    @apply flex items-center justify-between px-6;
  }
  .breakdown-title-row {
    @apply flex items-center;
    gap: 6px;
  }

  .breakdown-title {
    @apply font-semibold;
    font-size: 16px;
    line-height: 16px;
  }
  .breakdown-card-prod .breakdown-title {
    color: #8b5cf6;
  }
  .breakdown-card-dev .breakdown-title {
    color: #65a30d;
  }
  .manage-btn {
    @apply text-xs font-medium text-primary-500 bg-transparent border border-primary-300 rounded-md px-3 py-1 cursor-pointer;
  }
  .manage-btn:hover {
    @apply bg-primary-50 text-primary-600 border-primary-400;
  }

  /* Breakdown body */
  .breakdown-body {
    @apply flex flex-col gap-4;
  }
  .breakdown-metric {
    @apply flex flex-col gap-1 px-6;
  }
  .metric-value {
    @apply text-2xl font-semibold text-fg-primary tabular-nums tracking-tight;
  }
  .metric-unit {
    @apply text-sm font-medium text-fg-tertiary;
    vertical-align: baseline;
    margin-left: 2px;
  }
  .metric-slash {
    @apply text-fg-tertiary mx-1;
  }
  .metric-value-empty {
    @apply text-fg-tertiary;
  }
  .metric-label {
    @apply text-xs text-fg-tertiary;
  }
  .breakdown-details {
    @apply flex flex-col gap-1.5 border-t border-border pt-3;
    min-width: 356px;
    padding-left: 24px;
    padding-right: 24px;
  }
  .detail-row {
    @apply flex items-center justify-between border-b border-border p-2;
    min-width: 85px;
    gap: 8px;
  }
  .detail-row:last-child {
    @apply border-b-0;
  }
  .detail-label {
    @apply text-sm font-medium text-fg-tertiary;
  }
  .detail-value {
    @apply text-xl font-extrabold text-fg-primary tabular-nums text-right;
    line-height: 100%;
  }
  .detail-value-sm {
    @apply text-sm font-semibold text-fg-primary tabular-nums text-right;
  }

  /* Dev deployments table */
  .dev-table-container {
    @apply border border-border rounded-xl overflow-hidden bg-surface-subtle;
  }
  .dev-table {
    @apply w-full text-sm;
  }
  .dev-table thead {
    @apply bg-surface-subtle;
  }
  .dev-table th {
    @apply text-left text-xs font-semibold text-fg-secondary uppercase tracking-wide px-4 py-2.5;
  }
  .dev-table td {
    @apply px-4 py-3 text-sm text-fg-primary border-t border-border;
  }
  .table-empty {
    @apply text-center text-fg-tertiary py-8;
  }
  .branch-cell {
    @apply font-medium;
  }
  .branch-name {
    @apply font-mono text-xs;
  }

  /* Status badges */
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
</style>
