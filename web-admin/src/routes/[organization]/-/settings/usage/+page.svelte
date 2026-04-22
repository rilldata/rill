<script lang="ts">
  import {
    createAdminServiceGetEmbeddedAnalytics,
    createAdminServiceListProjectsForOrganization,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: organization = data.organization;

  // Projects
  $: projectsQuery =
    createAdminServiceListProjectsForOrganization(organization);
  $: projects = $projectsQuery.data?.projects ?? [];

  // Storage per project (for the by-project table)
  $: usageMetrics = getOrganizationUsageMetrics(organization);
  $: storageByProject = new Map(
    ($usageMetrics?.data ?? []).map((m) => [m.project_name, m.size]),
  );

  // Embedded analytics canvases replacing the two KPI sections.
  $: topLevelQuery = createAdminServiceGetEmbeddedAnalytics(
    organization,
    "usage_top_level",
  );
  $: middleLevelQuery = createAdminServiceGetEmbeddedAnalytics(
    organization,
    "usage_middle_level",
  );

  const RATE_PER_UNIT_HR = 0.15;

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
  <div class="embed-wrapper">
    {#if $topLevelQuery.isLoading}
      <div class="embed-placeholder">Loading…</div>
    {:else if $topLevelQuery.error}
      <div class="embed-error">
        Failed to load usage analytics: {$topLevelQuery.error.message ??
          "unknown error"}
      </div>
    {:else if $topLevelQuery.data?.iframeUrl}
      <iframe
        src={$topLevelQuery.data.iframeUrl}
        title="Usage – top level"
        class="embed-iframe embed-iframe-top"
      ></iframe>
    {/if}
  </div>

  <!-- Current usage by project -->
  <h2 class="section-title mt-10 mb-3">Current Usage by Project</h2>
  <div class="embed-wrapper">
    {#if $middleLevelQuery.isLoading}
      <div class="embed-placeholder">Loading…</div>
    {:else if $middleLevelQuery.error}
      <div class="embed-error">
        Failed to load usage analytics: {$middleLevelQuery.error.message ??
          "unknown error"}
      </div>
    {:else if $middleLevelQuery.data?.iframeUrl}
      <iframe
        src={$middleLevelQuery.data.iframeUrl}
        title="Usage – by project"
        class="embed-iframe embed-iframe-middle"
      ></iframe>
    {/if}
  </div>

  <div class="table-wrapper">
    <table class="project-table">
      <thead>
        <tr>
          <th>Project</th>
          <th>Type</th>
          <th>Prod compute units</th>
          <th>Dev compute units</th>
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

  /* Embedded canvas wrappers */
  .embed-wrapper {
    @apply border border-border rounded-xl overflow-hidden bg-white mb-4;
  }
  .embed-iframe {
    @apply w-full block border-0;
  }
  .embed-iframe-top {
    height: 180px;
  }
  .embed-iframe-middle {
    height: 240px;
  }
  .embed-placeholder {
    @apply text-sm text-fg-tertiary p-5;
  }
  .embed-error {
    @apply text-sm text-red-600 p-5;
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
