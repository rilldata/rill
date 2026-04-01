<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceListOrganizationResources } from "@rilldata/web-admin/client";
  import OrgResourceTable from "@rilldata/web-admin/features/projects/admin-console/OrgResourceTable.svelte";

  $: organization = $page.params.organization;

  $: resourcesQuery = createAdminServiceListOrganizationResources(organization);

  $: resources = ($resourcesQuery.data?.resources ?? []).map((r) => ({
    projectName: r.projectName ?? "",
    kind: r.kind ?? "",
    name: r.name ?? "",
    reconcileStatus: r.reconcileError ? "ERROR" : (r.reconcileStatus ?? ""),
    reconcileError: r.reconcileError ?? "",
    stateUpdatedOn: r.stateUpdatedOn ?? "",
  }));
</script>

{#if $resourcesQuery.isLoading}
  <p class="text-fg-secondary text-sm">Loading resources...</p>
{:else if $resourcesQuery.isError}
  <p class="text-red-500 text-sm">Failed to load resources</p>
{:else}
  <OrgResourceTable {organization} {resources} />
{/if}
