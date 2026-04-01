<script lang="ts">
  import OrgResourceTable from "@rilldata/web-admin/features/projects/admin-console/OrgResourceTable.svelte";

  // TODO: Wire up with createAdminServiceListOrganizationResources once the
  // generated client is available. That endpoint returns resources across all
  // projects in the org, including reconcile status and errors.
  //
  // For now, show a placeholder. The OrgResourceTable component is ready to
  // accept data; pass it an array of { projectName, kind, name,
  // reconcileStatus, reconcileError, stateUpdatedOn }.
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium text-fg-primary">Resources</h2>
  <div
    class="border border-dashed border-border rounded-lg p-8 flex items-center justify-center min-h-[200px]"
  >
    <p class="text-fg-secondary text-sm">
      Organization-wide resource view will be available once the API client is
      generated. Run <code class="font-mono text-xs">make proto.generate</code> after
      the backend endpoint is merged.
    </p>
  </div>
</section>
