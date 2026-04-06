<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceGetBillingSubscription,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
  import {
    isTrialPlan,
    isTeamPlan,
    isFreePlan,
    isGrowthPlan,
    isEnterprisePlan,
  } from "@rilldata/web-admin/features/billing/plans/utils";
  import { useProjectDeployment } from "../selectors";
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
  let olapType = $derived(projectData?.olapConnector ?? "");
  let isRillManaged = $derived(olapType === "" || olapType === "duckdb");
  let slotsModalOpen = $state(false);

  // Billing
  let subscriptionQuery = $derived(
    createAdminServiceGetBillingSubscription(organization),
  );
  let planName = $derived(
    $subscriptionQuery?.data?.subscription?.plan?.name ?? "",
  );
  let isTrial = $derived(isTrialPlan(planName));
  let isTeam = $derived(isTeamPlan(planName));
  let isFree = $derived(isFreePlan(planName));
  let isGrowth = $derived(isGrowthPlan(planName));
  let isEnterprise = $derived(planName !== "" && isEnterprisePlan(planName));

  // Estimated costs
  let rillMonthlyCost = $derived(
    Math.round(currentSlots * SLOT_RATE_PER_HR * HOURS_PER_MONTH),
  );
</script>

{#if !isEnterprise}
  <div class="page">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center gap-3">
        <h2 class="page-title">Deployments</h2>
        <span class="status-badge {getStatusDotClass(deploymentStatus)}">
          {getStatusLabel(deploymentStatus)}
        </span>
      </div>
      {#if canManage && !$subscriptionQuery?.isLoading}
        <button class="manage-btn" onclick={() => (slotsModalOpen = true)}>
          Manage Slots
        </button>
      {/if}
    </div>

    <!-- Summary cards -->
    <div class="slot-cards">
      <div class="slot-card">
        <span class="slot-card-label">Rill Slots</span>
        <span class="slot-card-value">{currentSlots}</span>
        <span class="slot-card-sub">
          @ ${SLOT_RATE_PER_HR}/slot/hr (~${rillMonthlyCost.toLocaleString()}/mo)
        </span>
        <a href="/{organization}/-/settings/usage" class="pricing-link">
          See price breakdown
        </a>
      </div>
    </div>
  </div>

  <ManageSlotsModal
    bind:open={slotsModalOpen}
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
  .slot-card-label {
    @apply text-xs text-fg-secondary uppercase tracking-wide font-medium;
  }
  .slot-card-value {
    @apply text-2xl font-semibold text-fg-primary tabular-nums;
  }
  .slot-card-sub {
    @apply text-xs text-fg-tertiary;
  }
  .pricing-link {
    @apply text-xs text-primary-500 no-underline mt-1;
  }
  .pricing-link:hover {
    @apply text-primary-600 underline;
  }

  /* Detail sections */
  .detail-section {
    @apply border border-border rounded-lg p-5;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide mb-4;
  }
</style>
