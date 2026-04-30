<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetOrganization,
    createAdminServiceGetBillingSubscription,
    createAdminServiceListDeployments,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isFreePlan,
    isProPlan,
    isManagedPlan,
    isTeamPlan,
    isEnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { isProdDeployment } from "./deployment-utils";
  import { V1DeploymentStatus } from "@rilldata/web-admin/client";
  import { isTransitoryStatus } from "@rilldata/web-admin/features/projects/status/display-utils";
  import { SLOT_RATE_PER_HR } from "@rilldata/web-admin/features/projects/status/overview/slots-utils";
  import ManageSlotsModal from "@rilldata/web-admin/features/projects/status/overview/ManageSlotsModal.svelte";
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

  // Billing
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let showDeploymentSection = $derived(
    planName !== "" &&
      (isTrialPlan(planName) ||
        isFreePlan(planName) ||
        isProPlan(planName) ||
        isManagedPlan(planName) ||
        isTeamPlan(planName) ||
        isEnterprisePlan(planName)),
  );

  // Plans without per-unit billing visibility — Trial and Team get a flat
  // allowance, so we hide cost estimates and surface the unit cap instead.
  let hasCostVisibility = $derived(
    !(isTrialPlan(planName) || isTeamPlan(planName)),
  );

  // Slot quotas come from the organization (set via plan defaults or sudo
  // overrides). Negative or missing values mean no limit.
  function quotaCap(value: number | undefined): number | undefined {
    if (value === undefined || value < 0) return undefined;
    return value;
  }
  let orgQuery = $derived(createAdminServiceGetOrganization(organization));
  let orgQuotas = $derived($orgQuery.data?.organization?.quotas);
  let maxSlotsPerDeployment = $derived(quotaCap(orgQuotas?.slotsPerDeployment));
  let planTotalSlotsCap = $derived(quotaCap(orgQuotas?.slotsTotal));

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
    return d.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  // Deployments — used to count active dev deployments for total compute
  let deploymentsQuery = $derived(
    createAdminServiceListDeployments(
      organization,
      project,
      {},
      {
        query: {
          refetchInterval: (query) => {
            const deployments = query.state.data?.deployments;
            if (deployments?.some((d) => isTransitoryStatus(d.status!))) {
              return 2000;
            }
            return false;
          },
        },
      },
    ),
  );
  // Only deployments that have actually allocated compute count toward the
  // running total. Pending deployments haven't reserved units yet.
  let activeDevDeploymentCount = $derived(
    ($deploymentsQuery.data?.deployments ?? []).filter(
      (d: V1Deployment) =>
        !isProdDeployment(d) &&
        (d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
          d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING),
    ).length,
  );

  // Slot types
  let prodSlots = $derived(parseInt(projectData?.prodSlots ?? "0", 10) || 0);
  let devSlotsPerDeployment = $derived(
    parseInt(projectData?.devSlots ?? "0", 10) || 0,
  );
  let runningDevSlots = $derived(
    devSlotsPerDeployment * activeDevDeploymentCount,
  );
  let totalSlots = $derived(prodSlots + runningDevSlots);

  // Per-modal max — clamps tier selection so the new total stays within
  // planTotalSlotsCap. When the cap is unlimited or the dev modal has no
  // active deployments, only the per-deployment cap applies.
  function clampMax(
    perDeploymentCap: number | undefined,
    totalRemaining: number | undefined,
  ): number | undefined {
    if (perDeploymentCap === undefined) return totalRemaining;
    if (totalRemaining === undefined) return perDeploymentCap;
    return Math.max(0, Math.min(perDeploymentCap, totalRemaining));
  }
  let prodMaxSlots = $derived(
    clampMax(
      maxSlotsPerDeployment,
      planTotalSlotsCap !== undefined
        ? planTotalSlotsCap - runningDevSlots
        : undefined,
    ),
  );
  let devMaxSlots = $derived(
    clampMax(
      maxSlotsPerDeployment,
      planTotalSlotsCap !== undefined && activeDevDeploymentCount > 0
        ? Math.floor((planTotalSlotsCap - prodSlots) / activeDevDeploymentCount)
        : undefined,
    ),
  );
  let exceedsTotalCap = $derived(
    planTotalSlotsCap !== undefined && totalSlots > planTotalSlotsCap,
  );

  // Cluster info (split into number + unit for display)
  let prodMemory = $derived(prodSlots * 4);
  let prodCpu = $derived(prodSlots);
  let devMemory = $derived(devSlotsPerDeployment * 4);
  let devCpu = $derived(devSlotsPerDeployment);

  // Cost calculations
  let prodHourlyCost = $derived((prodSlots * SLOT_RATE_PER_HR).toFixed(2));
  let runningDevHourlyCost = $derived(
    (runningDevSlots * SLOT_RATE_PER_HR).toFixed(2),
  );
  let totalHourlyCost = $derived((totalSlots * SLOT_RATE_PER_HR).toFixed(2));

  // Manage units modals
  let prodSlotsModalOpen = $state(false);
  let devSlotsModalOpen = $state(false);
</script>

{#if showDeploymentSection}
  <div class="page">
    <!-- Page header -->
    <h2 class="page-title">Deployments</h2>

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
          {runningDevSlots} development
        </span>
        {#if planTotalSlotsCap !== undefined}
          <span class="summary-cycle" class:summary-warning={exceedsTotalCap}>
            Plan limit: {totalSlots} of {planTotalSlotsCap} units used
            {#if exceedsTotalCap}
              · over plan cap
            {/if}
          </span>
        {/if}
      </div>

      {#if hasCostVisibility}
        <div class="summary-divider"></div>

        <div class="summary-panel">
          <span class="summary-label">Est. hourly project cost</span>
          <span class="summary-value text-fg-secondary"
            >${totalHourlyCost}/hr</span
          >
          <span class="summary-breakdown-plain">
            ${prodHourlyCost} prod + ${runningDevHourlyCost} dev
          </span>
          {#if cycleStart && cycleEnd}
            <span class="summary-cycle">
              Billing cycle: {formatCycleDate(cycleStart)} – {formatCycleDate(
                cycleEnd,
              )}
            </span>
          {/if}
        </div>
      {/if}
    </div>

    <div class="section-grid">
      <!-- Production card -->
      <div class="breakdown-card breakdown-card-prod">
        <div class="breakdown-header">
          <h4 class="breakdown-title">Production</h4>
          <button
            class="manage-btn"
            onclick={() => (prodSlotsModalOpen = true)}
          >
            Manage units
          </button>
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <div class="metric-row">
              <span class="metric-value">
                {prodMemory}<span class="metric-unit">GiB</span>
                <span class="metric-slash">/</span>
                {prodCpu}<span class="metric-unit">vCPU</span>
              </span>
            </div>
            <span class="metric-label">Cluster size</span>
          </div>
          <div class="breakdown-details">
            <div class="detail-row">
              <span class="detail-label">Units</span>
              <span class="detail-value">{prodSlots}</span>
            </div>
            {#if hasCostVisibility}
              <div class="detail-row">
                <span class="detail-label">Est. cost</span>
                <span class="detail-value-sm">~${prodHourlyCost}/hr</span>
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Development card -->
      <div class="breakdown-card breakdown-card-dev">
        <div class="breakdown-header">
          <h4 class="breakdown-title">Development</h4>
          <button class="manage-btn" onclick={() => (devSlotsModalOpen = true)}>
            Manage units
          </button>
        </div>
        <div class="breakdown-body">
          <div class="breakdown-metric">
            <div class="metric-row">
              <span
                class="metric-value"
                class:metric-value-empty={devSlotsPerDeployment === 0}
              >
                {#if devSlotsPerDeployment > 0}
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
              <span class="detail-label">Units (per deployment)</span>
              <span class="detail-value">{devSlotsPerDeployment}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">Active deployments</span>
              <span class="detail-value">{activeDevDeploymentCount}</span>
            </div>
            {#if hasCostVisibility}
              <div class="detail-row">
                <span class="detail-label">Est. cost</span>
                <span class="detail-value-sm">
                  {runningDevSlots > 0 ? `~$${runningDevHourlyCost}/hr` : "—"}
                </span>
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </div>

  <ManageSlotsModal
    bind:open={prodSlotsModalOpen}
    {organization}
    {project}
    currentSlots={prodSlots}
    title="Manage Prod Cluster Size"
    maxSlots={prodMaxSlots}
    showCost={hasCostVisibility}
  />

  <ManageSlotsModal
    bind:open={devSlotsModalOpen}
    {organization}
    {project}
    currentSlots={devSlotsPerDeployment}
    title="Manage Dev Cluster Size"
    minSlots={0}
    slotType="dev"
    maxSlots={devMaxSlots}
    showCost={hasCostVisibility}
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
  .summary-warning {
    @apply text-red-600 font-medium;
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
    @apply inline-flex items-center gap-1.5 text-xs font-medium text-fg-secondary bg-transparent border border-border rounded-md px-2 py-1 cursor-pointer;
  }
  .manage-btn:hover {
    @apply bg-surface-subtle text-fg-primary;
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
</style>
