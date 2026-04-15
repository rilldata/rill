<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListProjectsForOrganization,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: organization = data.organization;

  // Billing subscription for cycle dates
  $: subscriptionQuery = createAdminServiceGetBillingSubscription(organization);
  $: subscription = $subscriptionQuery?.data?.subscription;
  $: cycleStart = subscription?.currentBillingCycleStartDate;
  $: cycleEnd = subscription?.currentBillingCycleEndDate;

  // Projects
  $: projectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  $: projects = $projectsQuery.data?.projects ?? [];

  // Aggregate slots
  $: totalProdSlots = projects.reduce(
    (sum, p) => sum + Number(p.prodSlots ?? 0),
    0,
  );
  $: totalDevSlots = projects.reduce(
    (sum, p) => sum + Number(p.devSlots ?? 0),
    0,
  );
  $: totalSlots = totalProdSlots + totalDevSlots;

  // Storage per project
  $: usageMetrics = getOrganizationUsageMetrics(organization);
  $: storageByProject = new Map(
    ($usageMetrics?.data ?? []).map((m) => [m.project_name, m.size]),
  );
  $: totalStorageBytes = projects.reduce(
    (sum, p) => sum + (storageByProject.get(p.name ?? "") ?? 0),
    0,
  );

  // Pricing constants
  const RATE_PER_UNIT_HR = 0.15;
  const FREE_STORAGE_GB = 1;

  // Hours elapsed in current billing cycle (placeholder until Orb API)
  $: hoursElapsed = (() => {
    if (!cycleStart) return 0;
    const start = new Date(cycleStart).getTime();
    const now = Date.now();
    return Math.max(0, Math.floor((now - start) / (1000 * 60 * 60)));
  })();

  // Current period costs
  $: prodCost = totalProdSlots * hoursElapsed * RATE_PER_UNIT_HR;
  $: devCost = totalDevSlots * hoursElapsed * RATE_PER_UNIT_HR;
  $: billableStorageGB = Math.max(totalStorageBytes / 1e9 - FREE_STORAGE_GB, 0);
  $: storageCost = billableStorageGB * 1; // $1/GB/mo
  $: totalCost = prodCost + devCost + storageCost;

  // Daily costs
  $: prodDailyRate = totalProdSlots * 24 * RATE_PER_UNIT_HR;
  $: devDailyRate = totalDevSlots * 24 * RATE_PER_UNIT_HR;
  $: totalDailyRate = totalSlots * 24 * RATE_PER_UNIT_HR;

  function fmtUSD(n: number): string {
    return n.toLocaleString(undefined, {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    });
  }

  function formatCyclePeriod(): string {
    const now = new Date();
    const start = new Date(now.getFullYear(), now.getMonth(), 1);
    const end = new Date(now.getFullYear(), now.getMonth() + 1, 0);
    const fmt = (d: Date) =>
      d.toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    return `${fmt(start)} – ${fmt(end)}`;
  }

  function getProjectType(provisioner: string | undefined): string {
    if (!provisioner || provisioner === "rill") return "Rill managed";
    return provisioner.toUpperCase();
  }

  function getProjectStorageFormatted(projectName: string): string {
    const bytes = storageByProject.get(projectName) ?? 0;
    return formatMemorySize(bytes);
  }

  function getProjectEstCost(prodSlots: number, devSlots: number): string {
    const total = prodSlots + devSlots;
    if (total === 0) return "$--";
    return fmtUSD(total * RATE_PER_UNIT_HR * 24 * 30);
  }
</script>

<div class="usage-page">
  <h1 class="page-title">Usage</h1>

  <!-- Current period cost -->
  <div class="section-header">
    <h2 class="section-title">Current period cost</h2>
    <span class="cycle-dates">
      {formatCyclePeriod()}
    </span>
  </div>

  <div class="summary-bar">
    <div class="summary-cell">
      <span class="summary-label">Total estimated cost</span>
      <span class="summary-value">{fmtUSD(totalCost)}</span>
      <a href="/{organization}/-/settings/billing" class="summary-link">
        View my plan
        <svg
          class="w-3 h-3"
          viewBox="0 0 12 12"
          fill="none"
          stroke="currentColor"
          stroke-width="1.5"
        >
          <path d="M1 6h9M7.5 3l3 3-3 3" />
        </svg>
      </a>
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Production</span>
      <span class="summary-value">{fmtUSD(prodCost)}</span>
      <span class="summary-desc"
        >{hoursElapsed} hrs · {totalProdSlots} units × ${RATE_PER_UNIT_HR.toFixed(
          2,
        )}</span
      >
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Development</span>
      <span class="summary-value">{fmtUSD(devCost)}</span>
      <span class="summary-desc"
        >{hoursElapsed} hrs · {totalDevSlots} units × ${RATE_PER_UNIT_HR.toFixed(
          2,
        )}</span
      >
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Storage</span>
      <span class="summary-value">{fmtUSD(storageCost)}</span>
      <span class="summary-desc"
        >{(totalStorageBytes / 1e9).toFixed(1)} GB · {FREE_STORAGE_GB} GB free ·
        $1/GB/mo</span
      >
    </div>
  </div>

  <!-- By project -->
  <h2 class="section-title mt-10 mb-3">By project</h2>

  <div class="summary-bar">
    <div class="summary-cell">
      <span class="summary-label">Total projects</span>
      <span class="summary-value-lg">{projects.length}</span>
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label inline-flex items-center gap-1">
        Total unit
        <Tooltip location="right" alignment="middle" distance={8}>
          <span class="text-fg-muted flex">
            <InfoCircle size="13px" />
          </span>
          <TooltipContent maxWidth="200px" slot="tooltip-content">
            1 unit = 4 GiB RAM / 1 vCPU
          </TooltipContent>
        </Tooltip>
      </span>
      <span class="summary-value-lg">{totalSlots}</span>
      <span class="summary-desc">{fmtUSD(totalDailyRate)}/day</span>
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Prod unit</span>
      <span class="summary-value-lg">{totalProdSlots}</span>
      <span class="summary-desc">{fmtUSD(prodDailyRate)}/day</span>
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Dev unit</span>
      <span class="summary-value-lg">{totalDevSlots}</span>
      <span class="summary-desc">{fmtUSD(devDailyRate)}/day</span>
    </div>
    <div class="summary-divider"></div>
    <div class="summary-cell">
      <span class="summary-label">Storage</span>
      <span class="summary-value-lg">{formatMemorySize(totalStorageBytes)}</span
      >
      <span class="summary-desc">{FREE_STORAGE_GB} GB free · $1/GB/mo</span>
    </div>
  </div>

  <!-- Project table -->
  <div class="table-wrapper">
    <table class="project-table">
      <thead>
        <tr>
          <th>Project</th>
          <th>Type</th>
          <th>Prod slots</th>
          <th>Dev slots</th>
          <th>Storage</th>
          <th>Est. cost</th>
          <th>Action</th>
        </tr>
      </thead>
      <tbody>
        {#each projects as project}
          {@const pProd = Number(project.prodSlots ?? 0)}
          {@const pDev = Number(project.devSlots ?? 0)}
          <tr>
            <td class="project-name">{project.name}</td>
            <td>{getProjectType(project.provisioner)}</td>
            <td>{pProd}</td>
            <td>{pDev}</td>
            <td>{getProjectStorageFormatted(project.name ?? "")}</td>
            <td>{getProjectEstCost(pProd, pDev)}</td>
            <td>
              <a
                href="/{organization}/{project.name}/-/status/branches"
                class="manage-btn"
              >
                Manage
              </a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<style lang="postcss">
  .usage-page {
    @apply flex flex-col w-full max-w-5xl;
  }

  .page-title {
    @apply text-2xl font-semibold text-fg-primary mb-6;
  }

  .section-header {
    @apply flex items-center justify-between mb-3;
  }

  .section-title {
    @apply font-sans font-medium text-lg leading-7 text-fg-primary;
  }

  .cycle-dates {
    @apply text-sm text-fg-tertiary;
  }

  /* Summary bar */
  .summary-bar {
    @apply flex border border-border rounded-xl overflow-hidden bg-white;
  }

  .summary-cell {
    @apply flex-1 flex flex-col gap-1 p-5;
  }

  .summary-divider {
    @apply w-px bg-border my-4;
  }

  .summary-label {
    @apply text-xs font-semibold text-fg-tertiary;
  }

  .summary-value {
    @apply font-sans font-medium text-2xl leading-8 tabular-nums text-fg-secondary;
  }

  .summary-value-lg {
    @apply font-sans font-medium text-2xl leading-8 tabular-nums text-fg-secondary;
  }

  .summary-desc {
    @apply text-xs text-fg-tertiary;
  }

  .summary-link {
    @apply flex items-center gap-1 text-xs font-medium text-primary-500 no-underline mt-0.5;
  }
  .summary-link:hover {
    @apply text-primary-600 underline;
  }

  /* Project table */
  .table-wrapper {
    @apply mt-4;
  }

  .project-table {
    @apply w-full text-sm;
    border-collapse: collapse;
  }

  .project-table th {
    @apply text-left text-xs font-medium text-fg-tertiary py-3 px-4 border-b border-border;
  }

  .project-table td {
    @apply py-3 px-4 text-sm text-fg-primary border-b border-border;
  }

  .project-table tr:last-child td {
    @apply border-b-0;
  }

  .project-name {
    @apply font-medium;
  }

  .manage-btn {
    @apply text-xs font-medium text-fg-primary bg-transparent border border-border rounded-sm no-underline inline-flex items-center justify-center;
    padding: 8px 12px;
    gap: 8px;
  }
  .manage-btn:hover {
    @apply bg-surface-subtle;
  }
</style>
