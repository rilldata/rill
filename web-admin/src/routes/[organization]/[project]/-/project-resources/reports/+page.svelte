<script lang="ts">
  import { page } from "$app/stores";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";
  import { useReports } from "@rilldata/web-admin/features/scheduled-reports/selectors";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: ({ instanceId } = $runtime);
  $: organization = $page.params.organization;
  $: project = $page.params.project;

  $: query = useReports(instanceId);
  $: reports = $query.data?.resources ?? [];
</script>

<div class="flex flex-col items-center gap-y-4 w-full">
  {#if $query.isLoading}
    <div class="m-auto mt-20">
      <DelayedSpinner isLoading size="24px" />
    </div>
  {:else if $query.isError}
    <p class="text-red-500">Error loading reports</p>
  {:else}
    <ReportsTable {organization} {project} data={reports} />
  {/if}
</div>
