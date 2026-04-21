<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isFreePlan,
    isProPlan,
    isManagedPlan,
    isTeamPlan,
    isEnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { SLOT_RATE_PER_HR } from "@rilldata/web-admin/features/projects/status/overview/slots-utils";
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

  // Slot types
  let prodSlots = $derived(parseInt(projectData?.prodSlots ?? "0", 10) || 0);
  let devSlots = $derived(parseInt(projectData?.devSlots ?? "0", 10) || 0);
  let totalSlots = $derived(prodSlots + devSlots);

  // Cluster info (split into number + unit for display)
  let prodMemory = $derived(prodSlots * 4);
  let prodCpu = $derived(prodSlots);
  let devMemory = $derived(devSlots * 4);
  let devCpu = $derived(devSlots);

  // Cost calculations
  let prodHourlyCost = $derived((prodSlots * SLOT_RATE_PER_HR).toFixed(2));
  let devHourlyCost = $derived((devSlots * SLOT_RATE_PER_HR).toFixed(2));
  let totalHourlyCost = $derived((totalSlots * SLOT_RATE_PER_HR).toFixed(2));
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
          {devSlots} development
        </span>
      </div>

      <div class="summary-divider"></div>

      <div class="summary-panel">
        <span class="summary-label">Est. hourly project cost</span>
        <span class="summary-value text-fg-secondary"
          >${totalHourlyCost}/hr</span
        >
        <span class="summary-breakdown-plain">
          ${prodHourlyCost} prod + ${devHourlyCost} dev
        </span>
        {#if cycleStart && cycleEnd}
          <span class="summary-cycle">
            Billing cycle: {formatCycleDate(cycleStart)} – {formatCycleDate(
              cycleEnd,
            )}
          </span>
        {/if}
      </div>
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
              <span class="detail-value-sm">~${prodHourlyCost}/hr</span>
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
                {devSlots > 0 ? `~$${devHourlyCost}/hr` : "—"}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
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
