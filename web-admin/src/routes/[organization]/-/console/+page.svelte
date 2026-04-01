<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListOrganizationProjectsWithHealth,
    createAdminServiceListOrganizationResources,
  } from "@rilldata/web-admin/client";
  import OverviewCard from "@rilldata/web-common/features/projects/status/overview/OverviewCard.svelte";
  import {
    prettyResourceKind,
    resourceKindStyleName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import {
    isProjectHealthy,
    hasProjectErrors,
  } from "@rilldata/web-admin/features/projects/admin-console/project-health-utils";

  $: organization = $page.params.organization;

  $: orgQuery = createAdminServiceGetOrganization(organization);
  $: org = $orgQuery.data?.organization;

  $: healthQuery = createAdminServiceListOrganizationProjectsWithHealth(
    organization,
    { pageSize: 50 },
  );
  $: projects = $healthQuery.data?.projects ?? [];

  $: resourcesQuery = createAdminServiceListOrganizationResources(organization);
  $: allResources = $resourcesQuery.data?.resources ?? [];

  $: totalProjects = projects.length;
  $: healthyCount = projects.filter(isProjectHealthy).length;
  $: errorCount = projects.filter(hasProjectErrors).length;

  // Group resources by kind
  $: resourcesByKind = allResources.reduce(
    (acc, r) => {
      const kind = r.kind ?? "unknown";
      acc[kind] = (acc[kind] ?? 0) + 1;
      return acc;
    },
    {} as Record<string, number>,
  );

  // Group errored resources by kind
  $: erroredByKind = allResources
    .filter((r) => !!r.reconcileError)
    .reduce(
      (acc, r) => {
        const kind = r.kind ?? "unknown";
        acc[kind] = (acc[kind] ?? 0) + 1;
        return acc;
      },
      {} as Record<string, number>,
    );

  $: hasAnyErrors = Object.keys(erroredByKind).length > 0;
</script>

<div class="flex flex-col gap-6">
  <OverviewCard title="Organization">
    {#if $orgQuery.isLoading}
      <p class="text-sm text-fg-secondary">Loading...</p>
    {:else if org}
      <div class="info-grid">
        <div class="info-row">
          <span class="info-label">Name</span>
          <span class="info-value">{org.name}</span>
        </div>
        {#if org.displayName}
          <div class="info-row">
            <span class="info-label">Display Name</span>
            <span class="info-value">{org.displayName}</span>
          </div>
        {/if}
        {#if org.description}
          <div class="info-row">
            <span class="info-label">Description</span>
            <span class="info-value">{org.description}</span>
          </div>
        {/if}
        {#if org.billingEmail}
          <div class="info-row">
            <span class="info-label">Billing Email</span>
            <span class="info-value">{org.billingEmail}</span>
          </div>
        {/if}
        {#if org.billingPlanDisplayName}
          <div class="info-row">
            <span class="info-label">Plan</span>
            <span class="info-value">{org.billingPlanDisplayName}</span>
          </div>
        {/if}
        {#if org.customDomain}
          <div class="info-row">
            <span class="info-label">Custom Domain</span>
            <span class="info-value">{org.customDomain}</span>
          </div>
        {/if}
        {#if org.createdOn}
          <div class="info-row">
            <span class="info-label">Created</span>
            <span class="info-value">
              {new Date(org.createdOn).toLocaleDateString("en-US", {
                year: "numeric",
                month: "long",
                day: "numeric",
              })}
            </span>
          </div>
        {/if}
      </div>
    {/if}
  </OverviewCard>

  <OverviewCard title="Projects" viewAllHref="/{organization}/-/console/projects">
    {#if $healthQuery.isLoading}
      <p class="text-sm text-fg-secondary">Loading projects...</p>
    {:else if projects.length === 0}
      <p class="text-sm text-fg-secondary">No projects found.</p>
    {:else}
      <div class="chips">
        <a href="/{organization}/-/console/projects" class="chip">
          <span class="font-medium">{totalProjects}</span>
          <span class="text-fg-secondary">{totalProjects === 1 ? "Project" : "Projects"}</span>
        </a>
        <a href="/{organization}/-/console/projects?status=healthy" class="chip chip-green">
          <span class="w-2 h-2 rounded-full bg-green-500"></span>
          <span class="font-medium">{healthyCount}</span>
          <span class="text-fg-secondary">Healthy</span>
        </a>
        <a href="/{organization}/-/console/projects?status=error" class="chip chip-red">
          <span class="w-2 h-2 rounded-full bg-red-500"></span>
          <span class="font-medium">{errorCount}</span>
          <span class="text-fg-secondary">Error</span>
        </a>
      </div>
    {/if}
  </OverviewCard>

  <OverviewCard title="Resources" viewAllHref="/{organization}/-/console/resources">
    {#if $resourcesQuery.isLoading}
      <p class="text-sm text-fg-secondary">Loading resources...</p>
    {:else if allResources.length === 0}
      <p class="text-sm text-fg-secondary">No resources found.</p>
    {:else}
      <div class="chips">
        {#each Object.entries(resourcesByKind).sort(([, a], [, b]) => b - a) as [kind, count]}
          <a
            href="/{organization}/-/console/resources?kind={encodeURIComponent(kind)}"
            class="chip {resourceKindStyleName(kind) ?? ''}"
          >
            {#if resourceIconMapping[kind]}
              <svelte:component this={resourceIconMapping[kind]} size="12px" />
            {/if}
            <span class="font-medium">{count}</span>
            <span>{prettyResourceKind(kind)}</span>
          </a>
        {/each}
      </div>
    {/if}
  </OverviewCard>

  {#if $resourcesQuery.isLoading}
    <div class="section">
      <div class="section-header">
        <h3 class="section-title">Errors</h3>
      </div>
      <p class="text-sm text-fg-secondary">Loading...</p>
    </div>
  {:else if !hasAnyErrors}
    <div class="section">
      <div class="section-header">
        <h3 class="section-title">Errors</h3>
      </div>
      <p class="text-sm text-fg-secondary">No errors detected.</p>
    </div>
  {:else}
    <a
      href="/{organization}/-/console/resources?status=error"
      class="section section-error section-clickable"
    >
      <div class="section-header">
        <h3 class="section-title flex items-center gap-2">
          Errors
          <span class="error-badge">{Object.values(erroredByKind).reduce((a, b) => a + b, 0)}</span>
        </h3>
      </div>
      <div class="error-chips">
        {#each Object.entries(erroredByKind).sort(([, a], [, b]) => b - a) as [kind, count]}
          <span class="error-chip">
            {#if resourceIconMapping[kind]}
              <svelte:component this={resourceIconMapping[kind]} size="12px" />
            {/if}
            <span class="font-medium">{count}</span>
            <span>{prettyResourceKind(kind)}</span>
          </span>
        {/each}
      </div>
    </a>
  {/if}
</div>

<style lang="postcss">
  .info-grid {
    @apply flex flex-col;
  }
  .info-row {
    @apply flex items-center py-2;
  }
  .info-label {
    @apply text-sm text-fg-secondary w-32 shrink-0;
  }
  .info-value {
    @apply text-sm text-fg-primary;
  }
  .chips {
    @apply flex flex-wrap gap-2;
  }
  .chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border border-border bg-surface-subtle no-underline text-inherit;
  }
  .chip:hover {
    @apply border-primary-500 text-primary-600;
  }
  .chip-green {
    @apply border-green-200;
  }
  .chip-red {
    @apply border-red-200;
  }
  .section {
    @apply block border border-border rounded-lg p-5 no-underline text-inherit;
  }
  .section-clickable {
    @apply cursor-pointer;
  }
  .section-error {
    @apply border-red-500;
  }
  .section-clickable:hover {
    @apply border-red-600;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .error-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .error-chips {
    @apply flex flex-wrap gap-2;
  }
  .error-chip {
    @apply flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md;
    @apply border border-red-300 bg-red-50 text-red-700 no-underline;
  }
</style>
