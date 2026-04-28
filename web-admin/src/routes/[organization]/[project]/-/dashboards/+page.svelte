<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";

  $: ({
    params: { project },
  } = $page);
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer maxWidth={1100} title="Dashboards">
  <div class="flex flex-col gap-y-4 w-full">
    <DashboardsTable />
  </div>
</ContentContainer>
