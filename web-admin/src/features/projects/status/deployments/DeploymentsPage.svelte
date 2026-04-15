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
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";

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
    return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
  }

  // Slot types
  let prodSlots = $derived(currentSlots);
  let devSlots = $derived(1); // TODO: wire to project data when dev slots are available
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
    <h2 class="page-title">Usage & Compute</h2>

    <!-- Summary bar -->
    <div class="summary-bar">
      <div class="summary-panel">
        <span class="summary-label inline-flex items-center gap-1">
          Total Compute Units
          <Tooltip location="right" alignment="middle" distance={8}>
            <span class="text-fg-muted flex">
              <InfoCircle size="13px" />
            </span>
            <TooltipContent maxWidth="200px" slot="tooltip-content">
              1 unit = 4 GiB RAM / 1 vCPU
            </TooltipContent>
          </Tooltip>
        </span>
        <span class="summary-value">{totalSlots}</span>
        <span class="summary-breakdown">
          {prodSlots} production
          <span class="text-fg-tertiary">&middot;</span>
          {devSlots} development
        </span>
      </div>

      <div class="summary-divider"></div>

      <div class="summary-panel">
        <span class="summary-label">Est. monthly project cost</span>
        <span class="summary-value text-fg-secondary">${totalMonthlyCost}</span>
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

    <!-- Unit Breakdown -->
    <div class="section-heading">
      <h3 class="section-heading-text">Unit breakdown</h3>
    </div>

    <div class="section-grid">
      <!-- Production card -->
      <div class="breakdown-card breakdown-card-prod">
        <div class="breakdown-header">
          <h4 class="breakdown-title">Production</h4>
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <div class="metric-row">
              <span class="metric-value">
                {prodMemory}<span class="metric-unit">GiB</span>
                <span class="metric-slash">/</span>
                {prodCpu}<span class="metric-unit">vCPU</span>
              </span>
              {#if canManage && !$subscriptionQuery?.isLoading}
                <button
                  class="manage-btn"
                  onclick={() => (prodModalOpen = true)}
                >
                  Manage units
                </button>
              {/if}
            </div>
            <span class="metric-label">Cluster size</span>
          </div>
          <div class="breakdown-details">
            <div class="detail-row">
              <span class="detail-label">Units</span>
              <span class="detail-value">{prodSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value-sm">~${prodMonthlyCost}/mo</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Development card -->
      <div class="breakdown-card breakdown-card-dev">
        <div class="breakdown-header">
          <h4 class="breakdown-title">Development</h4>
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <div class="metric-row">
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
            </div>
            <span class="metric-label">Cluster size</span>
          </div>
          <div class="breakdown-details">
            <div class="detail-row">
              <span class="detail-label">Units</span>
              <span class="detail-value">{devSlots}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Est. cost</span>
              <span class="detail-value-sm">
                {devSlots > 0 ? `~$${devMonthlyCost}/mo` : "—"}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Development Details table -->
    <div class="section-heading">
      <h3 class="section-heading-text">Development details</h3>
    </div>

    <div class="dev-table-container">
      <table class="dev-table">
        <thead>
          <tr>
            <th>Branch</th>
            <th>Author</th>
            <th>Status</th>
            <th>Units</th>
            <th>Last updated</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#if $deploymentsQuery.isLoading}
            <tr>
              <td colspan="6" class="table-empty">Loading...</td>
            </tr>
          {:else if devDeployments.length === 0}
            <tr>
              <td colspan="6" class="table-empty">No development deployments</td
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
                <td>
                  <button class="overflow-btn">···</button>
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
    title="Manage Prod Cluster Size"
    {organization}
    {project}
    {currentSlots}
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
    @apply flex rounded-xl overflow-hidden bg-surface-background border border-border;
  }
  .summary-panel {
    @apply flex-1 flex flex-col gap-2 p-5;
  }
  .summary-divider {
    @apply w-px bg-border my-4;
  }
  .summary-label {
    @apply font-sans text-xs font-semibold leading-none text-fg-tertiary;
  }
  .summary-value {
    @apply text-3xl font-bold text-fg-primary tabular-nums tracking-tight;
  }
  .summary-breakdown {
    @apply text-sm flex items-center gap-1.5;
  }
  .summary-breakdown-plain {
    @apply font-medium text-fg-muted;
    font-size: 12px;
    line-height: 18px;
  }
  .summary-cycle {
    @apply text-xs text-fg-tertiary;
  }
  /* Section headings */
  .section-heading {
    @apply mt-2;
  }
  .section-heading-text {
    @apply font-sans text-xs font-semibold leading-none text-fg-tertiary;
  }

  /* Breakdown cards */
  .section-grid {
    @apply grid gap-5;
    grid-template-columns: repeat(2, 1fr);
  }
  .breakdown-card {
    @apply border border-border rounded-xl bg-surface-background flex flex-col;
    padding-top: 24px;
    padding-bottom: 24px;
    gap: 12px;
  }
  .breakdown-header {
    @apply flex items-center justify-between px-6;
  }
  .breakdown-title {
    @apply font-sans text-base font-semibold leading-none;
  }
  .manage-btn {
    @apply text-xs font-medium text-primary-500 bg-transparent border border-primary-300 rounded-none px-3 py-1 cursor-pointer;
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
  .metric-row {
    @apply flex items-center justify-between;
  }
  .metric-value {
    @apply font-sans text-3xl font-semibold text-fg-primary tabular-nums;
    vertical-align: baseline;
  }
  .metric-unit {
    @apply font-sans text-lg font-bold text-fg-secondary leading-none;
    vertical-align: baseline;
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
    @apply text-sm font-medium text-fg-primary;
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
    @apply border border-border rounded-xl overflow-hidden bg-surface-background;
  }
  .dev-table {
    @apply w-full text-sm;
  }
  .dev-table thead {
    @apply bg-surface-subtle;
  }
  .dev-table th {
    @apply text-left text-xs font-semibold text-fg-secondary tracking-wide px-4 py-2.5;
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
  .overflow-btn {
    @apply bg-transparent border-none text-fg-tertiary cursor-pointer text-base px-1;
  }
  .overflow-btn:hover {
    @apply text-fg-primary;
  }
</style>
