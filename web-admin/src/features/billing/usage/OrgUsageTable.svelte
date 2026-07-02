<script lang="ts">
  import { createAdminServiceListProjectsForOrganization } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

  let {
    org,
  }: {
    org: string;
  } = $props();

  let projectsQuery = $derived(
    createAdminServiceListProjectsForOrganization(org),
  );
  let projects = $derived($projectsQuery.data?.projects ?? []);

  // Storage per project (for the by-project table)
  let usageMetrics = $derived(getOrganizationUsageMetrics(org));
  let storageByProject = $derived(
    new Map(($usageMetrics?.data ?? []).map((m) => [m.project_name, m.size])),
  );

  function getProjectType(provisioner: string | undefined): string {
    if (!provisioner || provisioner === "rill") return "Rill managed";
    return provisioner.toUpperCase();
  }

  function getProjectStorageFormatted(projectName: string): string {
    const bytes = storageByProject.get(projectName) ?? 0;
    return formatMemorySize(bytes);
  }
</script>

<div class="table-wrapper">
  <table class="project-table">
    <thead>
      <tr>
        <th>Project</th>
        <th>Type</th>
        <th>Prod compute units</th>
        <th>Dev compute units</th>
        <th>Storage</th>
        <th>Action</th>
      </tr>
    </thead>
    <tbody>
      {#each projects as project (project.id)}
        {@const pProd = Number(project.prodSlots ?? 0)}
        {@const pDev = Number(project.devSlots ?? 0)}
        <tr>
          <td class="project-name">{project.name}</td>
          <td>{getProjectType(project.provisioner)}</td>
          <td>{pProd}</td>
          <td>{pDev}</td>
          <td>{getProjectStorageFormatted(project.name ?? "")}</td>
          <td>
            <a
              href="/{org}/{project.name}/-/status/branches"
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

<style lang="postcss">
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
