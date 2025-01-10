<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "../../../features/dashboards/listing/DashboardsTable.svelte";

  $: project = $page.params.project;
</script>

<svelte:head>
  <title>{project} overview - Rill</title>
</svelte:head>

<ContentContainer maxWidth={800} title="Project dashboards">
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable />
  </div>
</ContentContainer>
