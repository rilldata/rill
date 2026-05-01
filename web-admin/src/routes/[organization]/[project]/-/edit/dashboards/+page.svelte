<script lang="ts">
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { page } from "$app/stores";

  $: ({
    params: { project },
  } = $page);
</script>

<svelte:head>
  <title>{project} dashboards (preview) - Rill</title>
</svelte:head>

<ContentContainer maxWidth={800} title="Project dashboards">
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
